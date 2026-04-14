<template>
  <section class="app-page app-inner-page page-wrap" :class="{ 'is-dark': resolvedTheme === 'dark' }" v-if="record">
    <header class="app-page-header page-header">
      <Button variant="outline" size="icon-sm" class="app-header-back mr-4" @click="goBack" aria-label="返回上一级"><</Button>
      <h1>家长复习批改</h1>
      <p>孩子线下作答后，家长在这里查看答案并登记对错，系统会自动更新复习节奏。</p>
    </header>

    <Card class="app-page-shell session-hero">
      <CardContent class="session-hero-content">
        <div>
          <p class="session-label">当前阶段</p>
          <p class="session-value">{{ stageLabel }}</p>
          <p class="session-sub">预计下次间隔 {{ nextIntervalPreview }} 天 · 已复习 {{ record.reviewCount }} 次</p>
        </div>
        <div class="session-progress-wrap">
          <div class="session-progress-track">
            <div class="session-progress-fill" :style="{ width: `${record.masteryLevel}%` }" />
          </div>
          <span class="session-progress-text">掌握度 {{ record.masteryLevel }}%</span>
        </div>
      </CardContent>
    </Card>

    <Card class="app-page-shell">
      <CardHeader>
        <CardTitle>{{ record.title || '未命名题目' }}</CardTitle>
      </CardHeader>
      <CardContent class="review-content">
        <div class="meta-row">
          <Badge variant="outline">{{ formatSubject(record.subject) }}</Badge>
          <Badge variant="outline">{{ formatQuestionType(record.questionType) }}</Badge>
          <Badge variant="outline">当前掌握度 {{ record.masteryLevel }}%</Badge>
        </div>

        <div v-if="stage === 'waiting'" class="actions-row">
          <Button @click="revealAnswer">孩子已完成作答，查看参考答案</Button>
        </div>

        <LatexView :source="record.latexSource || ''" class="latex-block" />

        <div v-if="stage !== 'waiting'" class="answer-block">
          <h4>标准答案</h4>
          <img
            v-if="record.solutionMode === 'image' && record.solutionImageDataUrl"
            :src="record.solutionImageDataUrl"
            alt="标准答案图片"
            class="media-preview"
          />
          <LatexView v-else :source="record.answerText || '暂无答案'" class="latex-block" />
        </div>

        <div v-if="stage === 'revealed'" class="judge-row">
          <Button :disabled="savingResult" @click="markResult(true)">孩子做对了</Button>
          <Button variant="destructive" :disabled="savingResult" @click="markResult(false)">孩子做错了</Button>
        </div>

        <div v-if="stage === 'judged-wrong'" class="analysis-block">
          <h4>解析</h4>
          <img
            v-if="record.solutionMode === 'image' && record.solutionImageDataUrl"
            :src="record.solutionImageDataUrl"
            alt="解析图片"
            class="media-preview"
          />
          <MarkdownView v-else :source="record.analysisText || '暂无解析'" class="markdown-block" />
        </div>

        <Alert v-if="errorMessage" variant="destructive">
          <AlertTitle>提交失败</AlertTitle>
          <AlertDescription>{{ errorMessage }}</AlertDescription>
        </Alert>

        <div v-if="stage === 'judged-right' || stage === 'judged-wrong'" class="next-row">
          <p>本次批改结果已记录，已纳入后续复习排序。</p>
          <Button variant="outline" @click="goBackToReview">返回复习列表</Button>
        </div>
      </CardContent>
    </Card>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useTheme } from '@/composables/useTheme';
import { useRecordStore } from '@/stores/record';
import { toSavePayload } from '@/services/records';
import LatexView from '@/components/LatexView.vue';
import MarkdownView from '@/components/MarkdownView.vue';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

type ReviewStage = 'waiting' | 'revealed' | 'judged-right' | 'judged-wrong';

const route = useRoute();
const router = useRouter();
const recordStore = useRecordStore();
const { resolvedTheme } = useTheme();

const stage = ref<ReviewStage>('waiting');
const savingResult = ref(false);
const errorMessage = ref('');

const record = computed(() => recordStore.records.find((item) => item.id === Number(route.params.id)));

const EBBINGHAUS_INTERVALS = [1, 2, 4, 7, 15, 30];

const nextIntervalPreview = computed(() => {
  if (!record.value) {
    return 1;
  }
  const reviewCount = Math.max(0, Number(record.value.reviewCount || 0));
  const intervalIndex = Math.min(reviewCount, EBBINGHAUS_INTERVALS.length - 1);
  const baseInterval = EBBINGHAUS_INTERVALS[intervalIndex];
  const fallbackFactor = Math.max(1.3, Math.min(3.0, Number(record.value.reviewEaseFactor || 2.5)));
  const mastery = Math.max(0, Math.min(100, Number(record.value.masteryLevel || 0)));
  const masteryFactor = mastery >= 80 ? 1.2 : mastery >= 60 ? 1.0 : mastery >= 40 ? 0.92 : 0.8;
  return Math.max(1, Math.round(baseInterval * fallbackFactor * masteryFactor * 0.55));
});

const stageLabel = computed(() => {
  if (stage.value === 'waiting') return '等待孩子作答';
  if (stage.value === 'revealed') return '家长判定中';
  if (stage.value === 'judged-right') return '孩子本题答对';
  return '孩子本题答错';
});

onMounted(async () => {
  if (!recordStore.records.length) {
    await recordStore.reload();
  }
  if (!record.value) {
    router.replace('/tabs/review');
  }
});

function revealAnswer() {
  stage.value = 'revealed';
}

async function markResult(isCorrect: boolean) {
  if (!record.value || savingResult.value) {
    return;
  }

  savingResult.value = true;
  errorMessage.value = '';
  try {
    const nextReviewCount = isCorrect
      ? record.value.reviewCount + 1
      : Math.max(0, record.value.reviewCount - 1);

    const nextMasteryLevel = isCorrect
      ? Math.min(100, Math.round(record.value.masteryLevel * 0.85 + 15))
      : Math.max(0, Math.round(record.value.masteryLevel * 0.7));

    const currentEase = Math.max(1.3, Math.min(3.0, Number(record.value.reviewEaseFactor || 2.5)));
    const nextEase = isCorrect
      ? Math.min(3.0, Number((currentEase + 0.12).toFixed(2)))
      : Math.max(1.3, Number((currentEase - 0.2).toFixed(2)));
    const reviewedAt = new Date();
    const intervalDays = Math.max(1, Math.round(nextIntervalPreview.value * (isCorrect ? 1 : 0.65)));
    const nextReviewAt = new Date(reviewedAt.getTime() + intervalDays * 86400000);

    await recordStore.updateById(record.value.id, {
      ...toSavePayload(record.value),
      mastery_level: nextMasteryLevel,
      review_count: nextReviewCount,
      review_ease_factor: nextEase,
      last_review_result: isCorrect ? 'correct' : 'wrong',
      last_reviewed_at: reviewedAt.toISOString(),
      next_review_at: nextReviewAt.toISOString(),
    });

    stage.value = isCorrect ? 'judged-right' : 'judged-wrong';
  } catch (error: any) {
    errorMessage.value = error?.message || '记录复习结果失败，请重试。';
  } finally {
    savingResult.value = false;
  }
}

function goBack() {
  if (window.history.length > 1) {
    router.back();
    return;
  }
  router.replace('/tabs/review');
}

function goBackToReview() {
  router.replace('/tabs/review');
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
</script>

<style scoped>
.page-wrap {
  gap: 12px;
}

.session-hero {
  overflow: hidden;
}

.session-hero-content {
  padding: 16px;
  border-radius: 14px;
  display: grid;
  gap: 12px;
  background: rgba(37, 99, 235, 0.08);
}

.session-label {
  margin: 0;
  font-size: 12px;
  color: #0369a1;
}

.session-value {
  margin: 6px 0 0;
  font-size: 24px;
  line-height: 1.15;
  font-weight: 700;
  color: #0f172a;
}

.session-sub {
  margin: 6px 0 0;
  font-size: 13px;
  color: #475569;
}

.session-progress-wrap {
  display: grid;
  gap: 8px;
}

.session-progress-track {
  height: 10px;
  border-radius: 999px;
  overflow: hidden;
  background: rgba(148, 163, 184, 0.24);
}

.session-progress-fill {
  height: 100%;
  border-radius: inherit;
  background: #2563eb;
}

.session-progress-text {
  font-size: 12px;
  color: #475569;
}

.review-content {
  display: grid;
  gap: 12px;
}

.meta-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.actions-row,
.judge-row,
.next-row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
}

.answer-block,
.analysis-block {
  display: grid;
  gap: 8px;
  border: 1px solid rgba(148, 163, 184, 0.24);
  border-radius: 12px;
  padding: 12px;
  background: rgba(248, 250, 252, 0.75);
}

.media-preview {
  width: 100%;
  max-height: 320px;
  object-fit: contain;
  border-radius: 10px;
  border: 1px solid rgba(148, 163, 184, 0.28);
  background: rgba(248, 250, 252, 0.78);
}

.answer-block h4,
.analysis-block h4 {
  margin: 0;
  font-size: 13px;
  color: #64748b;
}

.next-row p {
  margin: 0;
  color: #64748b;
  font-size: 13px;
}

.is-dark .session-hero-content {
  background: rgba(37, 99, 235, 0.16);
}

.is-dark .session-label,
.is-dark .session-progress-text {
  color: #7dd3fc;
}

.is-dark .session-value {
  color: #f8fafc;
}

.is-dark .session-sub,
.is-dark .next-row p,
.is-dark .answer-block h4,
.is-dark .analysis-block h4 {
  color: #cbd5e1;
}

.is-dark .session-progress-track {
  background: rgba(148, 163, 184, 0.3);
}

.is-dark .form-item,
.is-dark .answer-block,
.is-dark .analysis-block {
  background: rgba(15, 23, 42, 0.5);
  border-color: rgba(148, 163, 184, 0.28);
}
</style>
