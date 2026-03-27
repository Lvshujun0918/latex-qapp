<template>
  <ion-page>
    <ion-header>
      <ion-toolbar>
        <ion-title>错题详情</ion-title>
      </ion-toolbar>
    </ion-header>
    <ion-content class="ion-padding" v-if="record">
      <ion-card>
        <ion-card-header>
          <ion-card-title>{{ record.title || '未命名题目' }}</ion-card-title>
        </ion-card-header>
        <ion-card-content>
          <p>学科：{{ record.subject }}</p>
          <p>同步状态：{{ record.syncStatus }}</p>
          <p class="latex-block">{{ record.latexSource }}</p>
        </ion-card-content>
      </ion-card>
      <ion-button expand="block" @click="toAnalysis">AI 解析</ion-button>
    </ion-content>
  </ion-page>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { IonButton, IonCard, IonCardContent, IonCardHeader, IonCardTitle, IonContent, IonHeader, IonPage, IonTitle, IonToolbar } from '@ionic/vue';
import { useRecordStore } from '@/stores/record';

const route = useRoute();
const router = useRouter();
const recordStore = useRecordStore();

const record = computed(() => recordStore.records.find((r) => r.id === Number(route.params.id)));

function toAnalysis() {
  router.push(`/records/${route.params.id}/analysis`);
}
</script>

<style scoped>
.latex-block {
  white-space: pre-wrap;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
}
</style>
