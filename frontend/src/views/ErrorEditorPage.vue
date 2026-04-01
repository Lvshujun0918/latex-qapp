<template>
  <section class="app-page app-inner-page page-wrap">
    <header class="app-page-header page-header">
      <Button variant="outline" size="icon-sm" class="app-header-back" @click="goBack" aria-label="返回上一级"><</Button>
      <span class="app-kicker">Draft Studio</span>
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
          <Label for="latex-answer">最终答案（可手动修改）</Label>
          <Textarea id="latex-answer" v-model="latexAnswer" placeholder="例如：x=2" />
        </div>

        <div class="form-item">
          <Label for="latex-solution">分步解答（可手动填写）</Label>
          <Textarea id="latex-solution" v-model="latexSolution" placeholder="点击“生成解答”或手动输入" />
        </div>

        <h4>解答预览（LaTeX）</h4>
        <LatexView :source="latexSolution" class="latex-panel" />
      </CardContent>
    </Card>

    <Card v-else class="app-soft-card">
      <CardContent class="empty-tip">未找到拍照识别结果，请从底部中间“新增”按钮发起拍照。</CardContent>
    </Card>

    <Button class="save-btn" :disabled="!hasDraft || recordStore.loading" @click="save">保存错题</Button>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useRecordStore } from '@/stores/record';
import { clearVisionDraftStorage, generateSolutionByLatexStream, loadVisionDraftFromStorage } from '@/services/ai';
import LatexView from '@/components/LatexView.vue';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';

const router = useRouter();
const recordStore = useRecordStore();

const title = ref('');
const subject = ref('未知');
const questionType = ref('未知');
const latexSource = ref('');
const latexPreview = ref('');
const latexAnswer = ref('');
const latexSolution = ref('');
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
  latexAnswer.value = draft.latexAnswer;
  latexSolution.value = draft.latexSolution ?? '';
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
        latexAnswer.value = text;
      });
    }
    if (solved.latexSolution) {
      await streamTextDisplay(solved.latexSolution, (text) => {
        latexSolution.value = text;
      });
    }
  } catch (error: any) {
    errorMessage.value = error?.message || '解答生成失败，请稍后重试或手动填写。';
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
  try {
    errorMessage.value = '';
    await recordStore.save({
      subject: subject.value,
      question_type: questionType.value,
      title: title.value,
      latex_source: latexSource.value,
      latex_answer: latexAnswer.value,
      question_tags: tagList.value,
      mistake_reason: latexSolution.value,
    });

    router.replace('/tabs/errors');
  } catch (error: any) {
    errorMessage.value = error?.message || '保存失败，请重试。';
  }
}

function mapQuestionTypeLabel(type: string) {
  const value = String(type || '').trim().toLowerCase();
  if (['choice', '选择', '选择题', 'single_choice', 'multiple_choice'].includes(value)) {
    return '选择题';
  }
  if (['fill_blank', '填空', '填空题'].includes(value)) {
    return '填空题';
  }
  if (['essay', '解答', '解答题', 'subjective', '大题'].includes(value)) {
    return '大题';
  }
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

function assembleLatexFromQuestionJson(questionJson: any): string {
  if (!questionJson || typeof questionJson !== 'object') {
    return '';
  }

  const stem = String(questionJson.stem ?? '').trim();
  if (!stem) {
    return '';
  }

  const type = String(questionJson.question_type ?? '').trim();
  if (type === '选择' && Array.isArray(questionJson.options) && questionJson.options.length > 0) {
    const body = questionJson.options.map((opt: any) => `\\item ${String(opt ?? '').trim()}`).join('\n');
    return `${stem}\n\\begin{choices}\n${body}\n\\end{choices}`;
  }

  if (type === '解答' && Array.isArray(questionJson.sub_questions) && questionJson.sub_questions.length > 0) {
    const body = questionJson.sub_questions.map((sub: any) => `\\item ${String(sub ?? '').trim()}`).join('\n');
    return `${stem}\n\\begin{enumerate}\n${body}\n\\end{enumerate}`;
  }

  return stem;
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

.empty-tip {
  text-align: center;
  color: #475569;
  padding: 20px;
}

.save-btn {
  width: 100%;
}
</style>
