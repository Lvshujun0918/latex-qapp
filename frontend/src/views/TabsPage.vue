<template>
  <ion-page>
    <ion-tabs>
      <ion-router-outlet />

      <ion-loading
        :is-open="isGenerating"
        :message="generatingMessage"
        spinner="crescent"
        backdrop-dismiss="false"
      />

      <ion-tab-bar slot="bottom" class="main-tabbar">
        <ion-tab-button tab="errors" href="/tabs/errors">
          <ion-icon aria-hidden="true" :icon="book" />
          <ion-label>错题</ion-label>
        </ion-tab-button>

        <ion-tab-button tab="review" href="/tabs/review">
          <ion-icon aria-hidden="true" :icon="school" />
          <ion-label>复习</ion-label>
        </ion-tab-button>

        <ion-tab-button
          tab="create"
          href="/records/new"
          class="create-tab"
          :class="{ 'is-generating': isGenerating }"
          :disabled="isGenerating"
          @click.prevent="handleCreateClick"
        >
          <ion-icon aria-hidden="true" :icon="addCircle" />
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
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { IonLoading, IonTabBar, IonTabButton, IonTabs, IonLabel, IonIcon, IonPage, IonRouterOutlet, actionSheetController, toastController } from '@ionic/vue';
import { addCircle, barChart, book, person, school } from 'ionicons/icons';
import { generateLatexDraftByVisionStream, pickImageAsBase64, saveVisionDraftToStorage } from '@/services/ai';

const router = useRouter();
const isGenerating = ref(false);
const generatingMessage = ref('正在识别题目与标签...');

async function handleCreateClick() {
  if (isGenerating.value) {
    return;
  }

  try {
    isGenerating.value = true;
    generatingMessage.value = '请选择图片来源...';
    const source = await chooseImageSource();
    const imageBase64 = await pickImageAsBase64(source);
    const draft = await generateLatexDraftByVisionStream(imageBase64, (evt) => {
      switch (evt.stage) {
        case 'classify':
          generatingMessage.value = '正在识别学科与题型...';
          break;
        case 'latex':
          generatingMessage.value = '正在生成题目 LaTeX...';
          break;
        case 'tags':
          generatingMessage.value = '正在生成标签...';
          break;
        case 'final':
          generatingMessage.value = '识别完成，正在进入编辑页...';
          break;
        default:
          break;
      }
    });

    if (!draft.latexQuestion.trim()) {
      throw new Error('识别结果为空');
    }

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
  } finally {
    isGenerating.value = false;
    generatingMessage.value = '正在识别题目与标签...';
  }
}

async function chooseImageSource(): Promise<'camera' | 'album' | 'file'> {
  const sheet = await actionSheetController.create({
    header: '选择图片来源',
    buttons: [
      { text: '拍照', data: { source: 'camera' } },
      { text: '相册', data: { source: 'album' } },
      { text: '文件', data: { source: 'file' } },
      { text: '取消', role: 'cancel' },
    ],
  });
  await sheet.present();
  const result = await sheet.onDidDismiss();
  return (result.data?.source as 'camera' | 'album' | 'file') || 'camera';
}
</script>

<style scoped>
.main-tabbar {
  --background: rgba(255, 255, 255, 0.5);
  backdrop-filter: saturate(185%) blur(28px);
  border-top: 1px solid rgba(255, 255, 255, 0.56);
  box-shadow:
    0 -8px 30px rgba(18, 34, 56, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 0.6);
  min-height: 66px;
  padding-top: 2px;
  padding-bottom: calc(env(safe-area-inset-bottom, 0px) + 2px);
  overflow: visible;
}

.main-tabbar ion-tab-button {
  --color: rgba(20, 32, 51, 0.7);
  --color-selected: var(--ion-color-primary);
  border-radius: 16px;
  margin: 3px 4px 1px;
  transition: transform 0.2s ease, background-color 0.2s ease;
}

.main-tabbar ion-tab-button.tab-selected:not(.create-tab) {
  background: linear-gradient(160deg, rgba(255, 255, 255, 0.7), rgba(219, 236, 255, 0.55));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.65),
    0 8px 20px rgba(31, 122, 255, 0.12);
}

.create-tab {
  --color-selected: #ffffff;
  margin-top: -8px;
}

.create-tab.is-generating ion-icon {
  animation: generating-pulse 1.1s ease-in-out infinite;
}

.create-tab ion-icon {
  font-size: 2.05rem;
  padding: 8px;
  border-radius: 999px;
  color: #ffffff;
  background:
    radial-gradient(circle at 30% 22%, rgba(255, 255, 255, 0.4), rgba(255, 255, 255, 0) 45%),
    linear-gradient(160deg, #5aa6ff 0%, #1f7aff 62%, #1669df 100%);
  box-shadow:
    0 8px 20px rgba(31, 122, 255, 0.35),
    inset 0 1px 0 rgba(255, 255, 255, 0.42);
}

.create-tab ion-label {
  font-weight: 700;
  letter-spacing: 0.2px;
}

.main-tabbar::before {
  content: '';
  position: absolute;
  inset: 0;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.46), rgba(255, 255, 255, 0.08)),
    radial-gradient(circle at 50% -60%, rgba(134, 193, 255, 0.3), rgba(134, 193, 255, 0) 62%);
  pointer-events: none;
  border-top: 1px solid rgba(255, 255, 255, 0.62);
}

@keyframes generating-pulse {
  0% {
    transform: scale(1);
    box-shadow:
      0 8px 20px rgba(31, 122, 255, 0.35),
      inset 0 1px 0 rgba(255, 255, 255, 0.42);
  }

  50% {
    transform: scale(1.08);
    box-shadow:
      0 12px 26px rgba(31, 122, 255, 0.5),
      inset 0 1px 0 rgba(255, 255, 255, 0.5);
  }

  100% {
    transform: scale(1);
    box-shadow:
      0 8px 20px rgba(31, 122, 255, 0.35),
      inset 0 1px 0 rgba(255, 255, 255, 0.42);
  }
}
</style>
