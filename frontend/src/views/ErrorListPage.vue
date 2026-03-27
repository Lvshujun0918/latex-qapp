<template>
  <section class="app-page page-wrap" :class="{ 'is-dark': resolvedTheme === 'dark' }">
    <header class="app-page-header page-header">
      <span class="app-kicker">Collection</span>
      <h1>错题</h1>
      <p>记录、检索并持续迭代你的 LaTeX 错题集。</p>
    </header>

    <Alert v-if="errorMessage" variant="destructive">
      <AlertTitle>拍照生成功能异常</AlertTitle>
      <AlertDescription>{{ errorMessage }}</AlertDescription>
    </Alert>

    <Card class="hero-card app-page-shell">
      <CardHeader>
        <CardDescription>AI 错题本</CardDescription>
        <CardTitle>今天继续攻克薄弱点</CardTitle>
      </CardHeader>
      <CardContent>
        当前共 {{ records.length }} 道题，优先复习高频错因与低掌握度题目。
      </CardContent>
    </Card>

    <div class="toolbar">
      <Input v-model="keyword" placeholder="搜索 LaTeX 题目" />
      <Button :disabled="isGenerating" @click="openSourceDialog">
        <Camera class="mr-2 h-4 w-4" />
        {{ isGenerating ? '识别中...' : '拍照录题' }}
      </Button>
    </div>

    <div v-if="isGenerating" class="loading-inline">
      {{ generatingMessage }}
    </div>

    <div v-if="filteredRecords.length" class="record-list">
      <button
        v-for="record in filteredRecords"
        :key="record.id"
        class="record-item app-interactive-surface"
        type="button"
        @click="toDetail(record.id)"
      >
        <div>
          <h3>{{ record.title || '未命名题目' }}</h3>
          <p>{{ record.subject }} · 难度 {{ record.difficulty }} · {{ record.syncStatus }}</p>
        </div>
        <Badge>LaTeX</Badge>
      </button>
    </div>

    <Card v-else class="empty-card">
      <CardContent class="empty-content">
        <div class="empty-icon-wrap">
          <Sparkles class="empty-icon" />
        </div>
        <h3>暂无错题</h3>
        <p>点击拍照录题后，由大模型自动生成 LaTeX 题目、答案与标签。</p>
        <Button :disabled="isGenerating" @click="openSourceDialog">
          <Camera class="mr-2 h-4 w-4" />
          {{ isGenerating ? '识别中...' : '立即拍照录题' }}
        </Button>
      </CardContent>
    </Card>

    <ImageSourceDialog
      :open="sourceDialogOpen"
      @update:open="(val) => (sourceDialogOpen = val)"
      @select="createFromCamera"
    />
  </section>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia';
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { Camera, Sparkles } from 'lucide-vue-next';
import ImageSourceDialog from '@/components/ImageSourceDialog.vue';
import { useTheme } from '@/composables/useTheme';
import { useRecordStore } from '@/stores/record';
import { generateLatexDraftByVisionStream, pickImageAsBase64, saveVisionDraftToStorage } from '@/services/ai';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';

const router = useRouter();
const recordStore = useRecordStore();
const { resolvedTheme } = useTheme();
const { records } = storeToRefs(recordStore);
const keyword = ref('');
const isGenerating = ref(false);
const generatingMessage = ref('正在识别题目与标签...');
const sourceDialogOpen = ref(false);
const errorMessage = ref('');

const filteredRecords = computed(() => {
  const term = keyword.value.trim().toLowerCase();

  if (!term) {
    return records.value;
  }

  return records.value.filter((record) => {
    const title = (record.title || '').toLowerCase();
    const subject = (record.subject || '').toLowerCase();
    return title.includes(term) || subject.includes(term) || String(record.id).includes(term);
  });
});

onMounted(() => {
  recordStore.reload();
});

function toDetail(id: number) {
  router.push(`/records/${id}`);
}

function openSourceDialog() {
  if (isGenerating.value) {
    return;
  }

  errorMessage.value = '';
  sourceDialogOpen.value = true;
}

async function createFromCamera(source: 'camera' | 'album' | 'file') {
  if (isGenerating.value) {
    return;
  }

  try {
    errorMessage.value = '';
    isGenerating.value = true;
    generatingMessage.value = '正在准备图片...';

    const imageBase64 = await pickImageAsBase64(source);
    const draft = await generateLatexDraftByVisionStream(imageBase64, (evt) => {
      switch (evt.stage) {
        case 'classify':
          generatingMessage.value = '正在识别学科与题型...';
          break;
        case 'latex':
          generatingMessage.value = '正在生成题目 LaTeX...';
          break;
        case 'tags':
          generatingMessage.value = '正在生成标签...';
          break;
        case 'final':
          generatingMessage.value = '识别完成，正在进入编辑页...';
          break;
        default:
          break;
      }
    });

    if (!draft.latexQuestion.trim()) {
      throw new Error('识别结果为空');
    }

    saveVisionDraftToStorage(draft);
    router.push('/records/new');
  } catch (error: any) {
    errorMessage.value = error?.message || '拍照或识别失败，请重试。';
  } finally {
    isGenerating.value = false;
    generatingMessage.value = '正在识别题目与标签...';
  }
}
</script>

<style scoped>
.page-wrap {
  gap: 12px;
}

.hero-card {
  border-radius: 18px;
  backdrop-filter: saturate(165%) blur(16px);
}

.toolbar {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 10px;
}

.loading-inline {
  font-size: 13px;
  color: #2563eb;
}

.record-list {
  display: grid;
  gap: 10px;
}

.record-item {
  width: 100%;
  border: 1px solid rgba(148, 163, 184, 0.34);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.72);
  backdrop-filter: saturate(150%) blur(12px);
  padding: 14px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  text-align: left;
  cursor: pointer;
  transition: transform 0.2s ease, border-color 0.2s ease, background-color 0.2s ease, box-shadow 0.2s ease;
}

.record-item:hover {
  border-color: rgba(59, 130, 246, 0.4);
  box-shadow: 0 12px 24px rgba(15, 23, 42, 0.08);
}

.record-item:focus-visible {
  outline: 0;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.26);
}

.record-item h3 {
  margin: 0;
  font-size: 16px;
}

.record-item p {
  margin: 6px 0 0;
  color: #475569;
  font-size: 13px;
}

.empty-card {
  margin-top: 6px;
}

.empty-content {
  display: flex;
  align-items: center;
  flex-direction: column;
  text-align: center;
  gap: 10px;
  padding: 20px;
}

.empty-icon-wrap {
  width: 64px;
  height: 64px;
  display: grid;
  place-items: center;
  border-radius: 20px;
  background: linear-gradient(145deg, rgba(31, 122, 255, 0.2), rgba(31, 122, 255, 0.06));
}

.empty-icon {
  width: 30px;
  height: 30px;
  color: #2563eb;
}

.empty-content h3 {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
}

.empty-content p {
  margin: 0;
  color: #475569;
}

.is-dark .record-item {
  border-color: rgba(148, 163, 184, 0.24);
  background: rgba(30, 41, 59, 0.56);
}

.is-dark .record-item p {
  color: #cbd5e1;
}

@media (max-width: 640px) {
  .toolbar {
    grid-template-columns: 1fr;
  }
}
</style>
