<template>
  <section class="app-page page-wrap">
    <header class="app-page-header page-header">
      <span class="app-kicker">Export Center</span>
      <h1>导出错题本 PDF</h1>
      <p>按学科与难度创建导出任务。</p>
    </header>

    <Card class="app-page-shell">
      <CardContent class="form-wrap">
        <div class="field">
          <Label for="subject">学科</Label>
          <Input id="subject" v-model="subject" />
        </div>

        <div class="field">
          <Label for="difficulty">难度</Label>
          <Input id="difficulty" v-model.number="difficulty" type="number" min="1" max="5" />
        </div>

        <Button @click="submit">提交导出任务</Button>
      </CardContent>
    </Card>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

const router = useRouter();
const subject = ref('math');
const difficulty = ref(3);

function submit() {
  router.push('/pdf/jobs/mock-job-1');
}
</script>

<style scoped>
.page-wrap {
  gap: 14px;
}

.form-wrap {
  display: grid;
  gap: 12px;
}

.field {
  display: grid;
  gap: 8px;
}
</style>
