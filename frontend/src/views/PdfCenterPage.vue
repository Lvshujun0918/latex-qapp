<template>
  <section class="app-page page-wrap" :class="{ 'is-dark': resolvedTheme === 'dark' }">
    <header class="app-page-header page-header">
      <h1>PDF 导出</h1>
      <p>统一管理导出任务、下载打印与批改登记。</p>
    </header>

    <Card class="app-page-shell hero-card">
      <CardContent class="hero-content">
        <div>
          <p class="hero-label">历史任务</p>
          <p class="hero-value">{{ jobs.length }}</p>
          <p class="hero-sub">可按任务查看题目与答案，并记录孩子作答结果。</p>
        </div>
        <Button @click="openCreateDialog">新建导出</Button>
      </CardContent>
    </Card>

    <Alert v-if="errorMessage" variant="destructive">
      <AlertTitle>操作失败</AlertTitle>
      <AlertDescription>{{ errorMessage }}</AlertDescription>
    </Alert>

    <div class="toolbar">
      <Button variant="outline" size="sm" :disabled="loading" @click="loadJobs">刷新列表</Button>
    </div>

    <div v-if="jobs.length" class="job-list">
      <button
        v-for="job in jobs"
        :key="job.jobId"
        type="button"
        class="job-item app-interactive-surface"
        @click="goJobDetail(job.jobId)"
      >
        <div class="job-main">
          <p class="job-title">任务 {{ job.jobId }}</p>
          <p class="job-meta">{{ formatTime(job.jobId) }} · {{ job.selected_count }} 题 · {{ statusText(job.status) }}</p>
        </div>
        <ChevronRight class="h-4 w-4 item-arrow" />
      </button>
    </div>

    <Card v-else class="empty-card">
      <CardContent class="empty-content">
        <h3>暂无导出任务</h3>
        <p>点击“新建导出”，选择题目后生成 PDF 任务。</p>
        <Button @click="openCreateDialog">立即新建</Button>
      </CardContent>
    </Card>

    <Dialog :open="createDialogOpen" @update:open="(v) => (createDialogOpen = v)">
      <DialogContent class="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>新建 PDF 导出</DialogTitle>
          <DialogDescription>选择要导出的题目，生成给孩子使用的题单。</DialogDescription>
        </DialogHeader>

        <div class="picker-toolbar">
          <Input v-model="keyword" placeholder="按题目或学科筛选" />
          <Button variant="outline" size="sm" @click="toggleSelectAll">{{ allSelected ? '取消全选' : '全选' }}</Button>
        </div>

        <div class="picker-list">
          <label v-for="record in filteredRecords" :key="record.id" class="picker-item">
            <input type="checkbox" :checked="selectedMap.has(record.id)" @change="toggleRecord(record.id)" />
            <div class="picker-main">
              <div class="title-row">
                <p class="review-title">{{ record.title || '未命名题目' }}</p>
                <span class="result-badge" :class="resultClass(record.lastReviewResult)">{{ resultText(record.lastReviewResult) }}</span>
              </div>
              <p class="review-meta">
                {{ formatSubject(record.subject) }} · 第 {{ record.reviewCount + 1 }} 次 · 目标 {{ record.nextInterval }} 天 · {{
                  formatQuestionType(record.questionType)
                }}
              </p>
              <div class="progress-row">
                <div class="progress-track">
                  <div class="progress-fill" :class="{ due: record.overdueDays >= 0 }" :style="{ width: `${record.progressPercent}%` }" />
                </div>
                <span class="progress-text">{{ record.progressPercent }}%</span>
              </div>
              <p class="urgency-text">
                {{ record.overdueDays >= 0 ? `到期 ${record.overdueDays} 天` : `距到期 ${Math.abs(record.overdueDays)} 天` }}
              </p>
            </div>
          </label>
        </div>

        <DialogFooter>
          <Button variant="outline" @click="createDialogOpen = false">取消</Button>
          <Button :disabled="!selectedIds.length || exporting" @click="createExport">
            {{ exporting ? '生成中...' : `导出 ${selectedIds.length} 题` }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { storeToRefs } from 'pinia';
import { useRoute, useRouter } from 'vue-router';
import { ChevronRight } from 'lucide-vue-next';
import { useTheme } from '@/composables/useTheme';
import { exportPdfByRecordIds, listPdfJobs } from '@/services/pdf';
import { useRecordStore } from '@/stores/record';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';

interface PdfJobRow {
  jobId: string;
  status: string;
  selected_count: number;
}

const router = useRouter();
const route = useRoute();
const recordStore = useRecordStore();
const { records } = storeToRefs(recordStore);
const { resolvedTheme } = useTheme();
const jobs = ref<PdfJobRow[]>([]);
const loading = ref(false);
const exporting = ref(false);
const createDialogOpen = ref(false);
const errorMessage = ref('');
const keyword = ref('');
const selectedMap = ref(new Set<number>());
const prefillHandled = ref(false);
const EBBINGHAUS_INTERVALS = [1, 2, 4, 7, 15, 30];

const scheduleRows = computed(() => {
  const now = Date.now();
  return records.value
    .map((record) => {
      const reviewCount = Math.max(0, Number(record.reviewCount || 0));
      const intervalIndex = Math.min(reviewCount, EBBINGHAUS_INTERVALS.length - 1);
      const baseInterval = EBBINGHAUS_INTERVALS[intervalIndex];
      const easeFactor = Math.max(1.3, Math.min(3.0, Number(record.reviewEaseFactor || 2.5)));
      const nextInterval = Math.max(1, Math.round(baseInterval * easeFactor * 0.55));

      const createdTime = Date.parse(record.createdAt || '');
      const lastReviewedTime = Date.parse(record.lastReviewedAt || '');
      const nextReviewTime = Date.parse(record.nextReviewAt || '');
      const base = Number.isFinite(lastReviewedTime)
        ? lastReviewedTime
        : Number.isFinite(createdTime)
          ? createdTime
          : now;

      const elapsedDays = Math.max(0, Math.floor((now - base) / 86400000));
      const overdueDays = Number.isFinite(nextReviewTime)
        ? Math.floor((now - nextReviewTime) / 86400000)
        : elapsedDays - nextInterval;
      const progressPercent = Number.isFinite(nextReviewTime)
        ? Math.max(0, Math.min(100, Math.round((1 - Math.max(0, nextReviewTime - now) / (nextInterval * 86400000)) * 100)))
        : Math.max(0, Math.min(100, Math.round((elapsedDays / nextInterval) * 100)));

      return {
        ...record,
        nextInterval,
        overdueDays,
        progressPercent,
      };
    })
    .sort((a, b) => b.overdueDays - a.overdueDays);
});

const filteredRecords = computed(() => {
  const term = keyword.value.trim().toLowerCase();
  if (!term) {
    return scheduleRows.value;
  }
  return scheduleRows.value.filter((record) => {
    const title = (record.title || '').toLowerCase();
    const subject = (record.subject || '').toLowerCase();
    const questionType = String(record.questionType || '').toLowerCase();
    return title.includes(term) || subject.includes(term) || questionType.includes(term) || String(record.id).includes(term);
  });
});

const selectedIds = computed(() => Array.from(selectedMap.value));
const allSelected = computed(() => filteredRecords.value.length > 0 && filteredRecords.value.every((item) => selectedMap.value.has(item.id)));

onMounted(async () => {
  if (!records.value.length) {
    await recordStore.reload();
  }
  applyPrefillFromQuery();
  await loadJobs();
});

watch(
  () => route.query.prefill,
  () => {
    applyPrefillFromQuery();
  },
);

async function loadJobs() {
  loading.value = true;
  try {
    errorMessage.value = '';
    const res = await listPdfJobs();
    const rows = res?.data ?? res ?? [];
    jobs.value = Array.isArray(rows)
      ? rows.map((item: any) => ({
          jobId: String(item?.jobId || ''),
          status: String(item?.status || 'queued'),
          selected_count: Number(item?.selected_count || 0),
        }))
      : [];
  } catch (error: any) {
    errorMessage.value = error?.message || '加载 PDF 列表失败';
  } finally {
    loading.value = false;
  }
}

function openCreateDialog() {
  createDialogOpen.value = true;
}

function applyPrefillFromQuery() {
  if (prefillHandled.value) {
    return;
  }

  const raw = String(route.query.prefill || '').trim();
  if (!raw) {
    return;
  }

  const ids = raw
    .split(',')
    .map((item) => Number(item.trim()))
    .filter((item) => Number.isFinite(item) && item > 0);
  if (!ids.length) {
    prefillHandled.value = true;
    return;
  }

  const available = new Set(records.value.map((item) => item.id));
  const picked = ids.filter((id) => available.has(id));
  selectedMap.value = new Set(picked);
  createDialogOpen.value = true;
  prefillHandled.value = true;

  router.replace({ path: '/tabs/pdfs' });
}

function toggleRecord(id: number) {
  const next = new Set(selectedMap.value);
  if (next.has(id)) {
    next.delete(id);
  } else {
    next.add(id);
  }
  selectedMap.value = next;
}

function toggleSelectAll() {
  const next = new Set(selectedMap.value);
  if (allSelected.value) {
    filteredRecords.value.forEach((item) => next.delete(item.id));
  } else {
    filteredRecords.value.forEach((item) => next.add(item.id));
  }
  selectedMap.value = next;
}

async function createExport() {
  if (!selectedIds.value.length || exporting.value) {
    return;
  }

  exporting.value = true;
  try {
    errorMessage.value = '';
    const res = await exportPdfByRecordIds(selectedIds.value);
    const payload = res?.data ?? res ?? {};
    const jobId = String(payload?.jobId || '');
    if (!jobId) {
      throw new Error('未获取到导出任务号');
    }

    createDialogOpen.value = false;
    selectedMap.value = new Set<number>();
    keyword.value = '';
    await loadJobs();
    router.push(`/pdf/jobs/${jobId}`);
  } catch (error: any) {
    errorMessage.value = error?.message || '导出失败';
  } finally {
    exporting.value = false;
  }
}

function goJobDetail(jobId: string) {
  router.push(`/pdf/jobs/${jobId}`);
}

function formatTime(jobId: string) {
  const match = String(jobId || '').match(/job-(\d{13})/);
  const ts = Number(match?.[1] || 0);
  if (!Number.isFinite(ts) || ts <= 0) {
    return '未知时间';
  }
  return new Date(ts).toLocaleString();
}

function statusText(status: string) {
  if (status === 'done') return '已完成';
  if (status === 'queued') return '排队中';
  if (status === 'running' || status === 'processing') return '处理中';
  if (status === 'failed' || status === 'error') return '失败';
  return status || '未知';
}

function formatSubject(subject?: string) {
  const value = String(subject || '').trim().toLowerCase();
  if (value === 'math' || value === '数学') return '数学';
  if (value === 'physics' || value === '物理') return '物理';
  if (value === 'chemistry' || value === '化学') return '化学';
  if (value === 'biology' || value === '生物') return '生物';
  return subject || '未知';
}

function formatQuestionType(questionType?: string) {
  const value = String(questionType || '').trim().toLowerCase();
  if (['choice', '选择', '选择题', 'single_choice', 'multiple_choice'].includes(value)) return '选择';
  if (['fill_blank', '填空', '填空题'].includes(value)) return '填空';
  if (['essay', '解答', '解答题', 'subjective'].includes(value)) return '解答';
  return questionType || '未知';
}

function resultText(result?: string) {
  if (result === 'correct') return '上次正确';
  if (result === 'wrong') return '上次错误';
  return '未判定';
}

function resultClass(result?: string) {
  if (result === 'correct') return 'ok';
  if (result === 'wrong') return 'bad';
  return 'none';
}
</script>

<style scoped>
.page-wrap { gap: 12px; }
.hero-content { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.hero-label { margin: 0; font-size: 12px; color: #64748b; }
.hero-value { margin: 4px 0 0; font-size: 28px; font-weight: 700; color: #0f172a; }
.hero-sub { margin: 6px 0 0; font-size: 12px; color: #64748b; }
.toolbar { display: flex; justify-content: flex-end; }
.job-list { display: grid; gap: 8px; }
.job-item { width: 100%; border: 1px solid rgba(148, 163, 184, 0.28); border-radius: 12px; background: #fff; padding: 12px; display: flex; justify-content: space-between; align-items: center; }
.job-main { min-width: 0; }
.job-title { margin: 0; font-size: 14px; font-weight: 700; color: #0f172a; }
.job-meta { margin: 4px 0 0; font-size: 12px; color: #64748b; }
.item-arrow { color: #94a3b8; }
.empty-content { display: grid; gap: 10px; text-align: center; justify-items: center; }
.picker-toolbar { display: grid; grid-template-columns: 1fr auto; gap: 8px; }
.picker-list { max-height: 320px; overflow: auto; display: grid; gap: 6px; margin-top: 10px; }
.picker-item { border: 1px solid rgba(148, 163, 184, 0.22); border-radius: 10px; padding: 8px 10px; display: flex; align-items: center; gap: 10px; }
.picker-main { width: 100%; display: grid; gap: 6px; }
.title-row { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
.review-title { margin: 0; font-size: 14px; color: #0f172a; font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.review-meta { margin: 0; font-size: 12px; color: #64748b; }
.progress-row { display: flex; align-items: center; gap: 8px; }
.progress-track { position: relative; flex: 1; height: 7px; border-radius: 999px; overflow: hidden; background: rgba(148, 163, 184, 0.22); }
.progress-fill { height: 100%; border-radius: 999px; background: linear-gradient(90deg, #22c55e 0%, #0ea5e9 100%); }
.progress-fill.due { background: linear-gradient(90deg, #f59e0b 0%, #ef4444 100%); }
.progress-text { width: 42px; text-align: right; font-size: 11px; color: #64748b; }
.urgency-text { margin: 0; font-size: 11px; color: #b45309; }
.result-badge { font-size: 10px; border-radius: 999px; padding: 2px 8px; border: 1px solid transparent; white-space: nowrap; }
.result-badge.ok { color: #065f46; background: rgba(16, 185, 129, 0.15); border-color: rgba(16, 185, 129, 0.28); }
.result-badge.bad { color: #991b1b; background: rgba(248, 113, 113, 0.16); border-color: rgba(248, 113, 113, 0.3); }
.result-badge.none { color: #475569; background: rgba(148, 163, 184, 0.14); border-color: rgba(148, 163, 184, 0.26); }
.is-dark .hero-value, .is-dark .job-title, .is-dark .picker-main p { color: #f8fafc; }
.is-dark .hero-label, .is-dark .hero-sub, .is-dark .job-meta, .is-dark .picker-main span { color: #cbd5e1; }
.is-dark .job-item, .is-dark .picker-item { background: rgba(30, 41, 59, 0.94); border-color: rgba(148, 163, 184, 0.3); }
.is-dark .review-title { color: #f8fafc; }
.is-dark .review-meta, .is-dark .progress-text { color: #cbd5e1; }
.is-dark .progress-track { background: rgba(148, 163, 184, 0.28); }
.is-dark .urgency-text { color: #fbbf24; }
.is-dark .result-badge.none { color: #cbd5e1; }
@media (max-width: 640px) {
  .hero-content { align-items: flex-start; flex-direction: column; }
}
</style>
