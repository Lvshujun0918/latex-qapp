<template>
  <section class="app-page app-inner-page page-wrap" :class="{ 'is-dark': resolvedTheme === 'dark' }">
    <header class="app-page-header page-header">
      <span class="app-kicker">Progress Pulse</span>
      <h1>统计</h1>
      <p>看清增长趋势、学科结构和下一步复习重点。</p>
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

      <Card class="app-page-shell kpi-card">
        <CardContent class="kpi-content">
          <div class="kpi-icon icon-due"><Clock3 class="h-4 w-4" /></div>
          <div>
            <p class="kpi-label">今日到期复习</p>
            <p class="kpi-value">{{ dueTodayCount }}</p>
          </div>
        </CardContent>
      </Card>
    </div>

    <div class="trend-legend">
      <span><i class="dot dot-new" />新增错题</span>
      <span><i class="dot dot-review" />复习记录</span>
    </div>

    <div class="trend-wrap" v-if="trendMaxValue > 0">
      <svg viewBox="0 0 320 160" class="trend-chart" role="img" aria-label="最近14天新增与复习趋势">
        <line x1="12" y1="132" x2="308" y2="132" class="axis" />
        <polyline :points="trendPointsNew" class="line-new" />
        <polyline :points="trendPointsReview" class="line-review" />
        <circle
          v-for="(v, idx) in recentNewSeries"
          :key="`n-${idx}`"
          :cx="trendX(idx)"
          :cy="trendY(v)"
          r="2.1"
          class="point-new"
        />
        <circle
          v-for="(v, idx) in recentReviewSeries"
          :key="`r-${idx}`"
          :cx="trendX(idx)"
          :cy="trendY(v)"
          r="2.1"
          class="point-review"
        />
      </svg>

      <div class="trend-x-labels">
        <span>{{ recentDayLabels[0] }}</span>
        <span>{{ recentDayLabels[Math.floor(recentDayLabels.length / 2)] }}</span>
        <span>{{ recentDayLabels[recentDayLabels.length - 1] }}</span>
      </div>
    </div>

    <p v-else class="empty-tip">最近14天暂无可视化趋势数据。</p>
    <div v-if="subjectRows.length" class="subject-viz-wrap">
      <div class="subject-donut" :style="subjectDonutStyle">
        <div class="subject-donut-inner">
          <p>总量</p>
          <strong>{{ totalCount }}</strong>
        </div>
      </div>

      <div class="subject-list">
        <div v-for="row in subjectRows" :key="row.subject" class="subject-row">
          <div class="subject-head">
            <span class="subject-name"><i class="dot" :style="{ background: row.color }" />{{ formatSubject(row.subject) }}</span>
            <span>{{ row.count }} 题 · {{ row.ratio }}%</span>
          </div>
          <div class="subject-track">
            <div class="subject-bar" :style="{ width: `${row.ratio}%`, background: row.color }" />
          </div>
        </div>
      </div>
    </div>

    <p v-else class="empty-tip">暂无统计数据，先去录入题目吧。</p>

    <Card class="app-page-shell">
      <CardHeader>
        <CardTitle>复习建议</CardTitle>
      </CardHeader>
      <CardContent class="insight-panel">
        <div class="insight-item">
          <p class="insight-title">优先学科</p>
          <p class="insight-main">{{ topWeakSubjectText }}</p>
          <p class="insight-sub">按平均掌握度最低优先排序</p>
        </div>

        <div class="insight-item">
          <p class="insight-title">今日动作</p>
          <p class="insight-main">{{ dueTodayCount }} 题到期</p>
          <p class="insight-sub">建议先完成到期题，再处理新错题</p>
        </div>

        <div class="insight-item">
          <p class="insight-title">高频复习池</p>
          <p class="insight-main">{{ highReviewCount }} 题</p>
          <p class="insight-sub">复习次数≥3，适合做错因归类</p>
        </div>
      </CardContent>
    </Card>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, type CSSProperties } from 'vue';
import { storeToRefs } from 'pinia';
import { Clock3, FileText, Gauge, RotateCcw } from 'lucide-vue-next';
import { useTheme } from '@/composables/useTheme';
import { useRecordStore } from '@/stores/record';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

const recordStore = useRecordStore();
const { records } = storeToRefs(recordStore);
const { resolvedTheme } = useTheme();

const totalCount = computed(() => records.value.length);
const totalReviews = computed(() => records.value.reduce((sum, item) => sum + (item.reviewCount || 0), 0));
const now = computed(() => Date.now());
const averageMastery = computed(() => {
  if (!records.value.length) {
    return 0;
  }
  const total = records.value.reduce((sum, item) => sum + (item.masteryLevel || 0), 0);
  return Math.round(total / records.value.length);
});

const dueTodayCount = computed(() => {
  const ms = now.value;
  return records.value.filter((item) => {
    const next = Date.parse(item.nextReviewAt || '');
    return Number.isFinite(next) && next <= ms;
  }).length;
});

const highReviewCount = computed(() => records.value.filter((item) => (item.reviewCount || 0) >= 3).length);

const chartPalette = ['var(--color-chart-1)', 'var(--color-chart-2)', 'var(--color-chart-3)', 'var(--color-chart-4)', 'var(--color-chart-5)'];

const subjectRows = computed(() => {
  const map = new Map<string, number>();
  const masteryMap = new Map<string, number[]>();
  records.value.forEach((item) => {
    const key = item.subject || '未知';
    map.set(key, (map.get(key) || 0) + 1);
    const arr = masteryMap.get(key) || [];
    arr.push(Number(item.masteryLevel || 0));
    masteryMap.set(key, arr);
  });
  const total = Math.max(records.value.length, 1);
  return [...map.entries()]
    .map(([subject, count], index) => {
      const masteryList = masteryMap.get(subject) || [];
      const avgMastery = masteryList.length
        ? Math.round(masteryList.reduce((sum, v) => sum + v, 0) / masteryList.length)
        : 0;
      return {
        subject,
        count,
        ratio: Math.round((count / total) * 100),
        avgMastery,
        color: chartPalette[index % chartPalette.length],
      };
    })
    .sort((a, b) => b.count - a.count);
});

const topWeakSubjectText = computed(() => {
  if (!subjectRows.value.length) {
    return '暂无数据';
  }
  const weak = [...subjectRows.value].sort((a, b) => a.avgMastery - b.avgMastery)[0];
  return `${formatSubject(weak.subject)}（${weak.avgMastery}%）`;
});

const subjectDonutStyle = computed<CSSProperties>(() => {
  if (!subjectRows.value.length) {
    return { background: 'conic-gradient(var(--muted) 0 100%)' };
  }

  let cursor = 0;
  const segments = subjectRows.value.map((row) => {
    const start = cursor;
    cursor += row.ratio;
    return `${row.color} ${start}% ${Math.min(cursor, 100)}%`;
  });

  if (cursor < 100) {
    segments.push(`var(--muted) ${cursor}% 100%`);
  }

  return {
    background: `conic-gradient(${segments.join(', ')})`,
  };
});

const recentDayLabels = computed(() => {
  const labels: string[] = [];
  const base = new Date();
  base.setHours(0, 0, 0, 0);

  for (let i = 13; i >= 0; i -= 1) {
    const d = new Date(base);
    d.setDate(base.getDate() - i);
    labels.push(`${d.getMonth() + 1}/${d.getDate()}`);
  }

  return labels;
});

const recentNewSeries = computed(() => {
  const counter = new Map<string, number>();
  records.value.forEach((item) => {
    const date = new Date(item.createdAt || '');
    if (Number.isNaN(+date)) {
      return;
    }
    const key = `${date.getMonth() + 1}/${date.getDate()}`;
    counter.set(key, (counter.get(key) || 0) + 1);
  });
  return recentDayLabels.value.map((label) => counter.get(label) || 0);
});

const recentReviewSeries = computed(() => {
  const counter = new Map<string, number>();
  records.value.forEach((item) => {
    const date = new Date(item.lastReviewedAt || '');
    if (Number.isNaN(+date)) {
      return;
    }
    const key = `${date.getMonth() + 1}/${date.getDate()}`;
    counter.set(key, (counter.get(key) || 0) + 1);
  });
  return recentDayLabels.value.map((label) => counter.get(label) || 0);
});

const trendMaxValue = computed(() => Math.max(1, ...recentNewSeries.value, ...recentReviewSeries.value));

function trendX(index: number) {
  if (recentDayLabels.value.length <= 1) {
    return 12;
  }
  return 12 + (296 * index) / (recentDayLabels.value.length - 1);
}

function trendY(value: number) {
  return 132 - (112 * value) / trendMaxValue.value;
}

const trendPointsNew = computed(() => recentNewSeries.value.map((v, i) => `${trendX(i)},${trendY(v)}`).join(' '));
const trendPointsReview = computed(() => recentReviewSeries.value.map((v, i) => `${trendX(i)},${trendY(v)}`).join(' '));

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
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
}

.kpi-card {
  min-height: 88px;
}

.kpi-content {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0px;
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

.icon-due {
  color: #7c3aed;
  background: rgba(124, 58, 237, 0.12);
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

.chart-grid {
  display: grid;
  grid-template-columns: 1.35fr 1fr;
  gap: 10px;
}

.trend-panel {
  display: grid;
  gap: 10px;
}

.trend-legend {
  display: flex;
  gap: 14px;
  flex-wrap: wrap;
  font-size: 12px;
  color: #64748b;
}

.dot {
  width: 10px;
  height: 10px;
  border-radius: 999px;
  display: inline-block;
  margin-right: 6px;
  transform: translateY(1px);
}

.dot-new {
  background: #2563eb;
}

.dot-review {
  background: #f97316;
}

.trend-wrap {
  border-radius: 14px;
  border: 1px solid rgba(148, 163, 184, 0.25);
  background: rgba(248, 250, 252, 0.7);
  padding: 8px 10px 10px;
}

.trend-chart {
  width: 100%;
  height: auto;
}

.axis {
  stroke: rgba(148, 163, 184, 0.55);
  stroke-width: 1;
}

.line-new {
  fill: none;
  stroke: #2563eb;
  stroke-width: 2.4;
}

.line-review {
  fill: none;
  stroke: #f97316;
  stroke-width: 2.4;
}

.point-new {
  fill: #2563eb;
}

.point-review {
  fill: #f97316;
}

.trend-x-labels {
  display: flex;
  justify-content: space-between;
  margin-top: 2px;
  font-size: 11px;
  color: #64748b;
}

.subject-list {
  display: grid;
  gap: 10px;
}

.subject-viz-wrap {
  display: grid;
  grid-template-columns: 132px 1fr;
  gap: 12px;
  align-items: center;
  border-radius: 14px;
  border: 1px solid rgba(148, 163, 184, 0.25);
  background: rgba(248, 250, 252, 0.7);
  padding: 8px 10px 10px;
}

.subject-donut {
  width: 120px;
  height: 120px;
  border-radius: 999px;
  display: grid;
  place-items: center;
}

.subject-donut-inner {
  width: 74px;
  height: 74px;
  border-radius: 999px;
  display: grid;
  place-content: center;
  text-align: center;
  background: rgba(255, 255, 255, 0.9);
}

.subject-donut-inner p {
  margin: 0;
  font-size: 11px;
  color: #64748b;
}

.subject-donut-inner strong {
  margin-top: 2px;
  font-size: 18px;
  color: #0f172a;
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

.subject-name {
  display: inline-flex;
  align-items: center;
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
}

.insight-panel {
  display: grid;
  gap: 10px;
}

.insight-item {
  border: 1px solid rgba(148, 163, 184, 0.28);
  border-radius: 12px;
  background: rgba(248, 250, 252, 0.74);
  padding: 10px 12px;
  display: grid;
  gap: 3px;
}

.insight-title {
  margin: 0;
  font-size: 12px;
  color: #64748b;
}

.insight-main {
  margin: 0;
  font-size: 15px;
  font-weight: 700;
  color: #0f172a;
}

.insight-sub {
  margin: 0;
  font-size: 12px;
  color: #64748b;
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
.is-dark .empty-tip,
.is-dark .trend-legend,
.is-dark .trend-x-labels,
.is-dark .subject-donut-inner p,
.is-dark .insight-title,
.is-dark .insight-sub {
  color: #cbd5e1;
}

.is-dark .kpi-value,
.is-dark .recent-title,
.is-dark .subject-head,
.is-dark .subject-donut-inner strong,
.is-dark .insight-main {
  color: #f1f5f9;
}

.is-dark .recent-item {
  border-color: rgba(148, 163, 184, 0.22);
  background: rgba(30, 41, 59, 0.42);
}

.is-dark .trend-wrap,
.is-dark .insight-item,
.is-dark .subject-viz-wrap {
  border-color: rgba(148, 163, 184, 0.22);
  background: rgba(30, 41, 59, 0.42);
}

.is-dark .subject-donut-inner {
  background: rgba(15, 23, 42, 0.82);
}

@media (max-width: 900px) {
  .kpi-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .chart-grid {
    grid-template-columns: 1fr;
  }

  .subject-viz-wrap {
    grid-template-columns: 1fr;
    justify-items: center;
  }
}
</style>
