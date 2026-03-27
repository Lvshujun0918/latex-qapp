<template>
  <ion-page>
    <ion-header>
      <ion-toolbar>
        <ion-title>错题</ion-title>
      </ion-toolbar>
    </ion-header>
    <ion-content class="ion-padding">
      <ion-card class="hero-card">
        <ion-card-header>
          <ion-card-subtitle>AI 错题本</ion-card-subtitle>
          <ion-card-title>今天继续攻克薄弱点</ion-card-title>
        </ion-card-header>
        <ion-card-content>
          当前共 {{ records.length }} 道题，优先复习高频错因与低掌握度题目。
        </ion-card-content>
      </ion-card>

      <ion-searchbar placeholder="搜索 LaTeX 题目" />

      <ion-list v-if="records.length" class="record-list">
        <ion-item v-for="record in records" :key="record.id" button detail @click="toDetail(record.id)">
          <ion-label>
            <h2>{{ record.title || '未命名题目' }}</h2>
            <p>{{ record.subject }} · 难度 {{ record.difficulty }} · {{ record.syncStatus }}</p>
          </ion-label>
          <ion-badge color="primary">LaTeX</ion-badge>
        </ion-item>
      </ion-list>

      <ion-card v-else class="empty-card">
        <ion-card-content class="empty-content">
          <div class="empty-icon-wrap">
            <ion-icon :icon="sparkles" class="empty-icon" />
          </div>
          <h3>暂无错题</h3>
          <p>点击底部中间新增，拍照后由大模型自动生成 LaTeX 题目、答案与标签。</p>
          <ion-button @click="createFromCamera">
            <ion-icon slot="start" :icon="camera" />
            立即拍照录题
          </ion-button>
        </ion-card-content>
      </ion-card>
    </ion-content>
  </ion-page>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia';
import { useRouter } from 'vue-router';
import { IonBadge, IonButton, IonCard, IonCardContent, IonCardHeader, IonCardSubtitle, IonCardTitle, IonContent, IonHeader, IonIcon, IonItem, IonLabel, IonList, IonPage, IonSearchbar, IonTitle, IonToolbar, toastController } from '@ionic/vue';
import { camera, sparkles } from 'ionicons/icons';
import { useRecordStore } from '@/stores/record';
import { capturePhotoAsBase64, generateLatexDraftByVision, saveVisionDraftToStorage } from '@/services/ai';

const router = useRouter();
const recordStore = useRecordStore();
const { records } = storeToRefs(recordStore);

function toDetail(id: number) {
  router.push(`/records/${id}`);
}

async function createFromCamera() {
  try {
    const imageBase64 = await capturePhotoAsBase64();
    const draft = await generateLatexDraftByVision(imageBase64);
    saveVisionDraftToStorage(draft);
    router.push('/records/new');
  } catch {
    const toast = await toastController.create({
      message: '拍照或识别失败，请重试。',
      duration: 1800,
      color: 'warning',
      position: 'top',
    });
    await toast.present();
  }
}
</script>

<style scoped>
.hero-card {
  margin-top: 6px;
}

.record-list {
  margin-top: 8px;
}

h2 {
  margin: 0;
  font-size: 16px;
  font-weight: 650;
}

p {
  margin-top: 6px;
  color: rgba(20, 32, 51, 0.72);
}

.empty-card {
  margin-top: 14px;
}

.empty-content {
  display: flex;
  align-items: center;
  flex-direction: column;
  text-align: center;
  gap: 10px;
  padding: 8px 4px;
}

.empty-icon-wrap {
  width: 64px;
  height: 64px;
  display: grid;
  place-items: center;
  border-radius: 20px;
  background: linear-gradient(145deg, rgba(31, 122, 255, 0.2), rgba(31, 122, 255, 0.06));
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.46);
}

.empty-icon {
  font-size: 34px;
  color: var(--ion-color-primary);
}

h3 {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
}
</style>
