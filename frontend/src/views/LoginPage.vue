<template>
  <ion-page>
    <ion-header>
      <ion-toolbar>
        <ion-title>{{ isRegisterMode ? '注册账号' : '登录' }}</ion-title>
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

      <ion-item v-if="isRegisterMode">
        <ion-label position="stacked">确认密码</ion-label>
        <ion-input v-model="confirmPassword" type="password" autocomplete="new-password" />
      </ion-item>

      <ion-text v-if="errorMessage" color="danger" class="error-tip">{{ errorMessage }}</ion-text>

      <ion-button class="ion-margin-top" expand="block" :disabled="!canSubmit || authStore.loading" @click="submit">
        {{ isRegisterMode ? '注册并登录' : '登录' }}
      </ion-button>
      <ion-button fill="clear" expand="block" @click="toggleMode">
        {{ isRegisterMode ? '已有账号？去登录' : '没有账号？去注册' }}
      </ion-button>
    </ion-content>
  </ion-page>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';
import { IonButton, IonContent, IonHeader, IonInput, IonItem, IonLabel, IonPage, IonText, IonTitle, IonToolbar } from '@ionic/vue';
import { useAuthStore } from '@/stores/auth';

const authStore = useAuthStore();
const router = useRouter();
const username = ref('');
const password = ref('');
const confirmPassword = ref('');
const isRegisterMode = ref(false);
const errorMessage = ref('');

const canSubmit = computed(() => {
  if (username.value.trim().length <= 2 || password.value.length <= 5) {
    return false;
  }
  if (!isRegisterMode.value) {
    return true;
  }
  return confirmPassword.value === password.value;
});

async function submit() {
  errorMessage.value = '';
  const payload = { username: username.value.trim(), password: password.value };

  if (isRegisterMode.value && confirmPassword.value !== password.value) {
    errorMessage.value = '两次输入的密码不一致';
    return;
  }

  try {
    if (isRegisterMode.value) {
      await authStore.register(payload);
    }
    await authStore.login(payload);
    router.replace('/tabs/errors');
  } catch (error: any) {
    errorMessage.value = error?.response?.data?.error || error?.message || '请求失败，请重试';
  }
}

function toggleMode() {
  isRegisterMode.value = !isRegisterMode.value;
  confirmPassword.value = '';
  errorMessage.value = '';
}
</script>

<style scoped>
.error-tip {
  display: block;
  margin-top: 10px;
  padding-left: 4px;
}
</style>
