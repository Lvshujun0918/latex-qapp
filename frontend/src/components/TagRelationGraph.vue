<template>
  <section class="tag-graph-shell app-soft-card" :class="{ 'is-dark': theme === 'dark' }">
    <div class="graph-header">
      <div class="graph-copy">
        <p class="graph-kicker">标签关系图谱</p>
        <h3>看标签如何成簇、互相关联、决定复习优先级</h3>
        <p>支持拖拽、缩放、节点点击，帮助把题目管理切换成标签管理。</p>
      </div>

      <div class="graph-toolbar" role="tablist" aria-label="标签图谱视图切换">
        <button
          v-for="mode in graphModes"
          :key="mode.key"
          type="button"
          class="graph-mode-btn"
          :class="{ active: graphMode === mode.key }"
          @click="graphMode = mode.key"
        >
          {{ mode.label }}
        </button>
      </div>
    </div>

    <div v-if="hasData" class="graph-stage">
      <div ref="graphEl" class="graph-canvas" />

      <aside class="graph-sidebar">
        <p class="sidebar-label">当前聚焦标签</p>
        <template v-if="activeSummary">
          <div class="sidebar-title-row">
            <h4>{{ activeSummary.tag }}</h4>
            <span class="priority-pill">优先级 {{ activeSummary.priorityScore }}</span>
          </div>

          <div class="sidebar-metrics">
            <div>
              <span>题目数</span>
              <strong>{{ activeSummary.recordCount }}</strong>
            </div>
            <div>
              <span>正确率</span>
              <strong>{{ activeSummary.correctRate }}%</strong>
            </div>
            <div>
              <span>错误率</span>
              <strong>{{ activeSummary.wrongRate }}%</strong>
            </div>
            <div>
              <span>复习频次</span>
              <strong>{{ activeSummary.avgReviewCount }}</strong>
            </div>
          </div>

          <div class="sidebar-section">
            <p class="sidebar-subtitle">相关标签</p>
            <div v-if="relatedTags.length" class="chip-list">
              <button
                v-for="item in relatedTags"
                :key="item.tag"
                type="button"
                class="relation-chip"
                @click="focusTag(item.tag)"
              >
                <span>{{ item.tag }}</span>
                <small>{{ item.weight }}</small>
              </button>
            </div>
            <p v-else class="sidebar-empty">暂无强关联标签。</p>
          </div>
        </template>
        <p v-else class="sidebar-empty">暂无标签数据。</p>
      </aside>
    </div>

    <div v-else class="graph-empty">
      <p>暂无标签关系数据，先为题目补充标签吧。</p>
    </div>

    <div v-if="hasData" class="graph-footer">
      <div class="graph-legend">
        <span><i class="legend-dot legend-hot" />高优先级</span>
        <span><i class="legend-dot legend-mid" />中优先级</span>
        <span><i class="legend-dot legend-calm" />稳定标签</span>
      </div>
      <p class="graph-tip">提示：拖动节点可以查看聚类结构，点击节点会同步右侧详情。</p>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import * as echarts from 'echarts';
import type { ErrorRecord } from '@/types/domain';
import { buildTagRelations, buildTagSummaries, type TagSummary } from '@/utils/tag-analytics';

interface Props {
  records: ErrorRecord[];
  theme: 'light' | 'dark';
}

type GraphMode = 'priority' | 'frequency' | 'mastery';

interface GraphModeItem {
  key: GraphMode;
  label: string;
}

interface GraphNodeView {
  tag: string;
  recordCount: number;
  correctRate: number;
  wrongRate: number;
  avgReviewCount: number;
  avgMastery: number;
  priorityScore: number;
  weight: number;
  category: number;
  symbolSize: number;
  color: string;
}

interface GraphEdgeView {
  source: string;
  target: string;
  weight: number;
}

const props = defineProps<Props>();

const graphEl = ref<HTMLDivElement | null>(null);
const graphChart = ref<echarts.ECharts | null>(null);
const graphMode = ref<GraphMode>('priority');
const activeTag = ref('');

const graphModes: GraphModeItem[] = [
  { key: 'priority', label: '按优先级' },
  { key: 'frequency', label: '按频次' },
  { key: 'mastery', label: '按掌握度' },
];

const hasData = computed(() => tagSummaries.value.length > 0);
const tagSummaries = computed(() => buildTagSummaries(props.records));

const modeSortedSummaries = computed<TagSummary[]>(() => {
  const list = [...tagSummaries.value];
  if (graphMode.value === 'frequency') {
    return list.sort((a, b) => b.recordCount - a.recordCount || b.priorityScore - a.priorityScore || a.tag.localeCompare(b.tag));
  }
  if (graphMode.value === 'mastery') {
    return list.sort((a, b) => a.avgMastery - b.avgMastery || b.priorityScore - a.priorityScore || a.tag.localeCompare(b.tag));
  }
  return list.sort((a, b) => b.priorityScore - a.priorityScore || b.recordCount - a.recordCount || a.tag.localeCompare(b.tag));
});

const visibleSummaries = computed(() => {
  const limit = Math.min(14, modeSortedSummaries.value.length);
  const picked = modeSortedSummaries.value.slice(0, limit);
  if (activeTag.value && !picked.some((item) => item.tag === activeTag.value)) {
    const selected = tagSummaries.value.find((item) => item.tag === activeTag.value);
    if (selected) {
      picked[picked.length - 1] = selected;
    }
  }
  return picked;
});

const visibleTagSet = computed(() => new Set(visibleSummaries.value.map((item) => item.tag)));

const visibleRelations = computed<GraphEdgeView[]>(() => {
  const edges = buildTagRelations(props.records, 36, visibleTagSet.value);
  return edges.map((edge) => ({ source: edge.source, target: edge.target, weight: edge.weight }));
});

const activeSummary = computed(() => {
  if (!tagSummaries.value.length) {
    return null;
  }
  const fromTag = tagSummaries.value.find((item) => item.tag === activeTag.value);
  return fromTag || visibleSummaries.value[0] || tagSummaries.value[0];
});

const relatedTags = computed(() => {
  const current = activeSummary.value?.tag;
  if (!current) {
    return [] as Array<{ tag: string; weight: number }>;
  }

  return visibleRelations.value
    .filter((edge) => edge.source === current || edge.target === current)
    .map((edge) => ({ tag: edge.source === current ? edge.target : edge.source, weight: edge.weight }))
    .sort((a, b) => b.weight - a.weight || a.tag.localeCompare(b.tag))
    .slice(0, 6);
});

const graphData = computed(() => {
  const topScore = Math.max(1, ...visibleSummaries.value.map((item) => item.priorityScore));
  return visibleSummaries.value.map((item) => {
    const category = item.priorityScore >= topScore * 0.72 ? 0 : item.priorityScore >= topScore * 0.42 ? 1 : 2;
    const baseSize = graphMode.value === 'mastery'
      ? 18 + Math.max(0, 100 - item.avgMastery) * 0.18
      : graphMode.value === 'frequency'
        ? 16 + item.recordCount * 1.7
        : 16 + item.priorityScore * 0.12;
    const symbolSize = Math.max(16, Math.min(42, Math.round(baseSize)));

    return {
      tag: item.tag,
      recordCount: item.recordCount,
      correctRate: item.correctRate,
      wrongRate: item.wrongRate,
      avgReviewCount: item.avgReviewCount,
      avgMastery: item.avgMastery,
      priorityScore: item.priorityScore,
      category,
      symbolSize,
      weight: item.priorityScore,
      color: nodeColor(category),
    } satisfies GraphNodeView;
  });
});

const graphOption = computed<echarts.EChartsCoreOption>(() => {
  const categories = [
    { name: '高优先级', itemStyle: { color: '#f97316' } },
    { name: '中优先级', itemStyle: { color: '#2563eb' } },
    { name: '稳定标签', itemStyle: { color: '#0f766e' } },
  ];

  return {
    backgroundColor: 'transparent',
    animationDuration: 450,
    tooltip: {
      trigger: 'item',
      backgroundColor: props.theme === 'dark' ? 'rgba(15, 23, 42, 0.96)' : 'rgba(255, 255, 255, 0.96)',
      borderColor: props.theme === 'dark' ? 'rgba(148, 163, 184, 0.2)' : 'rgba(148, 163, 184, 0.28)',
      textStyle: {
        color: props.theme === 'dark' ? '#e2e8f0' : '#0f172a',
      },
      formatter: (params: any) => {
        if (params.dataType === 'edge') {
          return `${params.data.source} ↔ ${params.data.target}<br/>共现次数：${params.data.value}`;
        }
        const data = params.data as GraphNodeView;
        return [
          `<strong>${data.tag}</strong>`,
          `题目数：${data.recordCount}`,
          `正确率：${data.correctRate}%`,
          `错误率：${data.wrongRate}%`,
          `复习频次：${data.avgReviewCount}`,
          `优先级：${data.priorityScore}`,
        ].join('<br/>');
      },
    },
    legend: {
      top: 8,
      right: 8,
      itemWidth: 12,
      itemHeight: 12,
      textStyle: {
        color: props.theme === 'dark' ? '#cbd5e1' : '#475569',
      },
      data: categories.map((item) => item.name),
    },
    series: [
      {
        name: '标签关系',
        type: 'graph',
        layout: 'force',
        roam: true,
        draggable: true,
        categories,
        data: graphData.value.map((item) => ({
          id: item.tag,
          name: item.tag,
          value: item.recordCount,
          category: item.category,
          symbolSize: item.symbolSize,
          draggable: true,
          itemStyle: {
            color: item.color,
          },
          label: {
            show: true,
            color: props.theme === 'dark' ? '#f8fafc' : '#0f172a',
          },
          emphasis: {
            focus: 'adjacency',
            scale: true,
          },
        })),
        links: visibleRelations.value.map((edge) => ({
          source: edge.source,
          target: edge.target,
          value: edge.weight,
          lineStyle: {
            width: Math.max(1, Math.min(7, 1 + edge.weight * 0.9)),
            opacity: 0.4,
            curveness: 0.18,
            color: props.theme === 'dark' ? 'rgba(148, 163, 184, 0.52)' : 'rgba(100, 116, 139, 0.46)',
          },
        })),
        force: {
          repulsion: 260,
          edgeLength: [68, 120],
          gravity: 0.03,
          friction: 0.2,
        },
        emphasis: {
          focus: 'adjacency',
          lineStyle: {
            width: 4,
          },
        },
        label: {
          position: 'right',
          formatter: '{b}',
        },
        edgeSymbol: ['none', 'none'],
        lineStyle: {
          opacity: 0.38,
        },
        tooltip: {
          show: true,
        },
      },
    ],
    graphic: graphData.value.length
      ? []
      : [
          {
            type: 'text',
            left: 'center',
            top: 'middle',
            style: {
              text: '暂无标签关系数据',
              fill: props.theme === 'dark' ? '#cbd5e1' : '#64748b',
              fontSize: 14,
            },
          },
        ],
  };
});

watch(
  () => graphData.value,
  () => {
    if (!activeTag.value && activeSummary.value) {
      activeTag.value = activeSummary.value.tag;
    }
    renderChart();
  },
  { deep: true, immediate: true },
);

watch(
  () => graphMode.value,
  () => {
    if (activeSummary.value) {
      activeTag.value = activeSummary.value.tag;
    }
    renderChart();
  },
);

watch(
  () => props.theme,
  () => {
    renderChart();
  },
);

onMounted(async () => {
  await nextTick();
  initChart();
  renderChart();
  window.addEventListener('resize', handleResize);
});

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize);
  graphChart.value?.dispose();
  graphChart.value = null;
});

function initChart() {
  if (!graphEl.value || graphChart.value) {
    return;
  }
  graphChart.value = echarts.init(graphEl.value);
  graphChart.value.on('click', (params: any) => {
    if (params?.dataType !== 'node' || !params?.data?.name) {
      return;
    }
    focusTag(String(params.data.name));
  });
}

function renderChart() {
  if (!graphChart.value) {
    initChart();
  }
  if (!graphChart.value) {
    return;
  }
  graphChart.value.setOption(graphOption.value, true);
}

function handleResize() {
  graphChart.value?.resize();
}

function focusTag(tag: string) {
  activeTag.value = tag;
  if (!graphChart.value) {
    return;
  }
  graphChart.value.dispatchAction({ type: 'highlight', name: tag });
}

function nodeColor(category: number) {
  if (category === 0) {
    return '#f97316';
  }
  if (category === 1) {
    return '#2563eb';
  }
  return '#0f766e';
}
</script>

<style scoped>
.tag-graph-shell {
  --panel-bg: rgba(248, 250, 252, 0.78);
  --panel-border: rgba(148, 163, 184, 0.24);
  --text-primary: #0f172a;
  --text-secondary: #64748b;
  --chip-bg: #ffffff;
  --canvas-bg: radial-gradient(circle at 15% 16%, rgba(14, 165, 233, 0.1), transparent 42%),
    radial-gradient(circle at 85% 12%, rgba(34, 197, 94, 0.08), transparent 36%),
    linear-gradient(180deg, rgba(248, 250, 252, 0.9), rgba(241, 245, 249, 0.7));
  display: grid;
  gap: 16px;
  border-radius: 22px;
  border: 1px solid var(--panel-border);
  background: var(--panel-bg);
  padding: 14px;
}

.graph-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 14px;
}

.graph-copy {
  min-width: 0;
}

.graph-kicker {
  margin: 0;
  font-size: 12px;
  font-weight: 700;
  color: #0369a1;
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.graph-copy h3 {
  margin: 7px 0 0;
  font-size: 20px;
  line-height: 1.2;
  color: var(--text-primary);
}

.graph-copy p:last-child {
  margin: 6px 0 0;
  font-size: 13px;
  line-height: 1.5;
  color: var(--text-secondary);
  max-width: 620px;
}

.graph-toolbar {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 4px;
  border-radius: 999px;
  border: 1px solid var(--panel-border);
  background: rgba(255, 255, 255, 0.84);
}

.graph-mode-btn {
  border: 1px solid transparent;
  border-radius: 999px;
  background: transparent;
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: 600;
  padding: 7px 14px;
  cursor: pointer;
  transition: background-color 0.18s ease, color 0.18s ease, transform 0.18s ease, border-color 0.18s ease;
}

.graph-mode-btn:hover {
  transform: translateY(-1px);
  border-color: rgba(148, 163, 184, 0.32);
}

.graph-mode-btn.active {
  border-color: rgba(29, 78, 216, 0.4);
  background: linear-gradient(135deg, #1d4ed8, #2563eb);
  color: #fff;
}

.graph-stage {
  display: grid;
  grid-template-columns: minmax(0, 1.65fr) minmax(270px, 0.85fr);
  gap: 12px;
  align-items: stretch;
}

.graph-canvas {
  min-height: 440px;
  border-radius: 18px;
  border: 1px solid var(--panel-border);
  background: var(--canvas-bg);
}

.graph-sidebar {
  border-radius: 18px;
  border: 1px solid var(--panel-border);
  background: color-mix(in srgb, var(--panel-bg) 88%, white 12%);
  padding: 14px;
  display: grid;
  gap: 12px;
  align-content: start;
}

.sidebar-label,
.sidebar-subtitle {
  margin: 0;
  font-size: 12px;
  color: var(--text-secondary);
}

.sidebar-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.sidebar-title-row h4 {
  margin: 0;
  font-size: 20px;
  color: var(--text-primary);
  max-width: 56%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.priority-pill {
  border-radius: 999px;
  border: 1px solid rgba(30, 64, 175, 0.26);
  background: rgba(37, 99, 235, 0.14);
  color: #1d4ed8;
  font-size: 11px;
  font-weight: 700;
  padding: 5px 10px;
  white-space: nowrap;
}

.sidebar-metrics {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

.sidebar-metrics div {
  border-radius: 12px;
  border: 1px solid var(--panel-border);
  background: var(--chip-bg);
  padding: 10px;
  display: grid;
  gap: 2px;
}

.sidebar-metrics span {
  font-size: 11px;
  color: var(--text-secondary);
}

.sidebar-metrics strong {
  font-size: 17px;
  color: var(--text-primary);
}

.sidebar-section {
  display: grid;
  gap: 8px;
}

.chip-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.relation-chip {
  border: 1px solid var(--panel-border);
  border-radius: 999px;
  background: var(--chip-bg);
  padding: 6px 10px;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: var(--text-primary);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.16s ease, border-color 0.16s ease, background-color 0.16s ease;
}

.relation-chip:hover {
  transform: translateY(-1px);
  border-color: rgba(37, 99, 235, 0.4);
  background: rgba(37, 99, 235, 0.08);
}

.relation-chip small {
  color: var(--text-secondary);
}

.sidebar-empty {
  margin: 0;
  font-size: 13px;
  color: var(--text-secondary);
}

.graph-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.graph-legend {
  display: flex;
  gap: 14px;
  flex-wrap: wrap;
  font-size: 12px;
  color: var(--text-secondary);
}

.legend-dot {
  width: 10px;
  height: 10px;
  border-radius: 999px;
  display: inline-block;
  margin-right: 6px;
  transform: translateY(1px);
}

.legend-hot {
  background: #f97316;
}

.legend-mid {
  background: #2563eb;
}

.legend-calm {
  background: #0f766e;
}

.graph-tip {
  margin: 0;
  font-size: 12px;
  color: var(--text-secondary);
}

.graph-empty {
  min-height: 220px;
  border-radius: 18px;
  border: 1px dashed rgba(148, 163, 184, 0.45);
  background: rgba(248, 250, 252, 0.62);
  display: grid;
  place-items: center;
  color: var(--text-secondary);
}

.is-dark.tag-graph-shell {
  --panel-bg: rgba(15, 23, 42, 0.62);
  --panel-border: rgba(148, 163, 184, 0.24);
  --text-primary: #f8fafc;
  --text-secondary: #cbd5e1;
  --chip-bg: rgba(30, 41, 59, 0.72);
  --canvas-bg: radial-gradient(circle at 16% 14%, rgba(56, 189, 248, 0.16), transparent 42%),
    radial-gradient(circle at 84% 10%, rgba(16, 185, 129, 0.14), transparent 36%),
    linear-gradient(180deg, rgba(15, 23, 42, 0.94), rgba(15, 23, 42, 0.78));
}

.is-dark .graph-kicker {
  color: #7dd3fc;
}

.is-dark .graph-toolbar {
  background: rgba(30, 41, 59, 0.76);
}

.is-dark .graph-empty {
  background: rgba(15, 23, 42, 0.58);
}

.is-dark .graph-mode-btn {
  color: #dbeafe;
}

.is-dark .graph-mode-btn.active {
  border-color: rgba(96, 165, 250, 0.46);
  color: #eff6ff;
}

.is-dark .priority-pill {
  background: rgba(37, 99, 235, 0.28);
  border-color: rgba(96, 165, 250, 0.42);
  color: #dbeafe;
}

.is-dark .relation-chip:hover {
  border-color: rgba(147, 197, 253, 0.42);
  background: rgba(59, 130, 246, 0.2);
}

@media (max-width: 960px) {
  .tag-graph-shell {
    padding: 12px;
  }

  .graph-stage {
    grid-template-columns: 1fr;
  }

  .graph-canvas {
    min-height: 340px;
  }

  .graph-sidebar {
    padding: 12px;
  }
}

@media (max-width: 640px) {
  .graph-header {
    flex-direction: column;
    gap: 10px;
  }

  .graph-copy h3 {
    font-size: 18px;
  }

  .sidebar-metrics {
    grid-template-columns: 1fr;
  }

  .graph-canvas {
    min-height: 300px;
  }

  .graph-toolbar {
    width: 100%;
    justify-content: space-between;
  }

  .graph-mode-btn {
    flex: 1;
    padding: 8px 6px;
  }

  .sidebar-title-row {
    flex-wrap: wrap;
  }

  .sidebar-title-row h4 {
    max-width: 100%;
    white-space: normal;
  }
}
</style>
