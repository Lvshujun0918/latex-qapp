<template>
  <ion-page>
    <ion-header>
      <ion-toolbar>
        <ion-title>我的</ion-title>
      </ion-toolbar>
    </ion-header>
    <ion-content class="ion-padding">
      <ion-item>
        <ion-label>
          <h2>{{ authStore.username || '未登录' }}</h2>
          <p>待同步：{{ recordStore.pendingCount }} 条</p>
        </ion-label>
      </ion-item>
      <ion-button expand="block" class="ion-margin-top" @click="toSync">手动同步</ion-button>
      <ion-button expand="block" fill="outline" class="ion-margin-top" @click="toPdf">导出错题本 PDF</ion-button>
      <ion-button expand="block" color="danger" fill="clear" class="ion-margin-top" @click="logout">退出登录</ion-button>
    </ion-content>
  </ion-page>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import { IonButton, IonContent, IonHeader, IonItem, IonLabel, IonPage, IonTitle, IonToolbar } from '@ionic/vue';
import { useAuthStore } from '@/stores/auth';
import { useRecordStore } from '@/stores/record';
import { syncNow } from '@/services/sync';

const authStore = useAuthStore();
const recordStore = useRecordStore();
const router = useRouter();

async function toSync() {
  await syncNow();
}

function toPdf() {
  router.push('/pdf/export');
}

function logout() {
  authStore.logout();
  router.replace('/login');
}
</script>
