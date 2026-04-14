<template>
  <section class="app-page app-inner-page page-wrap pt-8" :class="{ 'is-dark': resolvedTheme === 'dark' }" v-if="record">
    <header class="app-page-header page-header">
      <Button variant="outline" size="icon-sm" class="app-header-back mr-4" @click="goBack" aria-label="返回上一级"><</Button>
      <h1>错题详情</h1>
      <p>查看题目元信息与 AI 解析结果。</p>
    </header>

    <Card class="app-page-shell detail-shell">
      <CardHeader class="detail-header p-0 pb-3">
        <CardTitle>{{ record.title || '未命名题目' }}</CardTitle>
      </CardHeader>
      <CardContent class="detail-content p-0">
        <div class="app-meta-grid">
          <p class="app-meta-item">学科：{{ formatSubject(record.subject) }}</p>
          <p class="app-meta-item">题型：{{ formatQuestionType(record.questionType) }}</p>
        </div>
        <LatexView :source="record.latexSource || ''" class="latex-block" />
      </CardContent>
    </Card>

    <Card class="app-page-shell detail-shell">
      <CardHeader class="detail-header p-0 pb-3">
        <CardTitle>答案与解析</CardTitle>
      </CardHeader>
      <CardContent class="detail-content p-0">
        <div class="analysis-generate">
          <p>可切换为 AI 文本解析或上传一张答案解析图片。</p>
          <div class="mode-row">
            <Button size="sm" :variant="solutionMode === 'ai' ? 'default' : 'outline'" @click="setSolutionMode('ai')">AI 解析</Button>
            <Button size="sm" :variant="solutionMode === 'image' ? 'default' : 'outline'" @click="setSolutionMode('image')">上传图片</Button>
            <Button v-if="solutionMode === 'image'" size="sm" variant="outline" @click="openImagePicker">选择图片</Button>
            <Button size="sm" variant="outline" :disabled="savingMode" @click="saveSolutionMode">保存设置</Button>
          </div>
          <img v-if="solutionMode === 'image' && solutionImageDataUrl" :src="solutionImageDataUrl" alt="答案解析图片" class="media-preview" />
        </div>

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
            <img
              v-if="solutionMode === 'image' && solutionImageDataUrl"
              :src="solutionImageDataUrl"
              alt="答案图片"
              class="media-preview"
            />
            <LatexView v-else :source="answerText || '暂无答案'" class="latex-block" />
          </div>
        </div>

        <div v-if="hasAnalysis" class="analysis-block">
          <button class="collapse-trigger" type="button" @click="analysisOpen = !analysisOpen">
            <span>分步解答 / 错因分析</span>
            <span>{{ analysisOpen ? '收起' : '展开' }}</span>
          </button>
          <div v-show="analysisOpen" class="collapse-content">
            <img
              v-if="solutionMode === 'image' && solutionImageDataUrl"
              :src="solutionImageDataUrl"
              alt="解析图片"
              class="media-preview"
            />
            <MarkdownView v-else :source="analysisText || '暂无解析'" class="markdown-block" />
          </div>
        </div>

        <p v-if="errorMessage" class="error-tip">{{ errorMessage }}</p>
      </CardContent>
    </Card>

    <ImageSourceDialog
      :open="sourceDialogOpen"
      @update:open="(val) => (sourceDialogOpen = val)"
      @select="pickModeImage"
    />
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useRecordStore } from '@/stores/record';
import { useTheme } from '@/composables/useTheme';
import LatexView from '@/components/LatexView.vue';
import MarkdownView from '@/components/MarkdownView.vue';
import { generateSolutionByLatexStream, pickImageAsDataUrl } from '@/services/ai';
import ImageSourceDialog from '@/components/ImageSourceDialog.vue';
import { toSavePayload } from '@/services/records';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

const route = useRoute();
const router = useRouter();
const recordStore = useRecordStore();
const { resolvedTheme } = useTheme();
const generatingAi = ref(false);
const savingMode = ref(false);
const answerOpen = ref(false);
const analysisOpen = ref(false);
const solutionMode = ref<'ai' | 'image'>('ai');
const answerText = ref('');
const analysisText = ref('');
const solutionImageDataUrl = ref('');
const sourceDialogOpen = ref(false);
const errorMessage = ref('');

const record = computed(() => recordStore.records.find((r) => r.id === Number(route.params.id)));
const hasAnswer = computed(() =>
  solutionMode.value === 'image'
    ? solutionImageDataUrl.value.trim().length > 0
    : answerText.value.trim().length > 0,
);
const hasAnalysis = computed(() =>
  solutionMode.value === 'image'
    ? solutionImageDataUrl.value.trim().length > 0
    : analysisText.value.trim().length > 0,
);
const needsAiGenerate = computed(() => !hasAnswer.value || !hasAnalysis.value);

watch(
  record,
  (value) => {
    solutionMode.value = value?.solutionMode === 'image' ? 'image' : 'ai';
    answerText.value = value?.answerText ?? '';
    analysisText.value = value?.analysisText ?? '';
    solutionImageDataUrl.value = value?.solutionImageDataUrl ?? '';
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
    errorMessage.value = '';
    const solved = await generateSolutionByLatexStream({
      subject: record.value.subject,
      questionType: record.value.questionType,
      latexQuestion: record.value.latexSource,
      latexSource: record.value.latexSource,
    }, () => {});

    // 逐字显示答案
    if (solved.latexAnswer) {
      await streamTextDisplay(solved.latexAnswer, (text) => {
        answerText.value = text;
      });
    }

    // 逐字显示解析
    if (solved.latexSolution) {
      await streamTextDisplay(solved.latexSolution, (text) => {
        analysisText.value = text;
      });
    }

    await recordStore.updateById(record.value.id, {
      ...toSavePayload(record.value),
      solution_mode: 'ai',
      answer_text: answerText.value || solved.latexAnswer,
      analysis_text: analysisText.value || solved.latexSolution,
      solution_image_data_url: '',
    });

    solutionMode.value = 'ai';
    solutionImageDataUrl.value = '';

    answerOpen.value = true;
    analysisOpen.value = true;
  } catch (error: any) {
    errorMessage.value = error?.message || 'AI 生成失败，请重试。';
  } finally {
    generatingAi.value = false;
  }
}

function setSolutionMode(mode: 'ai' | 'image') {
  solutionMode.value = mode;
  if (mode === 'ai') {
    solutionImageDataUrl.value = '';
    return;
  }
  answerText.value = '';
  analysisText.value = '';
}

function openImagePicker() {
  sourceDialogOpen.value = true;
}

async function pickModeImage(source: 'camera' | 'album') {
  try {
    errorMessage.value = '';
    solutionMode.value = 'image';
    solutionImageDataUrl.value = await pickImageAsDataUrl(source);
    answerText.value = '';
    analysisText.value = '';
  } catch (error: any) {
    errorMessage.value = error?.message || '上传图片失败，请重试。';
  }
}

async function saveSolutionMode() {
  if (!record.value || savingMode.value) {
    return;
  }
  if (solutionMode.value === 'image' && !solutionImageDataUrl.value) {
    errorMessage.value = '请先上传答案解析图片。';
    return;
  }

  savingMode.value = true;
  try {
    errorMessage.value = '';
    await recordStore.updateById(record.value.id, {
      ...toSavePayload(record.value),
      solution_mode: solutionMode.value,
      answer_text: solutionMode.value === 'ai' ? answerText.value : '',
      analysis_text: solutionMode.value === 'ai' ? analysisText.value : '',
      solution_image_data_url: solutionMode.value === 'image' ? solutionImageDataUrl.value : '',
    });
  } catch (error: any) {
    errorMessage.value = error?.message || '保存失败，请重试。';
  } finally {
    savingMode.value = false;
  }
}

async function streamTextDisplay(text: string, onUpdate: (text: string) => void) {
  let displayed = '';
  const chars = text.split('');
  for (const char of chars) {
    displayed += char;
    onUpdate(displayed);
    // 每个字符延迟2ms，营造逐字显示效果
    await new Promise((resolve) => setTimeout(resolve, 2));
  }
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

.mode-row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
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

.error-tip {
  margin: 0;
  color: #dc2626;
  font-size: 13px;
}

.media-preview {
  width: 100%;
  max-height: 340px;
  object-fit: contain;
  border-radius: 10px;
  border: 1px solid rgba(148, 163, 184, 0.28);
  background: rgba(248, 250, 252, 0.78);
}

.is-dark .collapse-trigger {
  border-color: rgba(148, 163, 184, 0.28);
  background: rgba(30, 41, 59, 0.58);
  color: #e2e8f0;
}
</style>
