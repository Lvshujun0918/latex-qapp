<template>
  <ion-page>
    <ion-header>
      <ion-toolbar>
        <ion-title>复习</ion-title>
      </ion-toolbar>
    </ion-header>
    <ion-content class="ion-padding">
      <ion-card>
        <ion-card-header>
          <ion-card-title>今日推荐</ion-card-title>
        </ion-card-header>
        <ion-card-content>根据错题频次和掌握度，系统将推荐重点复习题目。</ion-card-content>
      </ion-card>
      <ion-card>
        <ion-card-header>
          <ion-card-title>高频错因</ion-card-title>
        </ion-card-header>
        <ion-card-content>后续联调后展示来自 AI 的易错点聚合。</ion-card-content>
      </ion-card>
    </ion-content>
  </ion-page>
</template>

<script setup lang="ts">
import { IonCard, IonCardContent, IonCardHeader, IonCardTitle, IonContent, IonHeader, IonPage, IonTitle, IonToolbar } from '@ionic/vue';
</script>
