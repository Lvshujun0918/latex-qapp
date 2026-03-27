<template>
  <ion-page>
    <ion-header>
      <ion-toolbar>
        <ion-title>错题</ion-title>
        <ion-buttons slot="end">
          <ion-button @click="toEditor">新增</ion-button>
        </ion-buttons>
      </ion-toolbar>
    </ion-header>
    <ion-content class="ion-padding">
      <ion-searchbar placeholder="搜索 LaTeX 题目" />
      <ion-list v-if="records.length">
        <ion-item v-for="record in records" :key="record.id" button @click="toDetail(record.id)">
          <ion-label>
            <h2>{{ record.title || '未命名题目' }}</h2>
            <p>{{ record.subject }} · 难度 {{ record.difficulty }} · {{ record.syncStatus }}</p>
          </ion-label>
        </ion-item>
      </ion-list>
      <ion-text v-else color="medium">暂无错题，点击右上角新增。</ion-text>
    </ion-content>
  </ion-page>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia';
import { useRouter } from 'vue-router';
import { IonButton, IonButtons, IonContent, IonHeader, IonItem, IonLabel, IonList, IonPage, IonSearchbar, IonText, IonTitle, IonToolbar } from '@ionic/vue';
import { useRecordStore } from '@/stores/record';

const router = useRouter();
const recordStore = useRecordStore();
const { records } = storeToRefs(recordStore);

function toEditor() {
  router.push('/records/new');
}

function toDetail(id: number) {
  router.push(`/records/${id}`);
}
</script>
