<template>
  <ion-page>
    <ion-tabs>
      <ion-router-outlet />

      <ion-tab-bar slot="bottom" class="main-tabbar">
        <ion-tab-button tab="errors" href="/tabs/errors">
          <ion-icon aria-hidden="true" :icon="book" />
          <ion-label>错题</ion-label>
        </ion-tab-button>

        <ion-tab-button tab="review" href="/tabs/review">
          <ion-icon aria-hidden="true" :icon="school" />
          <ion-label>复习</ion-label>
        </ion-tab-button>

        <ion-tab-button tab="create" href="/records/new" class="create-tab">
          <ion-icon aria-hidden="true" :icon="addCircle" />
          <ion-label>新增</ion-label>
        </ion-tab-button>

        <ion-tab-button tab="stats" href="/tabs/stats">
          <ion-icon aria-hidden="true" :icon="barChart" />
          <ion-label>统计</ion-label>
        </ion-tab-button>

        <ion-tab-button tab="profile" href="/tabs/profile">
          <ion-icon aria-hidden="true" :icon="person" />
          <ion-label>我的</ion-label>
        </ion-tab-button>
      </ion-tab-bar>
    </ion-tabs>
  </ion-page>
</template>

<script setup lang="ts">
import { IonTabBar, IonTabButton, IonTabs, IonLabel, IonIcon, IonPage, IonRouterOutlet } from '@ionic/vue';
import { addCircle, barChart, book, person, school } from 'ionicons/icons';
</script>

<style scoped>
.main-tabbar {
  --background: rgba(255, 255, 255, 0.94);
  backdrop-filter: saturate(180%) blur(14px);
  border-top: 1px solid rgba(18, 34, 56, 0.08);
}

.create-tab {
  --color-selected: var(--ion-color-primary);
}

.create-tab ion-icon {
  font-size: 1.8rem;
}
</style>
