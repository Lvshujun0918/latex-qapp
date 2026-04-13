<template>
  <section class="app-page app-inner-page page-wrap pt-8" :class="{ 'is-dark': resolvedTheme === 'dark' }">
    <header class="app-page-header page-header">
      <Button variant="outline" size="icon-sm" class="app-header-back mr-4" @click="goBack" aria-label="返回上一级"><</Button>
      <h1>PDF 结果</h1>
      <p>任务完成，可直接下载文件。</p>
    </header>

    <Card class="app-page-shell hero-shell">
      <CardHeader>
        <div class="hero-top">
          <CardTitle>任务 {{ route.params.jobId }}</CardTitle>
          <span class="status-pill" :class="`status-${jobStatus}`">{{ statusLabel }}</span>
        </div>
      </CardHeader>
      <CardContent class="hero-content">
        <div v-if="jobStatus === 'done' || pdfUrl" class="success-wrap">
          <div class="checkmark-circle" aria-hidden="true">
            <svg viewBox="0 0 52 52" class="checkmark-svg">
              <circle class="checkmark-ring" cx="26" cy="26" r="24" fill="none" />
              <path class="checkmark-path" fill="none" d="M14 27l8 8 16-16" />
            </svg>
          </div>
          <p class="success-title">PDF 已生成完成</p>
          <p class="success-subtitle">你现在可以直接下载并查看文件。</p>
        </div>

        <div class="action-row">
          <template v-if="pdfUrl">
            <Button
              v-if="isNativePlatform"
              variant="outline"
              :disabled="openingNative"
              @click="openPdfNative"
            >
              {{ openingNative ? '打开中...' : '打开 PDF' }}
            </Button>

            <a
              v-if="!isNativePlatform"
              class="download-link"
              :href="pdfUrl"
              target="_blank"
              rel="noopener noreferrer"
            >
              下载 PDF 文件
            </a>
          </template>

          <Button v-else variant="outline" :disabled="loading" @click="fetchJob">刷新状态</Button>
        </div>

        <p v-if="saveMessage" class="save-message">{{ saveMessage }}</p>

        <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
      </CardContent>
    </Card>
    <Card class="app-page-shell tips-shell">
      <CardHeader>
        <CardTitle>生成信息</CardTitle>
      </CardHeader>
      <CardContent class="tips-content">
        <div v-if="questionList.length" class="question-list-wrap">
          <h4 class="question-list-title">本次生成题目</h4>
          <ul class="question-list">
            <li v-for="question in questionList" :key="`${question.id}-${question.index}`" class="question-item">
              <div class="question-main">
                <span class="question-index">{{ question.index }}.</span>
                <strong class="question-title">{{ question.title }}</strong>
              </div>
              <p class="question-meta">{{ question.subject || '未分类' }} · {{ question.question_type || '未标注题型' }}</p>
              <div class="question-block">
                <p class="question-block-title">题目</p>
                <LatexView :source="toDisplayLatex(question.latex_source)" class="latex-embed" />
              </div>
              <div class="question-block">
                <p class="question-block-title">参考答案</p>
                <LatexView :source="question.latex_answer || '暂无答案'" class="latex-embed" />
              </div>
              <div class="review-action-row">
                <span class="review-tag" :class="`is-${question.child_result || 'none'}`">{{ childResultText(question.child_result) }}</span>
                <Button size="sm" :disabled="savingReview" @click="markChildResult(question.record_id || question.id, true)">孩子做对</Button>
                <Button variant="destructive" size="sm" :disabled="savingReview" @click="markChildResult(question.record_id || question.id, false)">孩子做错</Button>
              </div>
            </li>
          </ul>
        </div>

        <p v-else class="question-empty">题目详情暂不可见，请稍后刷新。</p>
      </CardContent>
    </Card>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { Capacitor } from '@capacitor/core';
import { GlobalWorkerOptions } from 'pdfjs-dist';
import PdfWorker from 'pdfjs-dist/build/pdf.worker.min.mjs?url';
import { useRoute } from 'vue-router';
import { useRouter } from 'vue-router';
import { useTheme } from '@/composables/useTheme';
import { upsertPdfExportHistory } from '@/services/pdf-history';
import { getPdfJob, updatePdfQuestionReview } from '@/services/pdf';
import { openPdfFromLocalUri, saveRemotePdfToDevice } from '@/services/pdf-native';
import { useAuthStore } from '@/stores/auth';
import LatexView from '@/components/LatexView.vue';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

GlobalWorkerOptions.workerSrc = PdfWorker;

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();
const { resolvedTheme } = useTheme();
const loading = ref(false);
const saving = ref(false);
const openingNative = ref(false);
const savingReview = ref(false);
const errorMessage = ref('');
const saveMessage = ref('');
const jobStatus = ref('queued');
const jobMessage = ref('');
const pdfPath = ref('');
const selectedCount = ref(0);
const savedPdfUri = ref('');
const questionList = ref<Array<{
  id: number;
  record_id: number;
  index: number;
  title: string;
  subject: string;
  question_type: string;
  latex_source: string;
  latex_answer: string;
  child_result: string;
}>>([]);
let pollTimer: number | null = null;
const isNativePlatform = Capacitor.isNativePlatform();

const pdfUrl = computed(() => {
  if (!pdfPath.value) {
    return '';
  }
  if (/^https?:\/\//i.test(pdfPath.value)) {
    return pdfPath.value;
  }
  const base = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';
  return `${base}${pdfPath.value}`;
});

const statusLabel = computed(() => {
  const s = jobStatus.value;
  if (s === 'done') return '已完成';
  if (s === 'queued') return '排队中';
  if (s === 'running' || s === 'processing') return '处理中';
  if (s === 'failed' || s === 'error') return '失败';
  return s || '未知';
});

const createdAtLabel = computed(() => {
  const raw = String(route.params.jobId || '');
  const m = raw.match(/job-(\d{13})/);
  if (!m) {
    return '未知';
  }
  const ts = Number(m[1]);
  if (!Number.isFinite(ts)) {
    return '未知';
  }
  return new Date(ts).toLocaleString();
});

onMounted(async () => {
  await fetchJob();
  schedulePoll();
});

onBeforeUnmount(() => {
  clearPoll();
});

async function fetchJob() {
  loading.value = true;
  try {
    errorMessage.value = '';
    saveMessage.value = '';
    const jobId = String(route.params.jobId || '');
    const res = await getPdfJob(jobId);
    const payload = res?.data ?? res ?? {};

    jobStatus.value = String(payload.status ?? 'queued');
    jobMessage.value = String(payload.message ?? '');
    pdfPath.value = String(payload.pdf_file_url ?? '');
    selectedCount.value = Number(payload.selected_count ?? selectedCount.value ?? 0);
    questionList.value = Array.isArray(payload.questions)
      ? payload.questions.map((item: any, idx: number) => ({
          id: Number(item?.id ?? 0),
          record_id: Number(item?.record_id ?? item?.id ?? 0),
          index: Number(item?.index ?? idx + 1),
          title: String(item?.title ?? '').trim() || `第 ${idx + 1} 题`,
          subject: String(item?.subject ?? '').trim(),
          question_type: String(item?.question_type ?? '').trim(),
          latex_source: String(item?.latex_source ?? '').trim(),
          latex_answer: String(item?.latex_answer ?? '').trim(),
          child_result: String(item?.child_result ?? 'none').trim() || 'none',
        }))
      : [];

    if (jobStatus.value === 'done' || pdfPath.value) {
      upsertPdfExportHistory(
        {
          userId: authStore.userId,
          username: authStore.username,
        },
        {
          jobId,
          pdfFileUrl: pdfPath.value,
          selectedCount: selectedCount.value,
          source: 'unknown',
        },
      );
      clearPoll();
    }
  } catch (error: any) {
    errorMessage.value = error?.response?.data?.error || error?.message || '获取任务状态失败';
  } finally {
    loading.value = false;
  }
}

async function savePdfNative() {
  if (!pdfUrl.value) {
    return;
  }

  saving.value = true;
  try {
    errorMessage.value = '';
    const jobId = String(route.params.jobId || 'pdf-job');
    const result = await saveRemotePdfToDevice(pdfUrl.value, `latex-qapp-${jobId}`);
    savedPdfUri.value = result.uri;
    saveMessage.value = `PDF 已保存：${result.fileName}`;
  } catch (error: any) {
    errorMessage.value = error?.message || '保存 PDF 失败';
  } finally {
    saving.value = false;
  }
}

async function openPdfNative() {
  if (!pdfUrl.value) {
    return;
  }

  openingNative.value = true;
  try {
    errorMessage.value = '';
    if (!savedPdfUri.value) {
      const jobId = String(route.params.jobId || 'pdf-job');
      const result = await saveRemotePdfToDevice(pdfUrl.value, `latex-qapp-${jobId}`);
      savedPdfUri.value = result.uri;
      saveMessage.value = `PDF 已保存：${result.fileName}`;
    }
    await openPdfFromLocalUri(savedPdfUri.value);
  } catch (error: any) {
    errorMessage.value = error?.message || '打开 PDF 失败';
  } finally {
    openingNative.value = false;
  }
}

function schedulePoll() {
  clearPoll();
  pollTimer = window.setInterval(() => {
    if (jobStatus.value !== 'done' && !pdfPath.value) {
      fetchJob();
    }
  }, 2500);
}

function clearPoll() {
  if (pollTimer !== null) {
    window.clearInterval(pollTimer);
    pollTimer = null;
  }
}

function goBack() {
  if (window.history.length > 1) {
    router.back();
    return;
  }

  router.replace('/tabs/pdfs');
}

async function markChildResult(recordId: number, isCorrect: boolean) {
  if (!recordId || savingReview.value) {
    return;
  }

  savingReview.value = true;
  try {
    errorMessage.value = '';
    const jobId = String(route.params.jobId || '');
    const res = await updatePdfQuestionReview(jobId, recordId, isCorrect);
    const payload = res?.data ?? res ?? {};
    questionList.value = Array.isArray(payload.questions)
      ? payload.questions.map((item: any, idx: number) => ({
          id: Number(item?.id ?? 0),
          record_id: Number(item?.record_id ?? item?.id ?? 0),
          index: Number(item?.index ?? idx + 1),
          title: String(item?.title ?? '').trim() || `第 ${idx + 1} 题`,
          subject: String(item?.subject ?? '').trim(),
          question_type: String(item?.question_type ?? '').trim(),
          latex_source: String(item?.latex_source ?? '').trim(),
          latex_answer: String(item?.latex_answer ?? '').trim(),
          child_result: String(item?.child_result ?? 'none').trim() || 'none',
        }))
      : questionList.value;
    saveMessage.value = isCorrect ? '已登记：孩子做对' : '已登记：孩子做错';
  } catch (error: any) {
    errorMessage.value = error?.message || '记录结果失败';
  } finally {
    savingReview.value = false;
  }
}

function toDisplayLatex(raw: string) {
  const text = String(raw || '').trim();
  if (!text) return '暂无题目';
  try {
    const data = JSON.parse(text);
    if (!data || typeof data !== 'object') {
      return text;
    }
    const stem = String((data as any).stem || '').trim();
    const options = Array.isArray((data as any).options) ? (data as any).options : [];
    const subQuestions = Array.isArray((data as any).sub_questions) ? (data as any).sub_questions : [];
    const lines = [stem || '题目'];
    options.forEach((item: string, idx: number) => lines.push(`${String.fromCharCode(65 + idx)}. ${item}`));
    subQuestions.forEach((item: string, idx: number) => lines.push(`(${idx + 1}) ${item}`));
    return lines.join('\n');
  } catch {
    return text;
  }
}

function childResultText(result?: string) {
  if (result === 'correct') return '孩子已做对';
  if (result === 'wrong') return '孩子已做错';
  return '待批改';
}
</script>

<style scoped>
.page-wrap {
  gap: 14px;
}

.hero-shell {
  overflow: hidden;
}

.hero-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.status-pill {
  border-radius: 999px;
  padding: 4px 10px;
  font-size: 12px;
  font-weight: 600;
}

.status-queued {
  background: rgba(234, 179, 8, 0.14);
  color: #92400e;
}

.status-running,
.status-processing {
  background: rgba(59, 130, 246, 0.14);
  color: #1d4ed8;
}

.status-done {
  background: rgba(16, 185, 129, 0.14);
  color: #047857;
}

.status-failed,
.status-error {
  background: rgba(239, 68, 68, 0.14);
  color: #b91c1c;
}

.hero-content {
  display: grid;
  gap: 14px;
}

.success-wrap {
  display: grid;
  gap: 8px;
  justify-items: center;
  text-align: center;
}

.checkmark-circle {
  width: 72px;
  height: 72px;
}

.checkmark-svg {
  width: 100%;
  height: 100%;
}

.checkmark-ring {
  stroke: #34d399;
  stroke-width: 3;
  stroke-dasharray: 151;
  stroke-dashoffset: 151;
  animation: ring-draw 0.55s ease forwards;
}

.checkmark-path {
  stroke: #10b981;
  stroke-width: 4;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-dasharray: 36;
  stroke-dashoffset: 36;
  animation: check-draw 0.45s ease 0.35s forwards;
}

.success-title {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  color: #065f46;
}

.success-subtitle {
  margin: 0;
  font-size: 13px;
  color: #475569;
}

.action-row {
  display: flex;
  align-items: center;
  gap: 10px;
  justify-content: center;
  flex-wrap: wrap;
}

.download-link {
  color: #f8fafc;
  text-decoration: none;
  background: linear-gradient(110deg, #2563eb 0%, #1d4ed8 55%, #1e40af 100%);
  border-radius: 12px;
  padding: 10px 16px;
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0.01em;
  box-shadow: 0 8px 18px rgba(37, 99, 235, 0.28);
  transition: transform 0.18s ease, box-shadow 0.18s ease, opacity 0.18s ease;
  width: fit-content;
}

.download-link:hover {
  transform: translateY(-1px);
  box-shadow: 0 10px 22px rgba(29, 78, 216, 0.34);
}

.download-link:active {
  transform: translateY(0);
  opacity: 0.92;
}

.error {
  margin: 0;
  color: #dc2626;
  font-size: 13px;
}

.save-message {
  margin: 0;
  font-size: 13px;
  color: #047857;
  text-align: center;
}

.preview-shell {
  overflow: hidden;
}

.preview-content {
  padding: 0;
  border-radius: 14px;
  overflow: hidden;
  border: 1px solid rgba(148, 163, 184, 0.28);
  background: rgba(255, 255, 255, 0.9);
}

.pdf-preview {
  width: 100%;
}

.tips-content {
  display: block;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.info-item {
  border: 1px solid rgba(148, 163, 184, 0.25);
  background: rgba(248, 250, 252, 0.75);
  border-radius: 12px;
  padding: 10px 12px;
  display: grid;
  gap: 4px;
}

.info-item span {
  margin: 0;
  font-size: 12px;
  color: #64748b;
}

.info-item strong {
  margin: 0;
  font-size: 13px;
  color: #0f172a;
  word-break: break-all;
}

.info-wide {
  grid-column: 1 / -1;
}

.is-dark .info-item span {
  color: #cbd5e1;
}

.is-dark .info-item strong {
  color: #f1f5f9;
}

.is-dark .success-title {
  color: #6ee7b7;
}

.is-dark .success-subtitle {
  color: #cbd5e1;
}

.is-dark .save-message {
  color: #6ee7b7;
}

.is-dark .preview-content {
  border-color: rgba(148, 163, 184, 0.24);
  background: rgba(15, 23, 42, 0.7);
}

.is-dark .info-item {
  background: rgba(15, 23, 42, 0.55);
  border-color: rgba(148, 163, 184, 0.2);
}

.question-list-wrap {
  border-top: 1px dashed rgba(148, 163, 184, 0.35);
  padding-top: 10px;
}

.question-list-title {
  margin: 0 0 8px;
  font-size: 13px;
  color: #475569;
}

.question-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 8px;
}

.question-item {
  border: 1px solid rgba(148, 163, 184, 0.25);
  background: rgba(248, 250, 252, 0.75);
  border-radius: 12px;
  padding: 9px 11px;
  display: grid;
  gap: 8px;
}

.question-main {
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.question-index {
  font-size: 12px;
  color: #64748b;
}

.question-title {
  font-size: 13px;
  color: #0f172a;
  word-break: break-word;
}

.question-meta {
  margin: 4px 0 0;
  font-size: 12px;
  color: #64748b;
}

.question-block {
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 10px;
  padding: 8px;
  background: rgba(248, 250, 252, 0.68);
}

.question-block-title {
  margin: 0 0 6px;
  font-size: 12px;
  color: #64748b;
}

.review-action-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.review-tag {
  border-radius: 999px;
  font-size: 11px;
  padding: 4px 8px;
  border: 1px solid rgba(148, 163, 184, 0.26);
  color: #475569;
  background: rgba(148, 163, 184, 0.12);
}

.review-tag.is-correct {
  color: #065f46;
  background: rgba(16, 185, 129, 0.14);
  border-color: rgba(16, 185, 129, 0.3);
}

.review-tag.is-wrong {
  color: #991b1b;
  background: rgba(248, 113, 113, 0.14);
  border-color: rgba(248, 113, 113, 0.3);
}

.question-empty {
  margin: 0;
  font-size: 12px;
  color: #64748b;
}

.is-dark .question-list-title,
.is-dark .question-meta,
.is-dark .question-empty,
.is-dark .question-index {
  color: #cbd5e1;
}

.is-dark .question-title {
  color: #f1f5f9;
}

.is-dark .question-item {
  background: rgba(15, 23, 42, 0.55);
  border-color: rgba(148, 163, 184, 0.2);
}

.is-dark .question-block {
  border-color: rgba(148, 163, 184, 0.3);
  background: rgba(30, 41, 59, 0.65);
}

.is-dark .question-block-title {
  color: #94a3b8;
}

@keyframes ring-draw {
  to {
    stroke-dashoffset: 0;
  }
}

@keyframes check-draw {
  to {
    stroke-dashoffset: 0;
  }
}

@media (max-width: 640px) {
  .hero-top {
    align-items: flex-start;
    flex-direction: column;
  }

  .info-grid {
    grid-template-columns: 1fr;
  }
}
</style>
