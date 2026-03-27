<template>
  <section class="page-wrap">
    <header class="page-header">
      <h1>错题</h1>
      <p>记录、检索并持续迭代你的 LaTeX 错题集。</p>
    </header>

    <Card class="hero-card">
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
      <Button :disabled="isGenerating" @click="createFromCamera">
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
        class="record-item"
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
        <Button :disabled="isGenerating" @click="createFromCamera">
          <Camera class="mr-2 h-4 w-4" />
          {{ isGenerating ? '识别中...' : '立即拍照录题' }}
        </Button>
      </CardContent>
    </Card>
  </section>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia';
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { Camera, Sparkles } from 'lucide-vue-next';
import { useRecordStore } from '@/stores/record';
import { generateLatexDraftByVisionStream, pickImageAsBase64, saveVisionDraftToStorage } from '@/services/ai';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';

const router = useRouter();
const recordStore = useRecordStore();
const { records } = storeToRefs(recordStore);
const keyword = ref('');
const isGenerating = ref(false);
const generatingMessage = ref('正在识别题目与标签...');

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

async function createFromCamera() {
  if (isGenerating.value) {
    return;
  }

  try {
    isGenerating.value = true;
    generatingMessage.value = '请选择图片来源...';
    const source = chooseImageSource();

    if (!source) {
      return;
    }

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
  } catch {
    window.alert('拍照或识别失败，请重试。');
  } finally {
    isGenerating.value = false;
    generatingMessage.value = '正在识别题目与标签...';
  }
}

function chooseImageSource(): 'camera' | 'album' | 'file' | null {
  const selected = window.prompt('选择图片来源: 1=拍照, 2=相册, 3=文件', '1');

  if (selected === null) {
    return null;
  }

  switch (selected.trim()) {
    case '1':
      return 'camera';
    case '2':
      return 'album';
    case '3':
      return 'file';
    default:
      return 'camera';
  }
}
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

@media (max-width: 640px) {
  .toolbar {
    grid-template-columns: 1fr;
  }
}
</style>
