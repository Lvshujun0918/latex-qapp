<template>
  <section class="app-page page-wrap">
    <header class="app-page-header page-header">
      <span class="app-kicker">Progress Pulse</span>
      <h1>统计</h1>
      <p>关注总量变化、学科分布和趋势演进。</p>
    </header>

    <Card class="app-page-shell">
      <CardHeader>
        <CardTitle>学习统计</CardTitle>
      </CardHeader>
      <CardContent>这里将展示总错题数、学科分布和掌握度趋势。</CardContent>
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
