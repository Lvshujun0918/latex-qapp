<template>
  <section class="app-page app-inner-page page-wrap pt-8">
    <header class="app-page-header page-header">
      <Button variant="outline" size="icon-sm" class="app-header-back" @click="goBack" aria-label="返回上一级"><</Button>
      <span class="app-kicker">Task Status</span>
      <h1>PDF 任务</h1>
      <p>查看任务状态与进度。</p>
    </header>

    <Card class="app-page-shell">
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
import { useRouter } from 'vue-router';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

const route = useRoute();
const router = useRouter();

function goBack() {
  if (window.history.length > 1) {
    router.back();
    return;
  }

  router.replace('/pdf/export');
}
</script>

<style scoped>
.page-wrap {
  gap: 14px;
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
