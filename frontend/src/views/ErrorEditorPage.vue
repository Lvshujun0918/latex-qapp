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
          <ion-card-subtitle>{{ subject || 'unknown' }}</ion-card-subtitle>
          <ion-card-title>{{ title || '识别题目' }}</ion-card-title>
        </ion-card-header>
        <ion-card-content>
          <h4>题目（LaTeX）</h4>
          <pre class="latex-block">{{ latexSource }}</pre>

          <h4>答案（LaTeX）</h4>
          <pre class="latex-block">{{ latexAnswer || '暂无' }}</pre>

          <h4>题目标签</h4>
          <div class="tag-wrap">
            <ion-chip v-for="tag in tagList" :key="tag" color="primary" outline>{{ tag }}</ion-chip>
            <ion-chip v-if="!tagList.length" color="medium" outline>未识别到标签</ion-chip>
          </div>
        </ion-card-content>
      </ion-card>

      <ion-card v-else>
        <ion-card-content class="empty-tip">
          未找到拍照识别结果，请从底部中间“新增”按钮发起拍照。
        </ion-card-content>
      </ion-card>

      <ion-button class="ion-margin-top" expand="block" :disabled="!hasDraft" @click="save">保存到本地</ion-button>
    </ion-content>
  </ion-page>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { IonButton, IonCard, IonCardContent, IonCardHeader, IonCardSubtitle, IonCardTitle, IonChip, IonContent, IonHeader, IonPage, IonTitle, IonToolbar } from '@ionic/vue';
import { useRecordStore } from '@/stores/record';
import { clearVisionDraftStorage, loadVisionDraftFromStorage } from '@/services/ai';

const router = useRouter();
const recordStore = useRecordStore();

const title = ref('');
const subject = ref('math');
const latexSource = ref('');
const latexAnswer = ref('');
const questionTags = ref<string[]>([]);

const hasDraft = computed(() => latexSource.value.trim().length > 0);
const tagList = computed(() => questionTags.value.filter((item) => item.trim().length > 0));

onMounted(() => {
  const draft = loadVisionDraftFromStorage();
  if (!draft) {
    return;
  }

  title.value = draft.title ?? title.value;
  subject.value = draft.subject ?? subject.value;
  latexSource.value = draft.latexQuestion;
  latexAnswer.value = draft.latexAnswer;
  questionTags.value = draft.tags;
  clearVisionDraftStorage();
});

function save() {
  const now = Date.now();
  recordStore.save({
    id: now,
    userId: 1,
    subject: subject.value,
    questionType: 'unknown',
    difficulty: 3,
    title: title.value,
    latexSource: latexSource.value,
    latexAnswer: latexAnswer.value,
    questionTags: tagList.value,
    latexVersion: 1,
    latexRenderStatus: 'pending',
    mistakeReason: '',
    masteryLevel: 0,
    reviewCount: 0,
    syncStatus: 'pending',
    localVersion: now,
    serverVersion: 0,
    createdAt: new Date(now).toISOString(),
    updatedAt: new Date(now).toISOString(),
  });

  router.replace('/tabs/errors');
}
</script>

<style scoped>
h4 {
  margin: 14px 0 8px;
  font-size: 13px;
  color: rgba(20, 32, 51, 0.72);
}

.latex-block {
  margin: 0;
  white-space: pre-wrap;
  font-size: 13px;
  line-height: 1.55;
  padding: 10px 12px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.52);
  border: 1px solid rgba(255, 255, 255, 0.48);
}

.tag-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.empty-tip {
  text-align: center;
  color: rgba(20, 32, 51, 0.76);
}
</style>
