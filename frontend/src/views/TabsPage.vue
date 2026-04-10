<template>
  <div class="tabs-layout" :class="{ 'is-dark': resolvedTheme === 'dark' }">
    <section v-if="errorMessage" class="app-page error-wrap">
      <Alert variant="destructive">
        <AlertTitle>拍照生成功能异常</AlertTitle>
        <AlertDescription>{{ errorMessage }}</AlertDescription>
      </Alert>
    </section>

    <main class="tabs-content">
      <RouterView />
    </main>

    <nav class="main-tabbar" aria-label="主导航">
      <RouterLink class="tab-link" :class="{ active: isActive('/tabs/errors') }" to="/tabs/errors">
        <BookOpen class="tab-icon" />
        <span>错题</span>
      </RouterLink>

      <RouterLink class="tab-link" :class="{ active: isActive('/tabs/review') }" to="/tabs/review">
        <GraduationCap class="tab-icon" />
        <span>复习</span>
      </RouterLink>

      <button
        class="tab-create"
        :class="{ 'is-generating': isGenerating }"
        :disabled="isGenerating"
        type="button"
        @click="openSourceDialog"
      >
        <PlusCircle class="create-icon" />
      </button>

      <RouterLink class="tab-link" :class="{ active: isActive('/tabs/stats') }" to="/tabs/stats">
        <BarChart3 class="tab-icon" />
        <span>统计</span>
      </RouterLink>

      <RouterLink class="tab-link" :class="{ active: isActive('/tabs/profile') }" to="/tabs/profile">
        <CircleUserRound class="tab-icon" />
        <span>我的</span>
      </RouterLink>
    </nav>

    <div v-if="isGenerating" class="loading-overlay">
      <div class="loading-card">
        <div class="loader" />
        <p>{{ generatingMessage }}</p>
      </div>
    </div>

    <ImageSourceDialog
      :open="sourceDialogOpen"
      @update:open="(val) => (sourceDialogOpen = val)"
      @select="handleCreateClick"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRoute, useRouter, RouterLink, RouterView } from 'vue-router';
import { BarChart3, BookOpen, CircleUserRound, GraduationCap, PlusCircle } from 'lucide-vue-next';
import ImageSourceDialog from '@/components/ImageSourceDialog.vue';
import { useTheme } from '@/composables/useTheme';
import { pickImageAsBase64 } from '@/services/ai';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';

const route = useRoute();
const router = useRouter();
const { resolvedTheme } = useTheme();
const isGenerating = ref(false);
const generatingMessage = ref('正在识别题目与标签...');
const sourceDialogOpen = ref(false);
const errorMessage = ref('');

function isActive(path: string) {
  return route.path.startsWith(path);
}

function openSourceDialog() {
  if (isGenerating.value) {
    return;
  }

  errorMessage.value = '';
  sourceDialogOpen.value = true;
}

async function handleCreateClick(source: 'camera' | 'album' | 'file') {
  if (isGenerating.value) {
    return;
  }

  try {
    errorMessage.value = '';
    isGenerating.value = true;
    generatingMessage.value = '正在准备图片...';

    const imageBase64 = await pickImageAsBase64(source);
    router.push({
      path: '/image/crop',
      query: { data: imageBase64 },
    });
  } catch (error: any) {
    errorMessage.value = error?.message || '拍照或识别失败，请重试。';
  } finally {
    isGenerating.value = false;
    generatingMessage.value = '正在识别题目与标签...';
  }
}
</script>

<style scoped>
.tabs-layout {
  min-height: 100dvh;
  background:
    radial-gradient(circle at 16% -8%, rgba(124, 186, 255, 0.28), transparent 40%),
    radial-gradient(circle at 95% 4%, rgba(89, 219, 171, 0.2), transparent 34%),
    radial-gradient(circle at 50% 100%, rgba(255, 255, 255, 0.7), transparent 48%),
    hsl(var(--background));
}

.tabs-layout.is-dark {
  background:
    radial-gradient(circle at 16% -8%, rgba(56, 189, 248, 0.18), transparent 40%),
    radial-gradient(circle at 95% 4%, rgba(16, 185, 129, 0.16), transparent 34%),
    radial-gradient(circle at 50% 100%, rgba(15, 23, 42, 0.5), transparent 48%),
    hsl(var(--background));
}

.tabs-content {
  padding: 16px 16px calc(88px + env(safe-area-inset-bottom, 0px));
}

.error-wrap {
  position: sticky;
  top: 8px;
  z-index: 30;
  padding-top: 10px;
}

.main-tabbar {
  position: fixed;
  left: 12px;
  right: 12px;
  bottom: calc(10px + env(safe-area-inset-bottom, 0px));
  height: 68px;
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  align-items: center;
  gap: 6px;
  border-radius: 20px;
  border: 1px solid rgba(255, 255, 255, 0.56);
  background: rgba(255, 255, 255, 0.66);
  box-shadow:
    0 -8px 30px rgba(18, 34, 56, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 0.6);
  backdrop-filter: saturate(185%) blur(24px);
  padding: 0 8px;
}

.tabs-layout.is-dark .main-tabbar {
  border: 1px solid rgba(148, 163, 184, 0.26);
  background: rgba(15, 23, 42, 0.72);
  box-shadow:
    0 -8px 26px rgba(2, 6, 23, 0.44),
    inset 0 1px 0 rgba(148, 163, 184, 0.12);
}

.tab-link {
  height: 56px;
  border-radius: 14px;
  text-decoration: none;
  color: hsl(var(--muted-foreground));
  display: inline-flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  font-weight: 600;
  transition: transform 0.2s ease, background-color 0.2s ease, color 0.2s ease;
}

.tab-link.active {
  color: hsl(var(--primary));
  background: linear-gradient(160deg, rgba(255, 255, 255, 0.7), rgba(219, 236, 255, 0.55));
}

.tabs-layout.is-dark .tab-link {
  color: rgba(203, 213, 225, 0.9);
}

.tabs-layout.is-dark .tab-link.active {
  color: #dbeafe;
  background: linear-gradient(160deg, rgba(30, 41, 59, 0.8), rgba(37, 99, 235, 0.22));
}

.tab-icon {
  width: 18px;
  height: 18px;
}

.tab-create {
  border: none;
  background: transparent;
  display: grid;
  place-items: center;
  cursor: pointer;
}

.tab-create:disabled {
  cursor: not-allowed;
}

.create-icon {
  width: 42px;
  height: 42px;
  color: #ffffff;
  border-radius: 999px;
  padding: 8px;
  background:
    radial-gradient(circle at 30% 22%, rgba(255, 255, 255, 0.4), rgba(255, 255, 255, 0) 45%),
    linear-gradient(160deg, #5aa6ff 0%, #1f7aff 62%, #1669df 100%);
  box-shadow:
    0 8px 20px rgba(31, 122, 255, 0.35),
    inset 0 1px 0 rgba(255, 255, 255, 0.42);
}

.tab-create.is-generating .create-icon {
  animation: generating-pulse 1.1s ease-in-out infinite;
}

.loading-overlay {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.32);
  backdrop-filter: blur(2px);
  display: grid;
  place-items: center;
  z-index: 50;
}

.loading-card {
  min-width: 220px;
  background: rgba(255, 255, 255, 0.95);
  border: 1px solid rgba(148, 163, 184, 0.4);
  border-radius: 16px;
  padding: 20px 16px;
  text-align: center;
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.2);
}

.tabs-layout.is-dark .loading-card {
  background: rgba(15, 23, 42, 0.9);
  border: 1px solid rgba(148, 163, 184, 0.32);
  box-shadow: 0 12px 32px rgba(2, 6, 23, 0.5);
}

.loading-card p {
  margin: 10px 0 0;
  font-size: 13px;
  color: #334155;
}

.tabs-layout.is-dark .loading-card p {
  color: #cbd5e1;
}

.loader {
  width: 26px;
  height: 26px;
  border: 3px solid rgba(37, 99, 235, 0.2);
  border-top-color: #2563eb;
  border-radius: 50%;
  margin: 0 auto;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@keyframes generating-pulse {
  0% {
    transform: scale(1);
  }

  50% {
    transform: scale(1.08);
  }

  100% {
    transform: scale(1);
  }
}
</style>
