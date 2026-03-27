<template>
  <section class="page-wrap">
    <header class="page-header">
      <h1>AI 解析</h1>
      <p>展示分步推理、知识点和复习建议。</p>
    </header>

    <Card>
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
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
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

p {
  margin: 0;
  color: #475569;
}
</style>
