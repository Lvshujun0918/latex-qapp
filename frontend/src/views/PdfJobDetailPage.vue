<template>
  <section class="app-page app-inner-page page-wrap pt-8" :class="{ 'is-dark': resolvedTheme === 'dark' }">
    <header class="app-page-header page-header">
      <Button variant="outline" size="icon-sm" class="app-header-back" @click="goBack" aria-label="返回上一级"><</Button>
      <span class="app-kicker">Task Status</span>
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
          <a
            v-if="pdfUrl"
            class="download-link"
            :href="pdfUrl"
            target="_blank"
            rel="noopener noreferrer"
          >
            下载 PDF 文件
          </a>
          <Button v-else variant="outline" :disabled="loading" @click="fetchJob">刷新状态</Button>
        </div>

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
import { useRoute } from 'vue-router';
import { useRouter } from 'vue-router';
import { useTheme } from '@/composables/useTheme';
import { getPdfJob } from '@/services/pdf';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

const route = useRoute();
const router = useRouter();
const { resolvedTheme } = useTheme();
const loading = ref(false);
const errorMessage = ref('');
const jobStatus = ref('queued');
const jobMessage = ref('');
const pdfPath = ref('');
const selectedCount = ref(0);
const questionList = ref<Array<{ id: number; index: number; title: string; subject: string; question_type: string }>>([]);
let pollTimer: number | null = null;

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
          index: Number(item?.index ?? idx + 1),
          title: String(item?.title ?? '').trim() || `第 ${idx + 1} 题`,
          subject: String(item?.subject ?? '').trim(),
          question_type: String(item?.question_type ?? '').trim(),
        }))
      : [];

    if (jobStatus.value === 'done' || pdfPath.value) {
      clearPoll();
    }
  } catch (error: any) {
    errorMessage.value = error?.response?.data?.error || error?.message || '获取任务状态失败';
  } finally {
    loading.value = false;
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

  router.replace('/pdf/export');
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
