<template>
  <section class="app-page page-wrap">
    <header class="app-page-header page-header">
      <h1>我的</h1>
      <p>管理账号、同步任务与 PDF 导出。</p>
    </header>

    <Alert v-if="message" :variant="messageVariant">
      <AlertTitle>{{ messageTitle }}</AlertTitle>
      <AlertDescription>{{ message }}</AlertDescription>
    </Alert>

    <Card>
      <CardHeader>
        <CardDescription>账号</CardDescription>
        <CardTitle>{{ authStore.displayName || authStore.username || '未登录' }}</CardTitle>
      </CardHeader>
      <CardContent>待同步：{{ recordStore.pendingCount }} 条</CardContent>
    </Card>

    <div class="actions">
      <Button variant="outline" :disabled="syncing" @click="toSync">
        {{ syncing ? '同步中...' : '手动同步' }}
      </Button>
      <Button variant="outline" @click="toPdf">导出错题本 PDF</Button>
    </div>

    <Button variant="destructive" @click="logout">退出登录</Button>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { useRecordStore } from '@/stores/record';
import { syncNow } from '@/services/sync';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

const authStore = useAuthStore();
const recordStore = useRecordStore();
const router = useRouter();
const syncing = ref(false);
const message = ref('');
const messageTitle = ref('');
const messageVariant = ref<'default' | 'destructive'>('default');

onMounted(() => {
  recordStore.reload();
});

async function toSync() {
  if (syncing.value) {
    return;
  }

  syncing.value = true;
  try {
    message.value = '';
    const result = await syncNow();
    messageTitle.value = result.ok ? '同步已触发' : '同步失败';
    messageVariant.value = result.ok ? 'default' : 'destructive';
    message.value = result.message;
    await recordStore.reload();
  } finally {
    syncing.value = false;
  }
}

function toPdf() {
  router.push('/pdf/export');
}

function logout() {
  authStore.logout();
  router.replace('/login');
}
</script>

<style scoped>
.page-wrap {
  gap: 12px;
}

.actions {
  display: grid;
  gap: 10px;
}
</style>
