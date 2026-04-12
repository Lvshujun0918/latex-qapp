<template>
  <section class="app-page page-wrap" :class="{ 'is-dark': resolvedTheme === 'dark' }">
    <header class="app-page-header page-header">
      <span class="app-kicker">Profile</span>
      <h1>我的</h1>
      <p>家长账号中心：管理陪学资料、导出记录与界面设置。</p>
    </header>

    <div class="profile-main">
      <div class="profile-avatar">{{ displayInitial }}</div>
      <div class="profile-meta">
        <h2>{{ authStore.displayName || authStore.username || '未登录' }}</h2>
        <div class="profile-badges">
          <Badge variant="secondary" class="profile-badge">家长陪学端</Badge>
          <Badge variant="outline" class="profile-badge">{{ resolvedTheme === 'dark' ? '深色主题' : '浅色主题' }}</Badge>
        </div>
      </div>
    </div>

    <div class="profile-stats">
      <div class="stat-item">
        <span>题库总量</span>
        <strong>{{ recordStore.records.length }}</strong>
      </div>
      <div class="stat-item">
        <span>主题模式</span>
        <strong>{{ themeModeLabel }}</strong>
      </div>
    </div>

    <div class="pdf-history app-soft-card">
      <div class="pdf-history-header">
        <div>
          <p>PDF 导出历史</p>
          <span>家长可回看并下载历史导出文档。</span>
        </div>
        <Button variant="outline" size="sm" @click="loadPdfHistory">刷新</Button>
      </div>

      <div v-if="pdfHistory.length" class="pdf-history-list">
        <article v-for="item in pdfHistory" :key="item.jobId" class="pdf-item">
          <div class="pdf-item-main">
            <p class="pdf-item-title">{{ item.source === 'review' ? '今日复习题单导出' : '错题列表导出' }}</p>
            <p class="pdf-item-meta">
              {{ formatHistoryTime(item.createdAt) }} · {{ item.selectedCount }} 题 · 任务 {{ item.jobId }}
            </p>
          </div>

          <div class="pdf-item-actions">
            <Button v-if="isNativePlatform" size="sm" :disabled="openingJobId === item.jobId" @click="openHistoryPdf(item)">
              {{ openingJobId === item.jobId ? '打开中...' : '下载并打开' }}
            </Button>
            <a
              v-else
              class="pdf-download-link"
              :href="resolvePdfUrl(item.pdfFileUrl)"
              target="_blank"
              rel="noopener noreferrer"
            >
              下载 PDF
            </a>
          </div>
        </article>
      </div>

      <p v-else class="pdf-empty">暂无导出记录。可在错题页或复习页导出后回到这里下载。</p>
      <p v-if="pdfHistoryMessage" class="pdf-message">{{ pdfHistoryMessage }}</p>
    </div>

    <div class="theme-switch-list">
      <div class="theme-switch-item" :class="{ disabled: followSystemEnabled }">
        <div class="theme-switch-text">
          <p>深色模式</p>
          <span>关闭为浅色，开启为深色</span>
        </div>
        <Switch :model-value="darkModeEnabled" :disabled="followSystemEnabled" @update:model-value="setDarkMode" />
      </div>

      <div class="theme-switch-item">
        <div class="theme-switch-text">
          <p>跟随系统</p>
          <span>自动跟随设备外观设置</span>
        </div>
        <Switch :model-value="followSystemEnabled" @update:model-value="setFollowSystem" />
      </div>
    </div>

    <button class="about-entry app-interactive-surface" type="button" @click="goAbout">
      <div class="about-entry-text">
        <p>关于应用</p>
        <span>版本信息、数据说明与产品介绍</span>
      </div>
      <ChevronRight class="h-4 w-4 about-entry-arrow" />
    </button>

    <Button variant="destructive" class="logout-btn" @click="logout">退出登录</Button>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { Capacitor } from '@capacitor/core';
import { ChevronRight } from 'lucide-vue-next';
import { useAuthStore } from '@/stores/auth';
import { useRecordStore } from '@/stores/record';
import { useTheme } from '@/composables/useTheme';
import type { PdfExportHistoryItem } from '@/services/pdf-history';
import { getPdfExportHistory, resolvePdfUrl } from '@/services/pdf-history';
import { openPdfFromLocalUri, saveRemotePdfToDevice } from '@/services/pdf-native';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Switch } from '@/components/ui/switch';

const authStore = useAuthStore();
const recordStore = useRecordStore();
const router = useRouter();
const { themeMode, resolvedTheme, setTheme } = useTheme();
const pdfHistory = ref<PdfExportHistoryItem[]>([]);
const pdfHistoryMessage = ref('');
const openingJobId = ref('');
const isNativePlatform = Capacitor.isNativePlatform();

const displayInitial = computed(() => {
  const source = authStore.displayName || authStore.username || '用户';
  return source.trim().slice(0, 1).toUpperCase();
});

const themeModeLabel = computed(() => {
  if (themeMode.value === 'light') {
    return '浅色';
  }
  if (themeMode.value === 'dark') {
    return '深色';
  }
  return '跟随系统';
});

const followSystemEnabled = computed(() => themeMode.value === 'system');
const darkModeEnabled = computed(() => resolvedTheme.value === 'dark');

function setDarkMode(checked: boolean) {
  if (followSystemEnabled.value) {
    return;
  }
  setTheme(checked ? 'dark' : 'light');
}

function setFollowSystem(checked: boolean) {
  if (checked) {
    setTheme('system');
    return;
  }
  setTheme(resolvedTheme.value === 'dark' ? 'dark' : 'light');
}

onMounted(() => {
  recordStore.reload();
  loadPdfHistory();
});

function loadPdfHistory() {
  pdfHistory.value = getPdfExportHistory({
    userId: authStore.userId,
    username: authStore.username,
  });
}

function formatHistoryTime(value: string) {
  const ts = Date.parse(value || '');
  if (!Number.isFinite(ts)) {
    return '未知时间';
  }
  return new Date(ts).toLocaleString();
}

async function openHistoryPdf(item: PdfExportHistoryItem) {
  if (!item.pdfFileUrl || openingJobId.value) {
    return;
  }

  openingJobId.value = item.jobId;
  pdfHistoryMessage.value = '';
  try {
    const url = resolvePdfUrl(item.pdfFileUrl);
    const fileName = `latex-qapp-${item.jobId}`;
    const result = await saveRemotePdfToDevice(url, fileName);
    await openPdfFromLocalUri(result.uri);
    pdfHistoryMessage.value = `已打开：${result.fileName}`;
  } catch (error: any) {
    pdfHistoryMessage.value = error?.message || '打开失败，请稍后重试';
  } finally {
    openingJobId.value = '';
  }
}

function goAbout() {
  router.push('/profile/about');
}

function logout() {
  authStore.logout();
  router.replace('/login');
}
</script>

<style scoped>
.page-wrap {
  gap: 14px;
}

.profile-hero {
  overflow: hidden;
}

.profile-hero-content {
  display: grid;
  gap: 16px;
}

.profile-main {
  display: flex;
  align-items: center;
  gap: 14px;
}

.profile-avatar {
  width: 52px;
  height: 52px;
  border-radius: 16px;
  display: grid;
  place-items: center;
  font-size: 20px;
  font-weight: 700;
  color: #ffffff;
  background: linear-gradient(145deg, #0ea5e9 0%, #2563eb 60%, #1d4ed8 100%);
  box-shadow: 0 10px 24px rgba(37, 99, 235, 0.32);
}

.profile-meta {
  min-width: 0;
}

.profile-label {
  margin: 0;
  font-size: 12px;
  color: #64748b;
}

.profile-meta h2 {
  margin: 2px 0 0;
  font-size: 22px;
  line-height: 1.1;
  color: #0f172a;
}

.profile-sub {
  margin: 6px 0 0;
  font-size: 13px;
  color: #64748b;
}

.profile-badges {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.profile-badge {
  border-radius: 999px;
}

.profile-stats {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

.stat-item {
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 14px;
  background: rgba(248, 250, 252, 0.75);
  padding: 10px 12px;
  display: grid;
  gap: 2px;
}

.stat-item span {
  font-size: 12px;
  color: #64748b;
}

.stat-item strong {
  font-size: 16px;
  color: #0f172a;
}

.pdf-history {
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 14px;
  padding: 12px;
  display: grid;
  gap: 10px;
}

.pdf-history-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.pdf-history-header p {
  margin: 0;
  font-size: 14px;
  font-weight: 700;
  color: #0f172a;
}

.pdf-history-header span {
  margin-top: 2px;
  display: block;
  font-size: 12px;
  color: #64748b;
}

.pdf-history-list {
  display: grid;
  gap: 8px;
}

.pdf-item {
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 12px;
  padding: 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  background: rgba(248, 250, 252, 0.7);
}

.pdf-item-main {
  min-width: 0;
}

.pdf-item-title {
  margin: 0;
  font-size: 13px;
  font-weight: 700;
  color: #0f172a;
}

.pdf-item-meta {
  margin: 4px 0 0;
  font-size: 12px;
  color: #64748b;
}

.pdf-item-actions {
  flex: 0 0 auto;
}

.pdf-download-link {
  color: #f8fafc;
  text-decoration: none;
  background: linear-gradient(110deg, #2563eb 0%, #1d4ed8 55%, #1e40af 100%);
  border-radius: 10px;
  padding: 8px 12px;
  font-size: 12px;
  font-weight: 600;
}

.pdf-empty,
.pdf-message {
  margin: 0;
  font-size: 12px;
  color: #64748b;
}

.theme-wrap {
  display: grid;
  gap: 10px;
}

.theme-switch-list {
  display: grid;
  gap: 10px;
}

.theme-switch-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 14px;
  background: rgba(248, 250, 252, 0.72);
  padding: 12px;
}

.theme-switch-item.disabled {
  opacity: 0.62;
}

.theme-switch-text {
  min-width: 0;
}

.theme-switch-text p {
  margin: 0;
  font-size: 14px;
  font-weight: 700;
  color: #0f172a;
}

.theme-switch-text span {
  margin-top: 2px;
  display: block;
  font-size: 12px;
  color: #64748b;
}

.about-entry {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 14px;
  background: rgba(248, 250, 252, 0.72);
  padding: 12px;
  text-align: left;
}

.about-entry-text {
  min-width: 0;
}

.about-entry-text p {
  margin: 0;
  font-size: 14px;
  font-weight: 700;
  color: #0f172a;
}

.about-entry-text span {
  margin-top: 2px;
  display: block;
  font-size: 12px;
  color: #64748b;
}

.about-entry-arrow {
  color: #64748b;
}

.actions-wrap {
  display: grid;
  gap: 10px;
}

.actions-tip {
  margin: 0;
  font-size: 13px;
  color: #64748b;
}

.logout-btn {
  width: 100%;
}

.theme-tip {
  margin: 0;
  font-size: 13px;
  color: #64748b;
}

.is-dark .theme-tip {
  color: #94a3b8;
}

.is-dark .profile-label,
.is-dark .profile-sub,
.is-dark .stat-item span,
.is-dark .actions-tip {
  color: #94a3b8;
}

.is-dark .profile-meta h2,
.is-dark .stat-item strong,
.is-dark .pdf-history-header p,
.is-dark .pdf-item-title {
  color: #f8fafc;
}

.is-dark .stat-item {
  border-color: rgba(148, 163, 184, 0.3);
  background: rgba(15, 23, 42, 0.5);
}

.is-dark .pdf-history,
.is-dark .pdf-item {
  border-color: rgba(148, 163, 184, 0.3);
  background: rgba(15, 23, 42, 0.5);
}

.is-dark .theme-switch-item {
  border-color: rgba(148, 163, 184, 0.3);
  background: rgba(15, 23, 42, 0.5);
}

.is-dark .about-entry {
  border-color: rgba(148, 163, 184, 0.3);
  background: rgba(15, 23, 42, 0.5);
}

.is-dark .theme-switch-text p {
  color: #f8fafc;
}

.is-dark .theme-switch-text span {
  color: #94a3b8;
}

.is-dark .pdf-history-header span,
.is-dark .pdf-item-meta,
.is-dark .pdf-empty,
.is-dark .pdf-message {
  color: #94a3b8;
}

.is-dark .about-entry-text p {
  color: #f8fafc;
}

.is-dark .about-entry-text span,
.is-dark .about-entry-arrow {
  color: #94a3b8;
}
</style>
