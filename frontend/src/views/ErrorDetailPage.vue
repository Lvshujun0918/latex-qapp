<template>
  <section class="app-page page-wrap" v-if="record">
    <header class="app-page-header page-header">
      <span class="app-kicker">Error Insight</span>
      <h1>错题详情</h1>
      <p>查看题目元信息并继续进入 AI 解析。</p>
    </header>

    <Card class="app-page-shell detail-shell">
      <CardHeader class="detail-header p-0 pb-3">
        <CardTitle>{{ record.title || '未命名题目' }}</CardTitle>
      </CardHeader>
      <CardContent class="detail-content p-0">
        <div class="app-meta-grid">
          <p class="app-meta-item">学科：{{ record.subject }}</p>
          <p class="app-meta-item">题型：{{ record.questionType || 'unknown' }}</p>
          <p class="app-meta-item">同步状态：{{ record.syncStatus }}</p>
        </div>
        <LatexView :source="record.latexSource || ''" class="latex-block" />
      </CardContent>
    </Card>

    <Button class="action-btn" @click="toAnalysis">进入 AI 深度解析</Button>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useRecordStore } from '@/stores/record';
import LatexView from '@/components/LatexView.vue';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

const route = useRoute();
const router = useRouter();
const recordStore = useRecordStore();

const record = computed(() => recordStore.records.find((r) => r.id === Number(route.params.id)));

onMounted(() => {
  recordStore.reload();
});

function toAnalysis() {
  router.push(`/records/${route.params.id}/analysis`);
}
</script>

<style scoped>
.page-wrap {
  gap: 14px;
}

.detail-shell {
  overflow: hidden;
}

.detail-header {
  border-bottom: 1px solid rgba(148, 163, 184, 0.24);
}

.detail-content {
  display: grid;
  gap: 12px;
}

.action-btn {
  width: 100%;
  height: 44px;
  font-weight: 700;
}
</style>
