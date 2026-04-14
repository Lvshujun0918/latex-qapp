<template>
  <section class="app-page page-wrap" :class="{ 'is-dark': resolvedTheme === 'dark' }">
    <header class="app-page-header page-header">
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

      <div class="theme-switch-item" :class="{ disabled: !androidReminderAvailable }">
        <div class="theme-switch-text">
          <p>每日复习提醒</p>
          <span>
            {{ androidReminderAvailable ? '安卓原生通知，每天定时提醒复习' : '仅安卓原生 App 可用' }}
          </span>
        </div>
        <Switch :model-value="reminderEnabled" :disabled="!androidReminderAvailable || reminderLoading" @update:model-value="toggleReminder" />
      </div>

      <div class="theme-switch-item" :class="{ disabled: !reminderEnabled || !androidReminderAvailable }">
        <div class="theme-switch-text">
          <p>提醒时间</p>
          <span>默认每天 20:00，可自定义</span>
        </div>
        <input class="time-input" type="time" :disabled="!reminderEnabled || !androidReminderAvailable || reminderLoading" :value="reminderTime" @change="onReminderTimeChange" />
      </div>
    </div>

    <p v-if="reminderMessage" class="theme-tip">{{ reminderMessage }}</p>

    <button class="about-entry app-interactive-surface" type="button" @click="goAbout">
      <div class="about-entry-text">
        <p>关于应用</p>
        <span>版本信息、数据说明与产品介绍</span>
      </div>
      <ChevronRight class="h-4 w-4 about-entry-arrow" />
    </button>

    <button class="about-entry app-interactive-surface" type="button" @click="goBackendStatus">
      <div class="about-entry-text">
        <p>后端状态</p>
        <span>查看后端时延、服务状态和版本</span>
      </div>
      <ChevronRight class="h-4 w-4 about-entry-arrow" />
    </button>

    <Button variant="destructive" class="logout-btn" @click="logout">退出登录</Button>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { ChevronRight } from 'lucide-vue-next';
import { useAuthStore } from '@/stores/auth';
import { useRecordStore } from '@/stores/record';
import { useTheme } from '@/composables/useTheme';
import {
  cancelDailyReminder,
  ensureReminderPermissions,
  getReminderEnabled,
  getReminderTime,
  isNativeAndroid,
  scheduleDailyReminder,
  setReminderEnabled,
  setReminderTime,
} from '@/services/reminder';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Switch } from '@/components/ui/switch';

const authStore = useAuthStore();
const recordStore = useRecordStore();
const router = useRouter();
const { themeMode, resolvedTheme, setTheme } = useTheme();
const reminderEnabled = ref(false);
const reminderTime = ref('20:00');
const reminderLoading = ref(false);
const reminderMessage = ref('');

const androidReminderAvailable = computed(() => isNativeAndroid());

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
  reminderEnabled.value = getReminderEnabled();
  reminderTime.value = getReminderTime();
});

async function toggleReminder(checked: boolean) {
  if (!androidReminderAvailable.value || reminderLoading.value) {
    return;
  }

  reminderLoading.value = true;
  try {
    reminderMessage.value = '';

    if (checked) {
      await ensureReminderPermissions();
      await scheduleDailyReminder(reminderTime.value);
      reminderEnabled.value = true;
      setReminderEnabled(true);
      reminderMessage.value = `已开启每日提醒：${reminderTime.value}`;
      return;
    }

    await cancelDailyReminder();
    reminderEnabled.value = false;
    setReminderEnabled(false);
    reminderMessage.value = '已关闭每日提醒';
  } catch (error: any) {
    reminderMessage.value = error?.message || '提醒设置失败，请重试';
    reminderEnabled.value = getReminderEnabled();
  } finally {
    reminderLoading.value = false;
  }
}

async function onReminderTimeChange(event: Event) {
  const next = (event.target as HTMLInputElement)?.value || '20:00';
  reminderTime.value = next;
  setReminderTime(next);

  if (!androidReminderAvailable.value || !reminderEnabled.value || reminderLoading.value) {
    return;
  }

  reminderLoading.value = true;
  try {
    await scheduleDailyReminder(next);
    reminderMessage.value = `提醒时间已更新为 ${next}`;
  } catch (error: any) {
    reminderMessage.value = error?.message || '更新时间失败，请重试';
  } finally {
    reminderLoading.value = false;
  }
}

function goAbout() {
  router.push('/profile/about');
}

function goBackendStatus() {
  router.push('/profile/backend');
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
  background: #2563eb;
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

.time-input {
  border: 1px solid rgba(148, 163, 184, 0.32);
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.85);
  color: #0f172a;
  padding: 6px 10px;
  font-size: 14px;
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
.is-dark .stat-item strong {
  color: #f8fafc;
}

.is-dark .stat-item {
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

.is-dark .time-input {
  background: rgba(15, 23, 42, 0.72);
  border-color: rgba(148, 163, 184, 0.35);
  color: #e2e8f0;
}

.is-dark .theme-switch-text span {
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
