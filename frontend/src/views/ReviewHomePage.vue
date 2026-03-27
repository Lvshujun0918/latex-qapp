<template>
  <section class="page-wrap">
    <header class="page-header">
      <h1>复习</h1>
      <p>围绕高频错因与掌握度安排每日复习。</p>
    </header>

    <Card>
      <CardHeader>
        <CardTitle>今日推荐</CardTitle>
      </CardHeader>
      <CardContent>根据错题频次和掌握度，系统将推荐重点复习题目。</CardContent>
    </Card>

    <Card>
      <CardHeader>
        <CardTitle>高频错因</CardTitle>
      </CardHeader>
      <CardContent>后续联调后展示来自 AI 的易错点聚合。</CardContent>
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
</style>
