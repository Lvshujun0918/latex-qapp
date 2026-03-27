<template>
  <ion-page>
    <ion-tabs>
      <ion-router-outlet />

      <ion-tab-bar slot="bottom" class="main-tabbar">
        <ion-tab-button tab="errors" href="/tabs/errors">
          <ion-icon aria-hidden="true" :icon="book" />
          <ion-label>错题</ion-label>
        </ion-tab-button>

        <ion-tab-button tab="review" href="/tabs/review">
          <ion-icon aria-hidden="true" :icon="school" />
          <ion-label>复习</ion-label>
        </ion-tab-button>

        <ion-tab-button tab="create" href="/records/new" class="create-tab" @click.prevent="handleCreateClick">
          <ion-icon aria-hidden="true" :icon="addCircle" />
          <ion-label>新增</ion-label>
        </ion-tab-button>

        <ion-tab-button tab="stats" href="/tabs/stats">
          <ion-icon aria-hidden="true" :icon="barChart" />
          <ion-label>统计</ion-label>
        </ion-tab-button>

        <ion-tab-button tab="profile" href="/tabs/profile">
          <ion-icon aria-hidden="true" :icon="person" />
          <ion-label>我的</ion-label>
        </ion-tab-button>
      </ion-tab-bar>
    </ion-tabs>
  </ion-page>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import { IonTabBar, IonTabButton, IonTabs, IonLabel, IonIcon, IonPage, IonRouterOutlet, toastController } from '@ionic/vue';
import { addCircle, barChart, book, person, school } from 'ionicons/icons';
import { capturePhotoAsBase64, generateLatexDraftByVision, saveVisionDraftToStorage } from '@/services/ai';

const router = useRouter();

async function handleCreateClick() {
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
.main-tabbar {
  --background: rgba(255, 255, 255, 0.94);
  backdrop-filter: saturate(180%) blur(14px);
  border-top: 1px solid rgba(18, 34, 56, 0.08);
  padding-top: 4px;
}

.create-tab {
  --color-selected: var(--ion-color-primary);
  transform: translateY(-10px);
}

.create-tab ion-icon {
  font-size: 2.2rem;
  padding: 4px;
  border-radius: 999px;
  color: var(--ion-color-primary);
  background: linear-gradient(180deg, rgba(31, 122, 255, 0.2), rgba(31, 122, 255, 0.08));
  box-shadow: 0 10px 22px rgba(31, 122, 255, 0.25);
}

.create-tab ion-label {
  font-weight: 700;
}
</style>
