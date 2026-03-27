<template>
  <RouterView />
</template>

<script setup lang="ts">
import { onMounted } from 'vue';
import { RouterView } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { useTheme } from '@/composables/useTheme';

const authStore = useAuthStore();
const { initTheme } = useTheme();

onMounted(() => {
  initTheme();
  authStore.fetchMe();
});
</script>
