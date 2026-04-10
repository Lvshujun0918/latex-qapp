<template>
  <RouterView />
</template>

<script setup lang="ts">
import { onMounted, watch } from 'vue';
import { RouterView } from 'vue-router';
import { StatusBar, Style } from '@capacitor/status-bar';
import { useAuthStore } from '@/stores/auth';
import { useTheme } from '@/composables/useTheme';

const authStore = useAuthStore();
const { initTheme, resolvedTheme } = useTheme();

async function syncStatusBar(theme: 'light' | 'dark') {
  if (typeof window === 'undefined') {
    return;
  }

  try {
    await StatusBar.setOverlaysWebView({ overlay: false });
    await StatusBar.setStyle({ style: theme === 'dark' ? Style.Dark : Style.Light });
  } catch {
    // Ignore runtime differences on platforms where the status bar plugin is unavailable.
  }
}

onMounted(() => {
  initTheme();
  authStore.fetchMe();
});

watch(
  resolvedTheme,
  (theme) => {
    syncStatusBar(theme);
  },
  { immediate: true },
);
</script>
