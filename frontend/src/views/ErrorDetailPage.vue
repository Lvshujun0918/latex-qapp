<template>
  <section class="app-page app-inner-page page-wrap" v-if="record">
    <Button variant="outline" size="sm" class="back-btn" @click="goBack">返回上一级</Button>

    <header class="app-page-header page-header">
      <span class="app-kicker">Error Insight</span>
      <h1>错题详情</h1>
      <p>查看题目元信息与 AI 解析结果。</p>
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

    <Card class="app-page-shell detail-shell">
      <CardHeader class="detail-header p-0 pb-3">
        <CardTitle>AI 解析</CardTitle>
      </CardHeader>
      <CardContent class="detail-content p-0">
        <div class="analysis-block">
          <h4>最终答案</h4>
          <LatexView :source="record.latexAnswer || '暂无答案'" class="latex-block" />
        </div>

        <div class="analysis-block">
          <h4>分步解答 / 错因分析</h4>
          <LatexView :source="record.mistakeReason || '暂无 AI 解析结果，请先在录入页生成解答。'" class="latex-block" />
        </div>
      </CardContent>
    </Card>
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

function goBack() {
  if (window.history.length > 1) {
    router.back();
    return;
  }

  router.replace('/tabs/errors');
}
</script>

<style scoped>
.page-wrap {
  gap: 14px;
}

.back-btn {
  justify-self: start;
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

.analysis-block {
  display: grid;
  gap: 8px;
}

.analysis-block h4 {
  margin: 0;
  font-size: 13px;
  color: #64748b;
}
</style>
