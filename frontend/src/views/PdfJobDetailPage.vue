<template>
  <ion-page>
    <ion-header>
      <ion-toolbar>
        <ion-title>PDF 任务</ion-title>
      </ion-toolbar>
    </ion-header>
    <ion-content class="ion-padding">
      <ion-card>
        <ion-card-header>
          <ion-card-title>任务号：{{ route.params.jobId }}</ion-card-title>
        </ion-card-header>
        <ion-card-content>
          <p>状态：queued</p>
          <p>进度：0%</p>
          <p>后续接入后端后展示实时状态。</p>
        </ion-card-content>
      </ion-card>
    </ion-content>
  </ion-page>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router';
import { IonCard, IonCardContent, IonCardHeader, IonCardTitle, IonContent, IonHeader, IonPage, IonTitle, IonToolbar } from '@ionic/vue';

const route = useRoute();
</script>
