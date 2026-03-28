<template>
  <section class="app-page app-inner-page page-wrap">
    <Button variant="outline" size="sm" class="back-btn" @click="goBack">返回上一级</Button>

    <header class="app-page-header page-header">
      <span class="app-kicker">AI Reasoning</span>
      <h1>AI 解析</h1>
      <p>展示分步推理、知识点和复习建议。</p>
    </header>

    <Card class="app-page-shell">
      <CardHeader>
        <CardTitle>解析结果</CardTitle>
      </CardHeader>
      <CardContent>
        <p>待接入后端接口后展示详细步骤、知识点和复习建议。</p>
      </CardContent>
    </Card>
  </section>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

const route = useRoute();
const router = useRouter();

function goBack() {
  if (window.history.length > 1) {
    router.back();
    return;
  }

  router.replace(`/records/${route.params.id}`);
}
</script>

<style scoped>
.page-wrap {
  gap: 14px;
}

.back-btn {
  justify-self: start;
}

p {
  margin: 0;
  color: #475569;
}
</style>
