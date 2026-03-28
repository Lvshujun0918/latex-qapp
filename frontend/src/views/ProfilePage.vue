<template>
  <section class="app-page page-wrap" :class="{ 'is-dark': resolvedTheme === 'dark' }">
    <header class="app-page-header page-header">
      <span class="app-kicker">Workspace</span>
      <h1>我的</h1>
      <p>管理账号与界面外观设置。</p>
    </header>

    <Card class="app-page-shell">
      <CardHeader>
        <CardDescription>账号</CardDescription>
        <CardTitle>{{ authStore.displayName || authStore.username || '未登录' }}</CardTitle>
      </CardHeader>
      <CardContent>当前题库：{{ recordStore.records.length }} 条</CardContent>
    </Card>

    <Card class="app-soft-card">
      <CardHeader>
        <CardDescription>主题</CardDescription>
        <CardTitle>界面外观</CardTitle>
      </CardHeader>
      <CardContent class="theme-wrap">
        <div class="theme-grid">
          <Button
            variant="outline"
            class="theme-btn"
            :class="{ active: themeMode === 'light' }"
            @click="setTheme('light')"
          >
            浅色
          </Button>
          <Button
            variant="outline"
            class="theme-btn"
            :class="{ active: themeMode === 'dark' }"
            @click="setTheme('dark')"
          >
            深色
          </Button>
          <Button
            variant="outline"
            class="theme-btn"
            :class="{ active: themeMode === 'system' }"
            @click="setTheme('system')"
          >
            跟随系统
          </Button>
        </div>
        <p class="theme-tip">当前生效：{{ resolvedTheme === 'dark' ? '深色' : '浅色' }}</p>
      </CardContent>
    </Card>

    <Button variant="destructive" @click="logout">退出登录</Button>
  </section>
</template>

<script setup lang="ts">
import { onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { useRecordStore } from '@/stores/record';
import { useTheme } from '@/composables/useTheme';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

const authStore = useAuthStore();
const recordStore = useRecordStore();
const router = useRouter();
const { themeMode, resolvedTheme, setTheme } = useTheme();

onMounted(() => {
  recordStore.reload();
});

function logout() {
  authStore.logout();
  router.replace('/login');
}
</script>

<style scoped>
.page-wrap {
  gap: 12px;
}

.theme-wrap {
  display: grid;
  gap: 10px;
}

.theme-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.theme-btn {
  width: 100%;
}

.theme-btn.active {
  border-color: rgba(59, 130, 246, 0.55);
  background: rgba(59, 130, 246, 0.12);
  color: #1d4ed8;
}

.is-dark .theme-btn.active {
  border-color: rgba(96, 165, 250, 0.58);
  background: rgba(59, 130, 246, 0.2);
  color: #bfdbfe;
}

.theme-tip {
  margin: 0;
  font-size: 13px;
  color: #64748b;
}

.is-dark .theme-tip {
  color: #94a3b8;
}

@media (max-width: 640px) {
  .theme-grid {
    grid-template-columns: 1fr;
  }
}
</style>
