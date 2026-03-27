<template>
  <ion-page>
    <ion-tabs>
      <ion-router-outlet />

      <ion-fab slot="fixed" vertical="bottom" horizontal="end">
        <ion-fab-button @click="toEditor">
          <ion-icon aria-hidden="true" :icon="add" />
        </ion-fab-button>
      </ion-fab>

      <ion-tab-bar slot="bottom">
        <ion-tab-button tab="errors" href="/tabs/errors">
          <ion-icon aria-hidden="true" :icon="book" />
          <ion-label>错题</ion-label>
        </ion-tab-button>

        <ion-tab-button tab="review" href="/tabs/review">
          <ion-icon aria-hidden="true" :icon="school" />
          <ion-label>复习</ion-label>
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
import { useRouter } from 'vue-router';
import { IonFab, IonFabButton, IonTabBar, IonTabButton, IonTabs, IonLabel, IonIcon, IonPage, IonRouterOutlet } from '@ionic/vue';
import { add, barChart, book, person, school } from 'ionicons/icons';

const router = useRouter();

function toEditor() {
  router.push('/records/new');
}
</script>
