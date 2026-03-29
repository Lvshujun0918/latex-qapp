<template>
  <section class="app-page app-inner-page page-wrap" :class="{ 'is-dark': resolvedTheme === 'dark' }">
    <header class="app-page-header page-header">
      <span class="app-kicker">Progress Pulse</span>
      <h1>统计</h1>
      <p>关注总量变化、学科分布和趋势演进。</p>
    </header>

    <div class="kpi-grid">
      <Card class="app-page-shell kpi-card">
        <CardContent class="kpi-content">
          <div class="kpi-icon icon-total"><FileText class="h-4 w-4" /></div>
          <div>
            <p class="kpi-label">总错题</p>
            <p class="kpi-value">{{ totalCount }}</p>
          </div>
        </CardContent>
      </Card>

      <Card class="app-page-shell kpi-card">
        <CardContent class="kpi-content">
          <div class="kpi-icon icon-mastery"><Gauge class="h-4 w-4" /></div>
          <div>
            <p class="kpi-label">平均掌握度</p>
            <p class="kpi-value">{{ averageMastery }}%</p>
          </div>
        </CardContent>
      </Card>

      <Card class="app-page-shell kpi-card">
        <CardContent class="kpi-content">
          <div class="kpi-icon icon-review"><RotateCcw class="h-4 w-4" /></div>
          <div>
            <p class="kpi-label">累计复习次数</p>
            <p class="kpi-value">{{ totalReviews }}</p>
          </div>
        </CardContent>
      </Card>
    </div>

    <Card class="app-page-shell">
      <CardHeader>
        <CardTitle>学科分布</CardTitle>
      </CardHeader>
      <CardContent class="subject-panel">
        <div v-if="subjectRows.length" class="subject-list">
          <div v-for="row in subjectRows" :key="row.subject" class="subject-row">
            <div class="subject-head">
              <span>{{ formatSubject(row.subject) }}</span>
              <span>{{ row.count }} 题</span>
            </div>
            <div class="subject-track">
              <div class="subject-bar" :style="{ width: `${row.ratio}%` }" />
            </div>
          </div>
        </div>
        <p v-else class="empty-tip">暂无统计数据，先去录入题目吧。</p>
      </CardContent>
    </Card>

    <Card class="app-page-shell">
      <CardHeader>
        <CardTitle>最近更新</CardTitle>
      </CardHeader>
      <CardContent>
        <div v-if="recentRows.length" class="recent-list">
          <div v-for="item in recentRows" :key="item.id" class="recent-item">
            <div>
              <p class="recent-title">{{ item.title || '未命名题目' }}</p>
              <p class="recent-meta">{{ formatSubject(item.subject) }} · {{ item.questionType || '未知题型' }}</p>
            </div>
            <span class="recent-time">{{ item.updatedAt.slice(0, 10) }}</span>
          </div>
        </div>
        <p v-else class="empty-tip">暂无最近记录。</p>
      </CardContent>
    </Card>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue';
import { storeToRefs } from 'pinia';
import { FileText, Gauge, RotateCcw } from 'lucide-vue-next';
import { useTheme } from '@/composables/useTheme';
import { useRecordStore } from '@/stores/record';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

const recordStore = useRecordStore();
const { records } = storeToRefs(recordStore);
const { resolvedTheme } = useTheme();

const totalCount = computed(() => records.value.length);
const totalReviews = computed(() => records.value.reduce((sum, item) => sum + (item.reviewCount || 0), 0));
const averageMastery = computed(() => {
  if (!records.value.length) {
    return 0;
  }
  const total = records.value.reduce((sum, item) => sum + (item.masteryLevel || 0), 0);
  return Math.round(total / records.value.length);
});

const subjectRows = computed(() => {
  const map = new Map<string, number>();
  records.value.forEach((item) => {
    const key = item.subject || '未知';
    map.set(key, (map.get(key) || 0) + 1);
  });
  const total = Math.max(records.value.length, 1);
  return [...map.entries()]
    .map(([subject, count]) => ({ subject, count, ratio: Math.round((count / total) * 100) }))
    .sort((a, b) => b.count - a.count);
});

const recentRows = computed(() => {
  return [...records.value]
    .sort((a, b) => +new Date(b.updatedAt) - +new Date(a.updatedAt))
    .slice(0, 6);
});

onMounted(() => {
  if (!records.value.length) {
    recordStore.reload();
  }
});

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

.kpi-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.kpi-card {
  min-height: 88px;
}

.kpi-content {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px;
}

.kpi-icon {
  width: 30px;
  height: 30px;
  border-radius: 10px;
  display: grid;
  place-items: center;
}

.icon-total {
  color: #2563eb;
  background: rgba(37, 99, 235, 0.12);
}

.icon-mastery {
  color: #059669;
  background: rgba(5, 150, 105, 0.12);
}

.icon-review {
  color: #d97706;
  background: rgba(217, 119, 6, 0.12);
}

.kpi-label {
  margin: 0;
  font-size: 12px;
  color: #64748b;
}

.kpi-value {
  margin: 2px 0 0;
  font-size: 20px;
  font-weight: 700;
  color: #0f172a;
}

.subject-panel {
  display: grid;
  gap: 8px;
}

.subject-list {
  display: grid;
  gap: 10px;
}

.subject-row {
  display: grid;
  gap: 6px;
}

.subject-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 13px;
  color: #334155;
}

.subject-track {
  width: 100%;
  height: 8px;
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.2);
  overflow: hidden;
}

.subject-bar {
  height: 100%;
  border-radius: 999px;
  background: linear-gradient(90deg, #2563eb 0%, #0ea5e9 100%);
}

.recent-list {
  display: grid;
  gap: 10px;
}

.recent-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 12px;
  border: 1px solid rgba(148, 163, 184, 0.28);
}

.recent-title {
  margin: 0;
  font-size: 14px;
  color: #0f172a;
}

.recent-meta {
  margin: 4px 0 0;
  font-size: 12px;
  color: #64748b;
}

.recent-time {
  font-size: 12px;
  color: #64748b;
  white-space: nowrap;
}

.empty-tip {
  margin: 0;
  color: #64748b;
  font-size: 13px;
}

.is-dark .kpi-label,
.is-dark .recent-meta,
.is-dark .recent-time,
.is-dark .empty-tip {
  color: #cbd5e1;
}

.is-dark .kpi-value,
.is-dark .recent-title,
.is-dark .subject-head {
  color: #f1f5f9;
}

.is-dark .recent-item {
  border-color: rgba(148, 163, 184, 0.22);
  background: rgba(30, 41, 59, 0.42);
}

@media (max-width: 900px) {
  .kpi-grid {
    grid-template-columns: 1fr;
  }
}
</style>
