<template>
  <section class="app-page app-inner-page page-wrap" :class="{ 'is-dark': resolvedTheme === 'dark' }">
    <header class="app-page-header page-header">
      <Button variant="outline" size="icon-sm" class="app-header-back mr-4" @click="goBack" aria-label="返回上一级"><</Button>
      <h1>后端状态</h1>
      <p>查看服务可用性、版本与请求时延。</p>
    </header>

    <Alert v-if="errorMessage" variant="destructive">
      <AlertTitle>获取失败</AlertTitle>
      <AlertDescription>{{ errorMessage }}</AlertDescription>
    </Alert>

    <div class="status-item">
      <span>服务状态</span>
      <strong :class="runtime?.status === 'ok' ? 'ok' : 'bad'">{{ runtime?.status || 'unknown' }}</strong>
    </div>
    <div class="status-item">
      <span>后端版本</span>
      <strong>{{ runtime?.version || '-' }}</strong>
    </div>
    <div class="status-item">
      <span>请求时延</span>
      <strong>{{ latencyText }}</strong>
    </div>
    <div class="status-item">
      <span>Go 版本</span>
      <strong>{{ runtime?.go_version || '-' }}</strong>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useTheme } from '@/composables/useTheme';
import { fetchBackendRuntime, type BackendRuntimeInfo } from '@/services/system';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

const router = useRouter();
const { resolvedTheme } = useTheme();

const loading = ref(false);
const errorMessage = ref('');
const runtime = ref<BackendRuntimeInfo | null>(null);
const latencyMs = ref<number | null>(null);
const lastRefreshedAt = ref('');

const latencyText = computed(() => (latencyMs.value === null ? '-' : `${latencyMs.value} ms`));
const uptimeText = computed(() => {
  const total = Number(runtime.value?.uptime_seconds || 0);
  if (!total) {
    return '-';
  }
  const days = Math.floor(total / 86400);
  const hours = Math.floor((total % 86400) / 3600);
  const minutes = Math.floor((total % 3600) / 60);
  const seconds = total % 60;
  const dayPart = days > 0 ? `${days}天 ` : '';
  return `${dayPart}${hours}小时 ${minutes}分 ${seconds}秒`;
});

onMounted(() => {
  loadRuntime();
});

async function loadRuntime() {
  loading.value = true;
  try {
    errorMessage.value = '';
    const result = await fetchBackendRuntime();
    runtime.value = result.data;
    latencyMs.value = result.latencyMs;
    lastRefreshedAt.value = new Date().toISOString();
  } catch (error: any) {
    errorMessage.value = error?.response?.data?.error || error?.message || '获取后端状态失败';
  } finally {
    loading.value = false;
  }
}

function formatDate(raw?: string) {
  const text = String(raw || '').trim();
  if (!text) {
    return '-';
  }
  const t = Date.parse(text);
  if (!Number.isFinite(t)) {
    return text;
  }
  return new Date(t).toLocaleString();
}

function goBack() {
  if (window.history.length > 1) {
    router.back();
    return;
  }
  router.replace('/tabs/profile');
}
</script>

<style scoped>
.page-wrap {
  gap: 12px;
}

.status-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.status-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

.status-item {
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 12px;
  background: rgba(248, 250, 252, 0.78);
  padding: 10px 12px;
  display: grid;
  gap: 2px;
}

.status-item span {
  font-size: 12px;
  color: #64748b;
}

.status-item strong {
  color: #0f172a;
  font-size: 15px;
}

.status-item strong.ok {
  color: #047857;
}

.status-item strong.bad {
  color: #b91c1c;
}

.runtime-list {
  display: grid;
  gap: 6px;
}

.runtime-list p {
  margin: 0;
  font-size: 13px;
  color: #334155;
}

.runtime-list span {
  color: #64748b;
}

.is-dark .status-item {
  background: rgba(30, 41, 59, 0.9);
  border-color: rgba(148, 163, 184, 0.28);
}

.is-dark .status-item span,
.is-dark .runtime-list span {
  color: #94a3b8;
}

.is-dark .status-item strong,
.is-dark .runtime-list p {
  color: #f8fafc;
}

@media (max-width: 640px) {
  .status-grid {
    grid-template-columns: 1fr;
  }
}
</style>
