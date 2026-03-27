<template>
  <section class="app-page page-wrap">
    <header class="app-page-header page-header">
      <span class="app-kicker">Daily Focus</span>
      <h1>复习</h1>
      <p>围绕高频错因与掌握度安排每日复习。</p>
    </header>

    <Card class="app-soft-card">
      <CardHeader>
        <CardTitle>今日推荐</CardTitle>
      </CardHeader>
      <CardContent>根据错题频次和掌握度，系统将推荐重点复习题目。</CardContent>
    </Card>

    <Card class="app-soft-card">
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
  gap: 14px;
}
</style>
