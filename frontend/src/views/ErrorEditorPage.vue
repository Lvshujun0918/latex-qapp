<template>
  <section class="page-wrap">
    <header class="page-header">
      <h1>录入错题</h1>
      <p>核对识别结果并补全答案、步骤和标签。</p>
    </header>

    <Card v-if="hasDraft">
      <CardHeader>
        <CardDescription>{{ subject || 'unknown' }} · {{ questionTypeLabel }}</CardDescription>
        <CardTitle>{{ title || '识别题目' }}</CardTitle>
      </CardHeader>
      <CardContent class="editor-content">
        <h4>题目（LaTeX）</h4>
        <LatexView :source="latexSource" class="latex-panel" />

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

    <Card v-else>
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
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';

const router = useRouter();
const recordStore = useRecordStore();

const title = ref('');
const subject = ref('math');
const questionType = ref('unknown');
const latexSource = ref('');
const latexAnswer = ref('');
const latexSolution = ref('');
const questionTags = ref<string[]>([]);
const isSolving = ref(false);
const solvingStage = ref('');

const hasDraft = computed(() => latexSource.value.trim().length > 0);
const tagList = computed(() => questionTags.value.filter((item) => item.trim().length > 0));
const questionTypeLabel = computed(() => mapQuestionTypeLabel(questionType.value));

onMounted(() => {
  const draft = loadVisionDraftFromStorage();
  if (!draft) {
    return;
  }

  title.value = draft.title ?? title.value;
  subject.value = draft.subject ?? subject.value;
  questionType.value = draft.questionType ?? questionType.value;
  latexSource.value = draft.latexQuestion;
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
    isSolving.value = true;
    solvingStage.value = '准备解答...';
    const solved = await generateSolutionByLatexStream(
      {
        latexQuestion: latexSource.value,
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

    latexAnswer.value = solved.latexAnswer || latexAnswer.value;
    latexSolution.value = solved.latexSolution || latexSolution.value;
  } catch {
    window.alert('解答生成失败，请稍后重试或手动填写。');
  } finally {
    isSolving.value = false;
    solvingStage.value = '';
  }
}

async function save() {
  try {
    await recordStore.save({
      subject: subject.value,
      question_type: questionType.value,
      difficulty: 3,
      title: title.value,
      latex_source: latexSource.value,
      latex_answer: latexAnswer.value,
      question_tags: tagList.value,
      mistake_reason: latexSolution.value,
    });

    router.replace('/tabs/errors');
  } catch {
    window.alert('保存失败，请重试。');
  }
}

function mapQuestionTypeLabel(type: string) {
  switch (type) {
    case 'choice':
      return '选择题';
    case 'fill_blank':
      return '填空题';
    case 'essay':
      return '大题';
    default:
      return '未知题型';
  }
}
</script>

<style scoped>
.page-wrap {
  max-width: 960px;
  margin: 0 auto;
  display: grid;
  gap: 14px;
}

.page-header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
}

.page-header p {
  margin: 4px 0 0;
  color: #64748b;
  font-size: 14px;
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
