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

      <ion-text v-else color="medium">暂无错题，点击底部中间“新增”开始录入。</ion-text>
    </ion-content>
  </ion-page>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia';
import { useRouter } from 'vue-router';
import { IonBadge, IonCard, IonCardContent, IonCardHeader, IonCardSubtitle, IonCardTitle, IonContent, IonHeader, IonItem, IonLabel, IonList, IonPage, IonSearchbar, IonText, IonTitle, IonToolbar } from '@ionic/vue';
import { useRecordStore } from '@/stores/record';

const router = useRouter();
const recordStore = useRecordStore();
const { records } = storeToRefs(recordStore);

function toDetail(id: number) {
  router.push(`/records/${id}`);
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
</style>
