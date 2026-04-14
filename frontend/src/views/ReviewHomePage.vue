<template>
    <section class="app-page app-inner-page page-wrap" :class="{ 'is-dark': resolvedTheme === 'dark' }">
        <header class="app-page-header page-header compact-header">
            <h1>复习</h1>
            <p>
                帮助家长按 {{ EBBINGHAUS_INTERVALS.join(' / ') }} 天节奏安排孩子复习，支持到期与提前复习。
            </p>
        </header>
        
        <div class="review-export-bar app-soft-card">
            <div class="review-export-copy">
                <p>今日待复习导出</p>
                <span>一键导出今日应复习题单，方便打印或分享给孩子。</span>
            </div>
            <Button size="sm" :disabled="!dueToday.length" @click="goExportDueToday">
                一键导出
            </Button>
        </div>

        <div class="overview-grid" role="list" aria-label="复习概览">
            <button type="button" class="metric-pill" @click="activeTab = 'due'" role="listitem">
                <CalendarCheck2 class="pill-icon" :size="24" />
                <div class="pill-main">
                    <p class="pill-label">今日应复习</p>
                    <p class="pill-value">{{ dueToday.length }}</p>
                </div>
            </button>

            <button type="button" class="metric-pill" @click="activeTab = 'manual'" role="listitem">
                <NotebookPen class="pill-icon" :size="24" />
                <div class="pill-main">
                    <p class="pill-label">手动复习池</p>
                    <p class="pill-value">{{ manualPool.length }}</p>
                </div>
            </button>

            <div class="metric-pill" role="listitem">
                <Brain class="pill-icon" :size="24" />
                <div class="pill-main">
                    <p class="pill-label">平均掌握度</p>
                    <p class="pill-value">{{ averageMastery }}%</p>
                </div>
            </div>

            <div class="metric-pill" role="listitem">
                <Clock3 class="pill-icon" :size="24" />
                <div class="pill-main">
                    <p class="pill-label">明日将到期</p>
                    <p class="pill-value">{{ dueTomorrow.length }}</p>
                </div>
            </div>
        </div>

        <div class="tab-switch" role="tablist" aria-label="复习列表切换">
            <button class="tab-button" :class="{ active: activeTab === 'due' }" type="button" role="tab"
                :aria-selected="activeTab === 'due'" @click="activeTab = 'due'">
                今日复习题单
            </button>
            <button class="tab-button" :class="{ active: activeTab === 'manual' }" type="button" role="tab"
                :aria-selected="activeTab === 'manual'" @click="activeTab = 'manual'">
                手动复习
            </button>
        </div>
        <div v-if="activeItems.length" class="review-list">
            <button v-for="item in activeItems" :key="`${activeTab}-${item.id}`" class="review-item" type="button"
                @click="goPractice(item.id)">
                <div class="review-main">
                    <div class="title-row">
                        <p class="review-title">{{ item.title || '未命名题目' }}</p>
                        <div class="title-badges">
                            <span class="tag-focus">{{ item.primaryTag }}</span>
                            <span class="result-badge" :class="resultClass(item.lastReviewResult)">{{
                                resultText(item.lastReviewResult) }}</span>
                        </div>
                    </div>
                    <p class="review-meta">
                        标签 {{ item.tagPreview.join(' / ') }} · {{ formatSubject(item.subject) }} · 第 {{ item.reviewCount + 1 }} 次 · 目标 {{
                            item.nextInterval }} 天
                    </p>
                    <div class="progress-row">
                        <div class="progress-track">
                            <div class="progress-fill" :class="{ due: item.overdueDays >= 0 }"
                                :style="{ width: `${item.progressPercent}%` }" />
                        </div>
                        <span class="progress-text">{{ item.progressPercent }}%</span>
                    </div>
                </div>
                <span class="review-urgency" :class="{ manual: activeTab === 'manual' }">
                    {{ activeTab === 'due' ? `到期 ${item.overdueDays} 天` : `距到期 ${Math.abs(item.overdueDays)} 天`
                    }}
                </span>
            </button>
        </div>
        <p v-else class="empty-tip">{{ activeTab === 'due' ? '今天没有到期题目，继续保持。' : '暂无可提前复习题目。' }}</p>
    </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { storeToRefs } from 'pinia';
import { useRouter } from 'vue-router';
import { Brain, CalendarCheck2, Clock3, NotebookPen } from 'lucide-vue-next';
import { useTheme } from '@/composables/useTheme';
import { Button } from '@/components/ui/button';
import { useRecordStore } from '@/stores/record';
import { buildTagSummaries, scoreRecordByTags, summarizeTags } from '@/utils/tag-analytics';

type ReviewTab = 'due' | 'manual';

const EBBINGHAUS_INTERVALS = [1, 2, 4, 7, 15, 30];

const router = useRouter();
const recordStore = useRecordStore();
const { records } = storeToRefs(recordStore);
const { resolvedTheme } = useTheme();
const activeTab = ref<ReviewTab>('due');

const tagPriorityMap = computed(() => {
    const summaries = buildTagSummaries(records.value);
    return new Map(summaries.map((item) => [item.tag, item.priorityScore]));
});

const scheduleRows = computed(() => {
    const now = Date.now();
    const priorityMap = tagPriorityMap.value;
    return records.value
        .map((record) => {
            const reviewCount = Math.max(0, Number(record.reviewCount || 0));
            const intervalIndex = Math.min(reviewCount, EBBINGHAUS_INTERVALS.length - 1);
            const baseInterval = EBBINGHAUS_INTERVALS[intervalIndex];
            const easeFactor = Math.max(1.3, Math.min(3.0, Number(record.reviewEaseFactor || 2.5)));
            const nextInterval = Math.max(1, Math.round(baseInterval * easeFactor * 0.55));

            const createdTime = Date.parse(record.createdAt || '');
            const lastReviewedTime = Date.parse(record.lastReviewedAt || '');
            const nextReviewTime = Date.parse(record.nextReviewAt || '');
            const base = Number.isFinite(lastReviewedTime)
                ? lastReviewedTime
                : Number.isFinite(createdTime)
                    ? createdTime
                    : now;

            const elapsedDays = Math.max(0, Math.floor((now - base) / 86400000));
            const overdueDays = Number.isFinite(nextReviewTime)
                ? Math.floor((now - nextReviewTime) / 86400000)
                : elapsedDays - nextInterval;
            const progressPercent = Number.isFinite(nextReviewTime)
                ? Math.max(0, Math.min(100, Math.round((1 - Math.max(0, nextReviewTime - now) / (nextInterval * 86400000)) * 100)))
                : Math.max(0, Math.min(100, Math.round((elapsedDays / nextInterval) * 100)));

            const tagScore = scoreRecordByTags(record, priorityMap);

            return {
                ...record,
                nextInterval,
                overdueDays,
                progressPercent,
                primaryTag: tagScore.primaryTag,
                tagPriorityScore: tagScore.score,
                tagPreview: summarizeTags(record.questionTags, 3),
            };
        })
        .sort((a, b) => b.overdueDays - a.overdueDays || b.tagPriorityScore - a.tagPriorityScore);
});

const dueToday = computed(() => scheduleRows.value.filter((row) => row.overdueDays >= 0));
const manualPool = computed(() =>
    scheduleRows.value
        .filter((row) => row.overdueDays < 0)
    .sort((a, b) => b.tagPriorityScore - a.tagPriorityScore || b.progressPercent - a.progressPercent || a.masteryLevel - b.masteryLevel)
        .slice(0, 16),
);
const activeItems = computed(() => (activeTab.value === 'due' ? dueToday.value : manualPool.value));
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

function goPractice(id: number) {
    router.push(`/review/session/${id}`);
}

function goExportDueToday() {
    if (!dueToday.value.length) {
        return;
    }

    const ids = dueToday.value.map((item) => item.id).join(',');
    router.push({
        path: '/tabs/pdfs',
        query: {
            prefill: ids,
        },
    });
}

function formatSubject(subject?: string) {
    const value = String(subject || '').trim().toLowerCase();
    if (value === 'math' || value === '数学') return '数学';
    if (value === 'physics' || value === '物理') return '物理';
    if (value === 'chemistry' || value === '化学') return '化学';
    if (value === 'biology' || value === '生物') return '生物';
    return subject || '未知';
}

function resultText(result?: string) {
    if (result === 'correct') return '上次正确';
    if (result === 'wrong') return '上次错误';
    return '未判定';
}

function resultClass(result?: string) {
    if (result === 'correct') return 'ok';
    if (result === 'wrong') return 'bad';
    return 'none';
}
</script>

<style scoped>
.page-wrap {
    gap: 12px;
}

.compact-header :deep(p) {
    margin-top: 4px;
}

.review-export-bar {
    border: 1px solid rgba(148, 163, 184, 0.24);
    border-radius: 14px;
    padding: 12px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
}

.review-export-copy {
    min-width: 0;
}

.review-export-copy p {
    margin: 0;
    font-size: 14px;
    font-weight: 700;
    color: #0f172a;
}

.review-export-copy span {
    margin-top: 2px;
    display: block;
    font-size: 12px;
    color: #64748b;
}

.overview-grid {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 10px;
}

.metric-pill {
    border: 1px solid rgba(148, 163, 184, 0.3);
    border-radius: 12px;
    background: #fff;
    min-height: 68px;
    padding: 10px 12px;
    display: flex;
    align-items: center;
    gap: 10px;
    text-align: left;
}

button.metric-pill {
    cursor: pointer;
    transition: border-color 150ms ease, transform 150ms ease;
}

button.metric-pill:hover {
    transform: translateY(-1px);
    border-color: rgba(14, 165, 233, 0.45);
}

.pill-icon {
    color: #0369a1;
    flex: 0 0 auto;
}

.pill-main {
    min-width: 0;
}

.pill-label {
    margin: 0;
    font-size: 11px;
    color: #64748b;
}

.pill-value {
    margin: 2px 0 0;
    font-size: 20px;
    line-height: 1;
    font-weight: 700;
    color: #0f172a;
}

.list-shell {
    margin-top: 2px;
}

.list-header {
    display: flex;
    flex-direction: column;
    gap: 10px;
}

.tab-switch {
    display: inline-flex;
    width: fit-content;
    padding: 3px;
    border-radius: 999px;
    border: 1px solid rgba(148, 163, 184, 0.28);
    background: rgba(248, 250, 252, 0.9);
}

.tab-button {
    border: none;
    background: transparent;
    border-radius: 999px;
    font-size: 12px;
    color: #64748b;
    padding: 6px 12px;
    cursor: pointer;
}

.tab-button.active {
    color: #fff;
    background: #2563eb;
}

.list-content {
    padding-top: 8px;
}

.review-list {
    display: grid;
    gap: 8px;
}

.review-item {
    width: 100%;
    border: 1px solid rgba(148, 163, 184, 0.26);
    border-radius: 12px;
    background: #fff;
    padding: 10px 12px;
    text-align: left;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    cursor: pointer;
}

.review-main {
    display: grid;
    gap: 5px;
    min-width: 0;
    flex: 1;
    width: 100%;
}

.title-row {
    display: flex;
    align-items: center;
    gap: 8px;
    justify-content: space-between;
}

.title-badges {
    display: inline-flex;
    align-items: center;
    gap: 6px;
}

.tag-focus {
    font-size: 10px;
    border-radius: 999px;
    padding: 2px 8px;
    color: #1d4ed8;
    background: rgba(37, 99, 235, 0.14);
    border: 1px solid rgba(37, 99, 235, 0.28);
    white-space: nowrap;
}

.review-title {
    margin: 0;
    font-size: 14px;
    font-weight: 600;
    color: #0f172a;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.result-badge {
    font-size: 10px;
    border-radius: 999px;
    padding: 2px 8px;
    border: 1px solid transparent;
    white-space: nowrap;
}

.result-badge.ok {
    color: #065f46;
    background: rgba(16, 185, 129, 0.15);
    border-color: rgba(16, 185, 129, 0.28);
}

.result-badge.bad {
    color: #991b1b;
    background: rgba(248, 113, 113, 0.16);
    border-color: rgba(248, 113, 113, 0.3);
}

.result-badge.none {
    color: #475569;
    background: rgba(148, 163, 184, 0.14);
    border-color: rgba(148, 163, 184, 0.26);
}

.review-meta {
    margin: 0;
    font-size: 12px;
    color: #64748b;
}

.progress-row {
    display: flex;
    align-items: center;
    gap: 8px;
}

.progress-track {
    position: relative;
    flex: 1;
    height: 8px;
    border-radius: 999px;
    overflow: hidden;
    background: rgba(148, 163, 184, 0.22);
}

.progress-fill {
    height: 100%;
    border-radius: 999px;
    background: #2563eb;
}

.progress-fill.due {
    background: #f59e0b;
}

.progress-text {
    width: 42px;
    text-align: right;
    font-size: 11px;
    color: #64748b;
}

.review-urgency {
    font-size: 12px;
    color: #b45309;
    white-space: nowrap;
}

.review-urgency.manual {
    color: #0f766e;
}

.empty-tip {
    margin: 0;
    color: #64748b;
    font-size: 13px;
}

.is-dark .metric-pill,
.is-dark .review-item,
.is-dark .review-export-bar {
    background: rgba(30, 41, 59, 0.94);
    border-color: rgba(148, 163, 184, 0.28);
}

.is-dark .pill-label,
.is-dark .review-meta,
.is-dark .progress-text,
.is-dark .empty-tip,
.is-dark .review-export-copy span {
    color: #cbd5e1;
}

.is-dark .pill-value,
.is-dark .review-title,
.is-dark .review-export-copy p {
    color: #f8fafc;
}

.is-dark .pill-icon {
    color: #7dd3fc;
}

.is-dark .tab-switch {
    border-color: rgba(148, 163, 184, 0.3);
    background: rgba(15, 23, 42, 0.55);
}

.is-dark .tab-button {
    color: #cbd5e1;
}

.is-dark .progress-track {
    background: rgba(148, 163, 184, 0.28);
}

.is-dark .result-badge.none {
    color: #cbd5e1;
}

.is-dark .tag-focus {
    color: #dbeafe;
    background: rgba(37, 99, 235, 0.28);
    border-color: rgba(96, 165, 250, 0.44);
}

.is-dark .review-urgency.manual {
    color: #5eead4;
}

@media (max-width: 960px) {
    .overview-grid {
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .review-item {
        align-items: flex-start;
        flex-direction: column;
    }
}

@media (max-width: 640px) {
    .tab-switch {
        width: 100%;
    }

    .tab-button {
        flex: 1;
    }
}
</style>
