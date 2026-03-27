<template>
  <ion-page>
    <ion-header>
      <ion-toolbar>
        <ion-title>AI 解析</ion-title>
      </ion-toolbar>
    </ion-header>
    <ion-content class="ion-padding">
      <ion-card>
        <ion-card-header>
          <ion-card-title>解析结果</ion-card-title>
        </ion-card-header>
        <ion-card-content>
          <p>待接入后端接口后展示详细步骤、知识点和复习建议。</p>
        </ion-card-content>
      </ion-card>
    </ion-content>
  </ion-page>
</template>

<script setup lang="ts">
import { IonCard, IonCardContent, IonCardHeader, IonCardTitle, IonContent, IonHeader, IonPage, IonTitle, IonToolbar } from '@ionic/vue';
</script>
