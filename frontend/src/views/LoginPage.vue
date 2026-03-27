<template>
  <ion-page>
    <ion-header>
      <ion-toolbar>
        <ion-title>登录</ion-title>
      </ion-toolbar>
    </ion-header>
    <ion-content class="ion-padding">
      <ion-item>
        <ion-label position="stacked">用户名</ion-label>
        <ion-input v-model="username" autocomplete="username" />
      </ion-item>
      <ion-item>
        <ion-label position="stacked">密码</ion-label>
        <ion-input v-model="password" type="password" autocomplete="current-password" />
      </ion-item>
      <ion-button class="ion-margin-top" expand="block" :disabled="!canSubmit" @click="submit">登录</ion-button>
    </ion-content>
  </ion-page>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';
import { IonButton, IonContent, IonHeader, IonInput, IonItem, IonLabel, IonPage, IonTitle, IonToolbar } from '@ionic/vue';
import { useAuthStore } from '@/stores/auth';

const authStore = useAuthStore();
const router = useRouter();
const username = ref('');
const password = ref('');

const canSubmit = computed(() => username.value.trim().length > 2 && password.value.length > 5);

async function submit() {
  await authStore.login({ username: username.value.trim(), password: password.value });
  router.replace('/tabs/errors');
}
</script>
