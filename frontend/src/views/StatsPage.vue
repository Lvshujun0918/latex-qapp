<template>
  <ion-page>
    <ion-header>
      <ion-toolbar>
        <ion-title>统计</ion-title>
      </ion-toolbar>
    </ion-header>
    <ion-content class="ion-padding">
      <ion-card>
        <ion-card-header>
          <ion-card-title>学习统计</ion-card-title>
        </ion-card-header>
        <ion-card-content>这里将展示总错题数、学科分布和掌握度趋势。</ion-card-content>
      </ion-card>
    </ion-content>
  </ion-page>
</template>

<script setup lang="ts">
import { IonCard, IonCardContent, IonCardHeader, IonCardTitle, IonContent, IonHeader, IonPage, IonTitle, IonToolbar } from '@ionic/vue';
</script>
