<template>
  <ion-page>
    <ion-header>
      <ion-toolbar>
        <ion-title>我的</ion-title>
      </ion-toolbar>
    </ion-header>
    <ion-content class="ion-padding">
      <ion-card>
        <ion-card-header>
          <ion-card-subtitle>账号</ion-card-subtitle>
          <ion-card-title>{{ authStore.username || '未登录' }}</ion-card-title>
        </ion-card-header>
        <ion-card-content>
          待同步：{{ recordStore.pendingCount }} 条
        </ion-card-content>
      </ion-card>

      <ion-list inset>
        <ion-item button detail @click="toSync">
          <ion-label>手动同步</ion-label>
        </ion-item>
        <ion-item button detail @click="toPdf">
          <ion-label>导出错题本 PDF</ion-label>
        </ion-item>
      </ion-list>

      <ion-button expand="block" color="danger" fill="outline" class="ion-margin-top" @click="logout">退出登录</ion-button>
    </ion-content>
  </ion-page>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import { IonButton, IonCard, IonCardContent, IonCardHeader, IonCardSubtitle, IonCardTitle, IonContent, IonHeader, IonItem, IonLabel, IonList, IonPage, IonTitle, IonToolbar } from '@ionic/vue';
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
