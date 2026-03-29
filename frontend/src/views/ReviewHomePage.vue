<template>
  <section class="app-page app-inner-page page-wrap" :class="{ 'is-dark': resolvedTheme === 'dark' }">
    <header class="app-page-header page-header">
      <span class="app-kicker">Daily Focus</span>
      <h1>复习</h1>
      <p>按艾宾浩斯遗忘曲线安排今日复习清单。</p>
    </header>

    <div class="overview-grid">
      <Card class="app-page-shell metric-card">
        <CardContent class="metric-content">
          <p class="metric-label">今日应复习</p>
          <p class="metric-value">{{ dueToday.length }}</p>
        </CardContent>
      </Card>

      <Card class="app-page-shell metric-card">
        <CardContent class="metric-content">
          <p class="metric-label">明日将到期</p>
          <p class="metric-value">{{ dueTomorrow.length }}</p>
        </CardContent>
      </Card>

      <Card class="app-page-shell metric-card">
        <CardContent class="metric-content">
          <p class="metric-label">平均掌握度</p>
          <p class="metric-value">{{ averageMastery }}%</p>
        </CardContent>
      </Card>
    </div>

    <Card class="app-page-shell">
      <CardHeader>
        <CardTitle>今日复习清单</CardTitle>
      </CardHeader>
      <CardContent>
        <div v-if="dueToday.length" class="review-list">
          <button
            v-for="item in dueToday"
            :key="item.id"
            class="review-item"
            type="button"
            @click="goDetail(item.id)"
          >
            <div>
              <p class="review-title">{{ item.title || '未命名题目' }}</p>
              <p class="review-meta">{{ formatSubject(item.subject) }} · 第 {{ item.reviewCount + 1 }} 次复习 · 间隔 {{ item.nextInterval }} 天</p>
            </div>
            <span class="review-urgency">到期 {{ item.overdueDays }} 天</span>
          </button>
        </div>
        <p v-else class="empty-tip">今天没有到期复习，保持节奏即可。</p>
      </CardContent>
    </Card>

    <Card class="app-page-shell">
      <CardHeader>
        <CardTitle>复习节奏</CardTitle>
      </CardHeader>
      <CardContent class="curve-text">
        系统使用间隔序列：1 天、2 天、4 天、7 天、15 天、30 天。
        每道题根据当前复习次数匹配下一个间隔，达到或超过间隔即进入“应复习”。
      </CardContent>
    </Card>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue';
import { storeToRefs } from 'pinia';
import { useRouter } from 'vue-router';
import { useTheme } from '@/composables/useTheme';
import { useRecordStore } from '@/stores/record';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

const EBBINGHAUS_INTERVALS = [1, 2, 4, 7, 15, 30];

const router = useRouter();
const recordStore = useRecordStore();
const { records } = storeToRefs(recordStore);
const { resolvedTheme } = useTheme();

const scheduleRows = computed(() => {
  const now = Date.now();
  return records.value
    .map((record) => {
      const reviewCount = Math.max(0, Number(record.reviewCount || 0));
      const intervalIndex = Math.min(reviewCount, EBBINGHAUS_INTERVALS.length - 1);
      const nextInterval = EBBINGHAUS_INTERVALS[intervalIndex];
      const baseTime = Date.parse(record.updatedAt || record.createdAt || '');
      const base = Number.isFinite(baseTime) ? baseTime : now;
      const elapsedDays = Math.floor((now - base) / 86400000);
      const overdueDays = elapsedDays - nextInterval;

      return {
        ...record,
        nextInterval,
        overdueDays,
      };
    })
    .sort((a, b) => b.overdueDays - a.overdueDays);
});

const dueToday = computed(() => scheduleRows.value.filter((row) => row.overdueDays >= 0));
const dueTomorrow = computed(() => scheduleRows.value.filter((row) => row.overdueDays === -1));
const averageMastery = computed(() => {
  if (!records.value.length) {
    return 0;
  }
  const total = records.value.reduce((sum, item) => sum + Number(item.masteryLevel || 0), 0);
  return Math.round(total / records.value.length);
});

onMounted(() => {
  if (!records.value.length) {
    recordStore.reload();
  }
});

function goDetail(id: number) {
  router.push(`/records/${id}`);
}

function formatSubject(subject?: string) {
  const value = String(subject || '').trim().toLowerCase();
  if (value === 'math' || value === '数学') return '数学';
  if (value === 'physics' || value === '物理') return '物理';
  if (value === 'chemistry' || value === '化学') return '化学';
  if (value === 'biology' || value === '生物') return '生物';
  return subject || '未知';
}
</script>

<style scoped>
.page-wrap {
  gap: 14px;
}

.overview-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.metric-card {
  min-height: 84px;
}

.metric-content {
  display: grid;
  gap: 4px;
  padding: 14px;
}

.metric-label {
  margin: 0;
  font-size: 12px;
  color: #64748b;
}

.metric-value {
  margin: 0;
  font-size: 24px;
  line-height: 1;
  font-weight: 700;
  color: #0f172a;
}

.review-list {
  display: grid;
  gap: 10px;
}

.review-item {
  width: 100%;
  border: 1px solid rgba(148, 163, 184, 0.28);
  border-radius: 12px;
  background: #fff;
  padding: 12px;
  text-align: left;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  cursor: pointer;
}

.review-title {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: #0f172a;
}

.review-meta {
  margin: 4px 0 0;
  font-size: 12px;
  color: #64748b;
}

.review-urgency {
  font-size: 12px;
  color: #b45309;
  white-space: nowrap;
}

.curve-text {
  color: #475569;
  font-size: 13px;
  line-height: 1.7;
}

.empty-tip {
  margin: 0;
  color: #64748b;
  font-size: 13px;
}

.is-dark .metric-label,
.is-dark .review-meta,
.is-dark .curve-text,
.is-dark .empty-tip {
  color: #cbd5e1;
}

.is-dark .metric-value,
.is-dark .review-title {
  color: #f1f5f9;
}

.is-dark .review-item {
  background: rgba(30, 41, 59, 0.95);
  border-color: rgba(148, 163, 184, 0.25);
}

@media (max-width: 900px) {
  .overview-grid {
    grid-template-columns: 1fr;
  }

  .review-item {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
