<template>
  <ion-page>
    <ion-header>
      <ion-toolbar>
        <ion-title>录入错题</ion-title>
      </ion-toolbar>
    </ion-header>
    <ion-content class="ion-padding">
      <ion-item>
        <ion-label position="stacked">标题</ion-label>
        <ion-input v-model="title" />
      </ion-item>
      <ion-item>
        <ion-label position="stacked">学科</ion-label>
        <ion-input v-model="subject" placeholder="例如：math" />
      </ion-item>
      <ion-item>
        <ion-label position="stacked">LaTeX</ion-label>
        <ion-textarea v-model="latexSource" :rows="8" placeholder="输入大模型视觉识别得到的 LaTeX" />
      </ion-item>
      <ion-item>
        <ion-label position="stacked">错因</ion-label>
        <ion-textarea v-model="mistakeReason" :rows="3" />
      </ion-item>
      <ion-button class="ion-margin-top" expand="block" :disabled="!latexSource.trim()" @click="save">保存到本地</ion-button>
    </ion-content>
  </ion-page>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { IonButton, IonContent, IonHeader, IonInput, IonItem, IonLabel, IonPage, IonTextarea, IonTitle, IonToolbar } from '@ionic/vue';
import { useRecordStore } from '@/stores/record';

const router = useRouter();
const recordStore = useRecordStore();

const title = ref('');
const subject = ref('math');
const latexSource = ref('');
const mistakeReason = ref('');

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
    latexVersion: 1,
    latexRenderStatus: 'pending',
    mistakeReason: mistakeReason.value,
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
