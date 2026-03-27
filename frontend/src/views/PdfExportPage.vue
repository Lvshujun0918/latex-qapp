<template>
  <ion-page>
    <ion-header>
      <ion-toolbar>
        <ion-title>导出错题本 PDF</ion-title>
      </ion-toolbar>
    </ion-header>
    <ion-content class="ion-padding">
      <ion-item>
        <ion-label position="stacked">学科</ion-label>
        <ion-input v-model="subject" />
      </ion-item>
      <ion-item>
        <ion-label position="stacked">难度</ion-label>
        <ion-input v-model.number="difficulty" type="number" />
      </ion-item>
      <ion-button class="ion-margin-top" expand="block" @click="submit">提交导出任务</ion-button>
    </ion-content>
  </ion-page>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { IonButton, IonContent, IonHeader, IonInput, IonItem, IonLabel, IonPage, IonTitle, IonToolbar } from '@ionic/vue';

const router = useRouter();
const subject = ref('math');
const difficulty = ref(3);

function submit() {
  router.push('/pdf/jobs/mock-job-1');
}
</script>
