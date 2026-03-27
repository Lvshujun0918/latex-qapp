<template>
  <section class="page-wrap">
    <header class="page-header">
      <h1>PDF 任务</h1>
      <p>查看任务状态与进度。</p>
    </header>

    <Card>
      <CardHeader>
        <CardTitle>任务号：{{ route.params.jobId }}</CardTitle>
      </CardHeader>
      <CardContent class="job-content">
        <p>状态：queued</p>
        <p>进度：0%</p>
        <p>后续接入后端后展示实时状态。</p>
      </CardContent>
    </Card>
  </section>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

const route = useRoute();
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

.job-content {
  display: grid;
  gap: 6px;
}

.job-content p {
  margin: 0;
  color: #475569;
}
</style>
