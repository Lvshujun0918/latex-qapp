<template>
  <section class="app-page app-inner-page page-wrap">
    <header class="app-page-header page-header">
      <Button variant="outline" size="icon-sm" class="app-header-back mr-4" @click="goBack" aria-label="返回上一级"><</Button>
      <h1>录入错题</h1>
      <p>核对识别结果并补全答案、步骤和标签。</p>
    </header>

    <Alert v-if="errorMessage" variant="destructive">
      <AlertTitle>处理失败</AlertTitle>
      <AlertDescription>{{ errorMessage }}</AlertDescription>
    </Alert>

    <Card v-if="hasDraft" class="app-page-shell">
      <CardHeader>
        <CardDescription>{{ subjectLabel }} · {{ questionTypeLabel }}</CardDescription>
        <CardTitle>{{ title || '识别题目' }}</CardTitle>
      </CardHeader>
      <CardContent class="editor-content">
        <h4>题目（LaTeX）</h4>
        <LatexView :source="latexPreview" class="latex-panel" />

        <h4>题目标签</h4>
        <div class="tag-wrap">
          <Badge v-for="tag in tagList" :key="tag" variant="outline">{{ tag }}</Badge>
          <Badge v-if="!tagList.length" variant="outline">未识别到标签</Badge>
        </div>

        <div class="solve-header">
          <h4>答案与解答</h4>
          <Button variant="outline" size="sm" :disabled="isSolving" @click="generateSolve">
            {{ isSolving ? solvingStage || '生成中...' : '生成解答' }}
          </Button>
        </div>

        <div class="form-item">
          <Label>答案来源</Label>
          <div class="mode-row">
            <Button size="sm" :variant="answerMode === 'ai' ? 'default' : 'outline'" @click="setAnswerMode('ai')">AI 解析</Button>
            <Button size="sm" :variant="answerMode === 'image' ? 'default' : 'outline'" @click="setAnswerMode('image')">拍照/相册上传</Button>
          </div>
          <div v-if="answerMode === 'ai'" class="preview-wrap">
            <LatexView :source="answerText || '暂无答案，点击上方“生成解答”'" class="latex-panel" />
          </div>
          <div v-else class="preview-wrap">
            <Button variant="outline" size="sm" @click="openImagePicker('answer')">选择答案图片</Button>
            <img v-if="answerImageDataUrl" :src="answerImageDataUrl" alt="答案图片" class="upload-preview" />
          </div>
        </div>

        <div class="form-item">
          <Label>解析来源</Label>
          <div class="mode-row">
            <Button size="sm" :variant="analysisMode === 'ai' ? 'default' : 'outline'" @click="setAnalysisMode('ai')">AI 解析</Button>
            <Button size="sm" :variant="analysisMode === 'image' ? 'default' : 'outline'" @click="setAnalysisMode('image')">拍照/相册上传</Button>
          </div>
          <div v-if="analysisMode === 'ai'" class="preview-wrap">
            <MarkdownView :source="analysisText || '暂无解析，点击上方“生成解答”'" class="latex-panel" />
          </div>
          <div v-else class="preview-wrap">
            <Button variant="outline" size="sm" @click="openImagePicker('analysis')">选择解析图片</Button>
            <img v-if="analysisImageDataUrl" :src="analysisImageDataUrl" alt="解析图片" class="upload-preview" />
          </div>
        </div>
      </CardContent>
    </Card>

    <Card v-else class="app-soft-card">
      <CardContent class="empty-tip">未找到拍照识别结果，请从底部中间“新增”按钮发起拍照。</CardContent>
    </Card>

    <Button class="save-btn" :disabled="!hasDraft || recordStore.loading" @click="save">保存错题</Button>

    <ImageSourceDialog
      :open="sourceDialogOpen"
      @update:open="(val) => (sourceDialogOpen = val)"
      @select="pickModeImage"
    />
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useRecordStore } from '@/stores/record';
import { clearVisionDraftStorage, generateSolutionByLatexStream, loadVisionDraftFromStorage, pickImageAsDataUrl } from '@/services/ai';
import { assembleLatexFromQuestionJson } from '@/utils/question-format';
import ImageSourceDialog from '@/components/ImageSourceDialog.vue';
import LatexView from '@/components/LatexView.vue';
import MarkdownView from '@/components/MarkdownView.vue';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';

const router = useRouter();
const recordStore = useRecordStore();

const title = ref('');
const subject = ref('未知');
const questionType = ref('未知');
const latexSource = ref('');
const latexPreview = ref('');
const answerMode = ref<'ai' | 'image'>('ai');
const answerText = ref('');
const answerImageDataUrl = ref('');
const analysisMode = ref<'ai' | 'image'>('ai');
const analysisText = ref('');
const analysisImageDataUrl = ref('');
const sourceDialogOpen = ref(false);
const pickingTarget = ref<'answer' | 'analysis'>('answer');
const questionTags = ref<string[]>([]);
const isSolving = ref(false);
const solvingStage = ref('');
const errorMessage = ref('');

const hasDraft = computed(() => latexSource.value.trim().length > 0);
const tagList = computed(() => questionTags.value.filter((item) => item.trim().length > 0));
const questionTypeLabel = computed(() => mapQuestionTypeLabel(questionType.value));
const subjectLabel = computed(() => mapSubjectLabel(subject.value));

onMounted(() => {
  const draft = loadVisionDraftFromStorage();
  if (!draft) {
    return;
  }

  title.value = draft.title ?? title.value;
  subject.value = draft.subject ?? subject.value;
  questionType.value = draft.questionType ?? questionType.value;
  latexSource.value = draft.latexSource || '';
  latexPreview.value = draft.latexQuestion || assembleLatexFromQuestionJson(draft.questionJson);
  answerMode.value = 'ai';
  answerText.value = draft.latexAnswer;
  analysisMode.value = 'ai';
  analysisText.value = draft.latexSolution ?? '';
  questionTags.value = draft.tags;
  clearVisionDraftStorage();
});

async function generateSolve() {
  if (isSolving.value || !hasDraft.value) {
    return;
  }

  try {
    errorMessage.value = '';
    isSolving.value = true;
    solvingStage.value = '准备解答...';
    const solved = await generateSolutionByLatexStream(
      {
        latexQuestion: latexPreview.value,
        questionType: questionType.value,
        subject: subject.value,
      },
      (evt) => {
        if (evt.stage === 'solve_start') {
          solvingStage.value = '正在推理...';
        }
        if (evt.stage === 'solve_final') {
          solvingStage.value = '解答完成';
        }
      },
    );

    // 逐字显示答案和解析
    if (solved.latexAnswer) {
      await streamTextDisplay(solved.latexAnswer, (text) => {
        answerText.value = text;
      });
    }
    if (solved.latexSolution) {
      await streamTextDisplay(solved.latexSolution, (text) => {
        analysisText.value = text;
      });
    }

    answerMode.value = 'ai';
    analysisMode.value = 'ai';
  } catch (error: any) {
    errorMessage.value = error?.message || '解答生成失败，请稍后重试。';
  } finally {
    isSolving.value = false;
    solvingStage.value = '';
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

async function save() {
  if (answerMode.value === 'image' && !answerImageDataUrl.value) {
    errorMessage.value = '请先上传答案图片。';
    return;
  }
  if (analysisMode.value === 'image' && !analysisImageDataUrl.value) {
    errorMessage.value = '请先上传解析图片。';
    return;
  }

  try {
    errorMessage.value = '';
    await recordStore.save({
      subject: subject.value,
      question_type: questionType.value,
      title: title.value,
      latex_source: latexSource.value,
      answer_mode: answerMode.value,
      answer_text: answerMode.value === 'ai' ? answerText.value : '',
      answer_image_data_url: answerMode.value === 'image' ? answerImageDataUrl.value : '',
      analysis_mode: analysisMode.value,
      analysis_text: analysisMode.value === 'ai' ? analysisText.value : '',
      analysis_image_data_url: analysisMode.value === 'image' ? analysisImageDataUrl.value : '',
      question_tags: tagList.value,
    });

    router.replace('/tabs/errors');
  } catch (error: any) {
    errorMessage.value = error?.message || '保存失败，请重试。';
  }
}

function setAnswerMode(mode: 'ai' | 'image') {
  answerMode.value = mode;
  if (mode === 'ai') {
    answerImageDataUrl.value = '';
  } else {
    answerText.value = '';
  }
}

function setAnalysisMode(mode: 'ai' | 'image') {
  analysisMode.value = mode;
  if (mode === 'ai') {
    analysisImageDataUrl.value = '';
  } else {
    analysisText.value = '';
  }
}

function openImagePicker(target: 'answer' | 'analysis') {
  pickingTarget.value = target;
  sourceDialogOpen.value = true;
}

async function pickModeImage(source: 'camera' | 'album') {
  try {
    errorMessage.value = '';
    const dataUrl = await pickImageAsDataUrl(source);
    if (pickingTarget.value === 'answer') {
      answerMode.value = 'image';
      answerImageDataUrl.value = dataUrl;
      answerText.value = '';
      return;
    }

    analysisMode.value = 'image';
    analysisImageDataUrl.value = dataUrl;
    analysisText.value = '';
  } catch (error: any) {
    errorMessage.value = error?.message || '上传图片失败，请重试。';
  }
}

function mapQuestionTypeLabel(type: string) {
  const value = String(type || '').trim().toLowerCase();
  if (['choice', '选择', '选择题', 'single_choice', 'multiple_choice'].includes(value)) return '选择题';
  if (['fill_blank', '填空', '填空题'].includes(value)) return '填空题';
  if (['essay', '解答', '解答题', 'subjective', '大题'].includes(value)) return '大题';
  return '未知题型';
}

function mapSubjectLabel(value: string) {
  const subjectValue = String(value || '').trim().toLowerCase();
  if (subjectValue === 'math' || subjectValue === '数学') return '数学';
  if (subjectValue === 'physics' || subjectValue === '物理') return '物理';
  if (subjectValue === 'chemistry' || subjectValue === '化学') return '化学';
  if (subjectValue === 'biology' || subjectValue === '生物') return '生物';
  return value || '未知';
}

function goBack() {
  if (window.history.length > 1) {
    router.back();
    return;
  }

  router.replace('/tabs/errors');
}
</script>

<style scoped>
.page-wrap {
  gap: 12px;
}

.editor-content {
  display: grid;
  gap: 10px;
}

h4 {
  margin: 8px 0 4px;
  font-size: 13px;
  color: #475569;
}

.latex-panel {
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.52);
  border: 1px solid rgba(148, 163, 184, 0.36);
  padding: 10px 12px;
}

.tag-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.solve-header {
  margin-top: 6px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.solve-header h4 {
  margin: 0;
}

.form-item {
  display: grid;
  gap: 8px;
}

.mode-row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.preview-wrap {
  display: grid;
  gap: 8px;
}

.upload-preview {
  width: 100%;
  max-height: 280px;
  object-fit: contain;
  border-radius: 12px;
  border: 1px solid rgba(148, 163, 184, 0.3);
  background: rgba(248, 250, 252, 0.75);
}

.empty-tip {
  text-align: center;
  color: #475569;
  padding: 20px;
}

.save-btn {
  width: 100%;
}
</style>
