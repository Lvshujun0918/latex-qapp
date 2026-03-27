<template>
  <section class="page-wrap">
    <header class="page-header">
      <h1>我的</h1>
      <p>管理账号、同步任务与 PDF 导出。</p>
    </header>

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
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

const authStore = useAuthStore();
const recordStore = useRecordStore();
const router = useRouter();
const syncing = ref(false);

onMounted(() => {
  recordStore.reload();
});

async function toSync() {
  if (syncing.value) {
    return;
  }

  syncing.value = true;
  try {
    const result = await syncNow();
    window.alert(result.message);
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
  max-width: 960px;
  margin: 0 auto;
  display: grid;
  gap: 14px;
}

.page-header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
}

.page-header p {
  margin: 4px 0 0;
  color: #64748b;
  font-size: 14px;
}

.actions {
  display: grid;
  gap: 10px;
}
</style>
