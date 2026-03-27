<template>
  <ion-page>
    <ion-header>
      <ion-toolbar>
        <ion-title>录入错题</ion-title>
      </ion-toolbar>
    </ion-header>
    <ion-content class="ion-padding">
      <ion-card v-if="hasDraft">
        <ion-card-header>
          <ion-card-subtitle>{{ subject || 'unknown' }} · {{ questionTypeLabel }}</ion-card-subtitle>
          <ion-card-title>{{ title || '识别题目' }}</ion-card-title>
        </ion-card-header>
        <ion-card-content>
          <h4>题目（LaTeX）</h4>
          <latex-view :source="latexSource" class="latex-panel" />

          <h4>题目标签</h4>
          <div class="tag-wrap">
            <ion-chip v-for="tag in tagList" :key="tag" color="primary" outline>{{ tag }}</ion-chip>
            <ion-chip v-if="!tagList.length" color="medium" outline>未识别到标签</ion-chip>
          </div>

          <div class="solve-header">
            <h4>答案与解答</h4>
            <ion-button size="small" fill="outline" :disabled="isSolving" @click="generateSolve">
              {{ isSolving ? solvingStage || '生成中...' : '生成解答' }}
            </ion-button>
          </div>

          <ion-item lines="none" class="form-item">
            <ion-label position="stacked">最终答案（可手动修改）</ion-label>
            <ion-textarea v-model="latexAnswer" auto-grow placeholder="例如：x=2" />
          </ion-item>

          <ion-item lines="none" class="form-item">
            <ion-label position="stacked">分步解答（可手动填写）</ion-label>
            <ion-textarea v-model="latexSolution" auto-grow placeholder="点击“生成解答”或手动输入" />
          </ion-item>

          <h4>解答预览（LaTeX）</h4>
          <latex-view :source="latexSolution" class="latex-panel" />
        </ion-card-content>
      </ion-card>

      <ion-card v-else>
        <ion-card-content class="empty-tip">
          未找到拍照识别结果，请从底部中间“新增”按钮发起拍照。
        </ion-card-content>
      </ion-card>
      <ion-button class="ion-margin-top" expand="block" :disabled="!hasDraft || recordStore.loading" @click="save">保存错题</ion-button>
    </ion-content>
  </ion-page>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { IonButton, IonCard, IonCardContent, IonCardHeader, IonCardSubtitle, IonCardTitle, IonChip, IonContent, IonHeader, IonItem, IonLabel, IonPage, IonTextarea, IonTitle, IonToolbar, toastController } from '@ionic/vue';
import { useRecordStore } from '@/stores/record';
import { clearVisionDraftStorage, generateSolutionByLatexStream, loadVisionDraftFromStorage } from '@/services/ai';
import LatexView from '@/components/LatexView.vue';

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
    const solved = await generateSolutionByLatexStream({
      latexQuestion: latexSource.value,
      questionType: questionType.value,
      subject: subject.value,
    }, (evt) => {
      if (evt.stage === 'solve_start') {
        solvingStage.value = '正在推理...';
      }
      if (evt.stage === 'solve_final') {
        solvingStage.value = '解答完成';
      }
    });

    latexAnswer.value = solved.latexAnswer || latexAnswer.value;
    latexSolution.value = solved.latexSolution || latexSolution.value;
  } catch {
    const toast = await toastController.create({
      message: '解答生成失败，请稍后重试或手动填写。',
      duration: 1800,
      color: 'warning',
      position: 'top',
    });
    await toast.present();
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
    const toast = await toastController.create({
      message: '保存失败，请重试。',
      duration: 1800,
      color: 'danger',
      position: 'top',
    });
    await toast.present();
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
h4 {
  margin: 14px 0 8px;
  font-size: 13px;
  color: rgba(20, 32, 51, 0.72);
}

.latex-panel {
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.52);
  border: 1px solid rgba(255, 255, 255, 0.48);
  padding: 10px 12px;
}

.tag-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.solve-header {
  margin-top: 12px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.solve-header h4 {
  margin: 0;
}

.form-item {
  --inner-padding-start: 0;
  --padding-start: 0;
  --background: transparent;
}

.empty-tip {
  text-align: center;
  color: rgba(20, 32, 51, 0.76);
}
</style>
