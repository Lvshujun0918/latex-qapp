<template>
  <section class="app-page app-inner-page stats-page" :class="{ 'is-dark': resolvedTheme === 'dark' }">

    <header class="app-page-header page-header stats-header">
      <h1>统计概览</h1>
      <p>先看关系图谱，再看标签指标，复习策略围绕标签网络而不是单题展开。</p>
    </header>
    <article class="insight-item">
      <p class="insight-title">最需要先复习</p>
      <p class="insight-main">{{ topPriorityTagText }}</p>
      <p class="insight-sub">优先级综合错误率、频次、掌握度</p>
    </article>
    <article class="insight-item">
      <p class="insight-title">最稳定知识点</p>
      <p class="insight-main">{{ bestTagText }}</p>
      <p class="insight-sub">正确率高，建议降频巡检</p>
    </article>
    <article class="insight-item">
      <p class="insight-title">掌握度低知识点</p>
      <p class="insight-main">{{ highRiskTagCount }}个</p>
      <p class="insight-sub">错误率≥50% 或掌握度≤55%</p>
    </article>
    <TagRelationGraph :records="records" :theme="resolvedTheme" />

    <section class="stats-kpi-strip" role="list" aria-label="标签统计概览">
      <article class="kpi-item" role="listitem">
        <p>题目总量</p>
        <strong>{{ totalCount }}</strong>
        <span>全部错题节点</span>
      </article>
      <article class="kpi-item" role="listitem">
        <p>标签数量</p>
        <strong>{{ tagCount }}</strong>
        <span>可追踪知识标签</span>
      </article>
      <article class="kpi-item" role="listitem">
        <p>累计复习</p>
        <strong>{{ totalReviews }}</strong>
        <span>标签网络总复习频次</span>
      </article>
      <article class="kpi-item" role="listitem">
        <p>今日到期</p>
        <strong>{{ dueTodayCount }}</strong>
        <span>待处理复习任务</span>
      </article>
    </section>

    <div v-if="tagSummaries.length" class="priority-list">
      <article v-for="(tag, index) in topPriorityTags" :key="tag.tag" class="priority-item">
        <div class="priority-rank">{{ index + 1 }}</div>
        <div class="priority-main">
          <p class="priority-title">{{ tag.tag }}</p>
          <p class="priority-sub">{{ tag.recordCount }} 题 · 频次 {{ tag.avgReviewCount }} · 掌握度 {{ tag.avgMastery
          }}%</p>
          <div class="meter-row" aria-hidden="true">
            <div class="meter-track meter-wrong">
              <div class="meter-fill" :style="{ width: `${tag.wrongRate}%` }" />
            </div>
            <span>错 {{ tag.wrongRate }}%</span>
          </div>
        </div>
        <div class="priority-score">{{ tag.priorityScore }}</div>
      </article>
    </div>
    <p v-else class="empty-tip">暂无标签优先级数据。</p>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue';
import { storeToRefs } from 'pinia';
import { useTheme } from '@/composables/useTheme';
import { useRecordStore } from '@/stores/record';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { buildTagSummaries } from '@/utils/tag-analytics';
import TagRelationGraph from '@/components/TagRelationGraph.vue';

const recordStore = useRecordStore();
const { records } = storeToRefs(recordStore);
const { resolvedTheme } = useTheme();

const totalCount = computed(() => records.value.length);
const totalReviews = computed(() => records.value.reduce((sum, item) => sum + (item.reviewCount || 0), 0));

const dueTodayCount = computed(() => {
  const now = Date.now();
  return records.value.filter((item) => {
    const next = Date.parse(item.nextReviewAt || '');
    return Number.isFinite(next) && next <= now;
  }).length;
});

const tagSummaries = computed(() => buildTagSummaries(records.value));
const tagCount = computed(() => tagSummaries.value.length);
const topPriorityTags = computed(() => tagSummaries.value.slice(0, 5));

const topPriorityTagText = computed(() => {
  if (!tagSummaries.value.length) {
    return '暂无数据';
  }
  const top = tagSummaries.value[0];
  return `${top.tag}（错误率 ${top.wrongRate}% · 频次 ${top.avgReviewCount}）`;
});

const bestTagText = computed(() => {
  if (!tagSummaries.value.length) {
    return '暂无数据';
  }
  const best = [...tagSummaries.value].sort((a, b) => b.correctRate - a.correctRate || a.wrongRate - b.wrongRate)[0];
  return `${best.tag}（正确率 ${best.correctRate}%）`;
});

const highRiskTagCount = computed(() =>
  tagSummaries.value.filter((tag) => tag.wrongRate >= 50 || tag.avgMastery <= 55).length,
);

onMounted(() => {
  if (!records.value.length) {
    recordStore.reload();
  }
});
</script>

<style scoped>
.stats-page {
  gap: 14px;
}

.stats-header :deep(p) {
  max-width: 620px;
}

.stats-kpi-strip {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
}

.kpi-item {
  border-radius: 14px;
  border: 1px solid rgba(148, 163, 184, 0.24);
  background: rgba(248, 250, 252, 0.72);
  padding: 12px;
  display: grid;
  gap: 4px;
}

.kpi-item p {
  margin: 0;
  font-size: 12px;
  color: #64748b;
}

.kpi-item strong {
  font-size: 24px;
  line-height: 1;
  color: #0f172a;
}

.kpi-item span {
  font-size: 12px;
  color: #64748b;
}

.stats-grid {
  display: grid;
  grid-template-columns: 1.25fr 1fr;
  gap: 12px;
}

.priority-list {
  display: grid;
  gap: 10px;
}

.priority-item {
  display: grid;
  grid-template-columns: 30px 1fr auto;
  align-items: center;
  gap: 10px;
  border: 1px solid rgba(148, 163, 184, 0.24);
  border-radius: 12px;
  background: rgba(248, 250, 252, 0.72);
  padding: 10px;
}

.priority-rank {
  width: 30px;
  height: 30px;
  border-radius: 999px;
  display: grid;
  place-items: center;
  font-size: 12px;
  font-weight: 700;
  color: #fff;
  background: linear-gradient(135deg, #f97316, #ea580c);
}

.priority-main {
  min-width: 0;
  display: grid;
  gap: 4px;
}

.priority-title {
  margin: 0;
  font-size: 14px;
  font-weight: 700;
  color: #0f172a;
}

.priority-sub {
  margin: 0;
  font-size: 12px;
  color: #64748b;
}

.meter-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.meter-track {
  height: 8px;
  flex: 1;
  border-radius: 999px;
  overflow: hidden;
  background: rgba(148, 163, 184, 0.2);
}

.meter-fill {
  height: 100%;
  border-radius: 999px;
}

.meter-wrong .meter-fill {
  background: #ef4444;
}

.meter-row span {
  width: 52px;
  text-align: right;
  font-size: 11px;
  color: #64748b;
}

.priority-score {
  font-size: 20px;
  font-weight: 800;
  color: #0f172a;
}

.insight-panel {
  display: grid;
  gap: 10px;
}

.insight-item {
  border: 1px solid rgba(148, 163, 184, 0.26);
  border-radius: 12px;
  background: rgba(248, 250, 252, 0.72);
  padding: 11px 12px;
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

.tag-table-wrap {
  display: grid;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 2px;
}

.tag-table-head,
.tag-table-row {
  min-width: 760px;
  display: grid;
  grid-template-columns: 1.4fr repeat(5, minmax(90px, 1fr));
  align-items: center;
  gap: 10px;
}

.tag-table-head {
  font-size: 12px;
  color: #64748b;
}

.tag-table-row {
  border: 1px solid rgba(148, 163, 184, 0.24);
  border-radius: 10px;
  background: rgba(248, 250, 252, 0.72);
  padding: 9px 10px;
  font-size: 13px;
  color: #334155;
}

.tag-name {
  font-weight: 700;
  color: #0f172a;
}

.empty-tip {
  margin: 0;
  color: #64748b;
  font-size: 13px;
}

.is-dark .kpi-item,
.is-dark .priority-item,
.is-dark .insight-item,
.is-dark .tag-table-row {
  border-color: rgba(148, 163, 184, 0.22);
  background: rgba(30, 41, 59, 0.46);
}

.is-dark .kpi-item p,
.is-dark .kpi-item span,
.is-dark .priority-sub,
.is-dark .insight-title,
.is-dark .insight-sub,
.is-dark .tag-table-head,
.is-dark .meter-row span,
.is-dark .empty-tip {
  color: #cbd5e1;
}

.is-dark .kpi-item strong,
.is-dark .priority-title,
.is-dark .priority-score,
.is-dark .insight-main,
.is-dark .tag-name,
.is-dark .tag-table-row {
  color: #f1f5f9;
}

.is-dark .meter-track {
  background: rgba(148, 163, 184, 0.26);
}

@media (max-width: 1100px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 900px) {
  .stats-kpi-strip {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .priority-score {
    grid-column: 2;
    justify-self: end;
    font-size: 16px;
  }
}
</style>
