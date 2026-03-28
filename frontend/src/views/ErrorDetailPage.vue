<template>
  <section class="app-page app-inner-page page-wrap pt-8" :class="{ 'is-dark': resolvedTheme === 'dark' }" v-if="record">
    <header class="app-page-header page-header">
      <Button variant="outline" size="icon-sm" class="app-header-back" @click="goBack" aria-label="返回上一级"><</Button>
      <span class="app-kicker">Error Insight</span>
      <h1>错题详情</h1>
      <p>查看题目元信息与 AI 解析结果。</p>
    </header>

    <Card class="app-page-shell detail-shell">
      <CardHeader class="detail-header p-0 pb-3">
        <CardTitle>{{ record.title || '未命名题目' }}</CardTitle>
      </CardHeader>
      <CardContent class="detail-content p-0">
        <div class="app-meta-grid">
          <p class="app-meta-item">学科：{{ record.subject }}</p>
          <p class="app-meta-item">题型：{{ record.questionType || 'unknown' }}</p>
          <p class="app-meta-item">难度：{{ record.difficulty }}</p>
        </div>
        <LatexView :source="record.latexSource || ''" class="latex-block" />
      </CardContent>
    </Card>

    <Card class="app-page-shell detail-shell">
      <CardHeader class="detail-header p-0 pb-3">
        <CardTitle>AI 解析</CardTitle>
      </CardHeader>
      <CardContent class="detail-content p-0">
        <div v-if="needsAiGenerate" class="analysis-generate">
          <p>当前题目尚未生成 AI 答案与解析。</p>
          <Button :disabled="generatingAi" @click="generateAiSolution">
            {{ generatingAi ? '生成中...' : '一键生成答案与解析' }}
          </Button>
        </div>

        <div v-if="hasAnswer" class="analysis-block">
          <button class="collapse-trigger" type="button" @click="answerOpen = !answerOpen">
            <span>最终答案</span>
            <span>{{ answerOpen ? '收起' : '展开' }}</span>
          </button>
          <div v-show="answerOpen" class="collapse-content">
            <LatexView :source="answerText || '暂无答案'" class="latex-block" />
          </div>
        </div>

        <div v-if="hasAnalysis" class="analysis-block">
          <button class="collapse-trigger" type="button" @click="analysisOpen = !analysisOpen">
            <span>分步解答 / 错因分析</span>
            <span>{{ analysisOpen ? '收起' : '展开' }}</span>
          </button>
          <div v-show="analysisOpen" class="collapse-content">
            <MarkdownView :source="analysisText || '暂无解析'" class="markdown-block" />
          </div>
        </div>
      </CardContent>
    </Card>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useRecordStore } from '@/stores/record';
import { useTheme } from '@/composables/useTheme';
import LatexView from '@/components/LatexView.vue';
import MarkdownView from '@/components/MarkdownView.vue';
import { generateSolutionByLatexStream } from '@/services/ai';
import { toSavePayload } from '@/services/records';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

const route = useRoute();
const router = useRouter();
const recordStore = useRecordStore();
const { resolvedTheme } = useTheme();
const generatingAi = ref(false);
const answerOpen = ref(false);
const analysisOpen = ref(false);
const answerText = ref('');
const analysisText = ref('');

const record = computed(() => recordStore.records.find((r) => r.id === Number(route.params.id)));
const hasAnswer = computed(() => answerText.value.trim().length > 0);
const hasAnalysis = computed(() => analysisText.value.trim().length > 0);
const needsAiGenerate = computed(() => !hasAnswer.value || !hasAnalysis.value);

watch(
  record,
  (value) => {
    answerText.value = value?.latexAnswer ?? '';
    analysisText.value = value?.mistakeReason ?? '';
  },
  { immediate: true },
);

onMounted(() => {
  recordStore.reload();
});

function goBack() {
  if (window.history.length > 1) {
    router.back();
    return;
  }

  router.replace('/tabs/errors');
}

async function generateAiSolution() {
  if (!record.value || generatingAi.value) {
    return;
  }

  generatingAi.value = true;
  try {
    const solved = await generateSolutionByLatexStream({
      subject: record.value.subject,
      questionType: record.value.questionType,
      latexQuestion: record.value.latexSource,
    }, () => {});

    answerText.value = solved.latexAnswer || answerText.value;
    analysisText.value = solved.latexSolution || analysisText.value;

    await recordStore.updateById(record.value.id, {
      ...toSavePayload(record.value),
      latex_answer: answerText.value,
      mistake_reason: analysisText.value,
    });

    answerOpen.value = true;
    analysisOpen.value = true;
  } finally {
    generatingAi.value = false;
  }
}
</script>

<style scoped>
.page-wrap {
  gap: 14px;
}

.detail-shell {
  overflow: hidden;
}

.detail-content {
  display: grid;
  gap: 12px;
}

.analysis-block {
  display: grid;
  gap: 8px;
}

.analysis-generate {
  display: grid;
  gap: 10px;
}

.analysis-generate p {
  margin: 0;
  color: #64748b;
  font-size: 13px;
}

.analysis-block h4 {
  margin: 0;
  font-size: 13px;
  color: #64748b;
}

.collapse-trigger {
  width: 100%;
  border: 1px solid rgba(148, 163, 184, 0.32);
  border-radius: 12px;
  background: rgba(241, 245, 249, 0.5);
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  cursor: pointer;
  color: #334155;
  font-size: 13px;
  font-weight: 600;
}

.collapse-content {
  padding: 0 2px;
}

.is-dark .collapse-trigger {
  border-color: rgba(148, 163, 184, 0.28);
  background: rgba(30, 41, 59, 0.58);
  color: #e2e8f0;
}
</style>
