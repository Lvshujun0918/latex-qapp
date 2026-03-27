<template>
  <section class="page-wrap" v-if="record">
    <header class="page-header">
      <h1>错题详情</h1>
      <p>查看题目元信息并继续进入 AI 解析。</p>
    </header>

    <Card>
      <CardHeader>
        <CardTitle>{{ record.title || '未命名题目' }}</CardTitle>
      </CardHeader>
      <CardContent class="detail-content">
        <p>学科：{{ record.subject }}</p>
        <p>题型：{{ record.questionType || 'unknown' }}</p>
        <p>同步状态：{{ record.syncStatus }}</p>
        <LatexView :source="record.latexSource" class="latex-block" />
      </CardContent>
    </Card>

    <Button class="action-btn" @click="toAnalysis">AI 解析</Button>
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
  max-width: 960px;
  margin: 0 auto;
  display: grid;
  gap: 14px;
}

.page-header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
}

.page-header p {
  margin: 4px 0 0;
  color: #64748b;
  font-size: 14px;
}

.detail-content {
  display: grid;
  gap: 8px;
}

.detail-content p {
  margin: 0;
  color: #475569;
}

.latex-block {
  margin-top: 6px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.52);
  border: 1px solid rgba(148, 163, 184, 0.36);
  padding: 10px 12px;
}

.action-btn {
  width: 100%;
}
</style>
