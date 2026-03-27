<template>
  <section class="page-wrap">
    <header class="page-header">
      <h1>统计</h1>
      <p>关注总量变化、学科分布和趋势演进。</p>
    </header>

    <Card>
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
