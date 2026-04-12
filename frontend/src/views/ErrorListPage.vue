<template>
  <section class="app-page page-wrap" :class="{ 'is-dark': resolvedTheme === 'dark' }">
    <header class="app-page-header page-header">
      <span class="app-kicker">Parent Console</span>
      <h1>错题</h1>
      <p>帮助家长记录、检索并持续迭代孩子的 LaTeX 错题集。</p>
    </header>

    <Alert v-if="errorMessage" variant="destructive">
      <AlertTitle>拍照生成功能异常</AlertTitle>
      <AlertDescription>{{ errorMessage }}</AlertDescription>
    </Alert>

    <Card class="hero-card app-page-shell">
      <CardHeader>
        <CardDescription>家长陪学面板</CardDescription>
        <CardTitle>今天优先攻克孩子的薄弱点</CardTitle>
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
      <div
        v-for="record in filteredRecords"
        :key="record.id"
        class="record-row"
        :class="{ opened: swipedId === record.id }"
      >
        <button class="delete-action" type="button" @click="requestDeleteRecord(record.id)">
          <Trash2 class="h-4 w-4" />
          删除
        </button>

        <button
          class="record-item app-interactive-surface"
          :class="{ swiped: swipedId === record.id, dragging: dragRecordId === record.id }"
          :style="recordStyle(record.id)"
          type="button"
          @touchstart="onTouchStart(record.id, $event)"
          @touchmove="onTouchMove(record.id, $event)"
          @touchend="onTouchEnd(record.id)"
          @touchcancel="onTouchCancel"
          @contextmenu.prevent
          @click="onRecordClick(record.id)"
        >
          <span class="subject-watermark" :class="watermarkClass(record.subject)">
            <component :is="subjectIcon(record.subject)" class="subject-watermark-icon" />
          </span>

          <div class="record-main">
            <h3>{{ record.title || '未命名题目' }}</h3>
            <p>{{ formatSubject(record.subject) }} · {{ formatQuestionType(record.questionType) }}</p>
          </div>
          <ChevronRight class="h-4 w-4 item-arrow" />
        </button>
      </div>
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

    <Dialog :open="deleteDialogOpen" @update:open="(val) => (deleteDialogOpen = val)">
      <DialogContent class="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>确认删除</DialogTitle>
          <DialogDescription>删除后无法恢复，是否继续删除该错题？</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" @click="deleteDialogOpen = false">取消</Button>
          <Button variant="destructive" @click="confirmDeleteRecord">确认删除</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </section>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia';
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { Atom, Calculator, Camera, ChevronRight, Dna, FlaskConical, Sparkles, Trash2 } from 'lucide-vue-next';
import ImageSourceDialog from '@/components/ImageSourceDialog.vue';
import { useTheme } from '@/composables/useTheme';
import { useRecordStore } from '@/stores/record';
import { pickImageAsBase64 } from '@/services/ai';
import { saveImagePayload } from '@/services/image-transfer';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
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
const swipedId = ref<number | null>(null);
const dragRecordId = ref<number | null>(null);
const dragOffsetX = ref(0);
const touchStartX = ref(0);
const deleteDialogOpen = ref(false);
const pendingDeleteId = ref<number | null>(null);

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

function onRecordClick(id: number) {
  if (swipedId.value && swipedId.value !== id) {
    swipedId.value = null;
    return;
  }

  if (swipedId.value === id) {
    swipedId.value = null;
    return;
  }

  toDetail(id);
}

function onTouchStart(id: number, event: TouchEvent) {
  touchStartX.value = event.touches[0]?.clientX ?? 0;
  dragRecordId.value = id;
  dragOffsetX.value = 0;
}

function onTouchMove(id: number, event: TouchEvent) {
  if (dragRecordId.value !== id) {
    return;
  }
  const currentX = event.touches[0]?.clientX ?? 0;
  const delta = currentX - touchStartX.value;
  dragOffsetX.value = Math.max(-88, Math.min(0, delta));
}

function onTouchEnd(id: number) {
  if (dragRecordId.value !== id) {
    return;
  }
  swipedId.value = dragOffsetX.value < -52 ? id : null;
  dragRecordId.value = null;
  dragOffsetX.value = 0;
}

function onTouchCancel() {
  dragRecordId.value = null;
  dragOffsetX.value = 0;
}

function recordStyle(id: number) {
  if (dragRecordId.value === id) {
    return { transform: `translateX(${dragOffsetX.value}px)` };
  }
  if (swipedId.value === id) {
    return { transform: 'translateX(-88px)' };
  }
  return { transform: 'translateX(0)' };
}

function requestDeleteRecord(id: number) {
  pendingDeleteId.value = id;
  deleteDialogOpen.value = true;
}

async function confirmDeleteRecord() {
  const id = pendingDeleteId.value;
  if (!id) {
    return;
  }

  try {
    await recordStore.deleteById(id);
    swipedId.value = null;
    pendingDeleteId.value = null;
    deleteDialogOpen.value = false;
  } catch (error: any) {
    errorMessage.value = error?.message || '删除失败，请重试。';
  }
}

function formatSubject(subject: string) {
  const value = String(subject || '').trim().toLowerCase();
  if (value === 'math' || value === '数学') return '数学';
  if (value === 'physics' || value === '物理') return '物理';
  if (value === 'chemistry' || value === '化学') return '化学';
  if (value === 'biology' || value === '生物') return '生物';
  return subject || '未知';
}

function formatQuestionType(questionType?: string) {
  const value = String(questionType || '').trim().toLowerCase();
  if (['choice', '选择', '选择题', 'single_choice', 'multiple_choice'].includes(value)) return '选择';
  if (['fill_blank', '填空', '填空题'].includes(value)) return '填空';
  if (['essay', '解答', '解答题', 'subjective'].includes(value)) return '解答';
  return questionType || '未知';
}

function subjectIcon(subject?: string) {
  const value = String(subject || '').trim().toLowerCase();
  if (value === 'math' || value === '数学') return Calculator;
  if (value === 'physics' || value === '物理') return Atom;
  if (value === 'chemistry' || value === '化学') return FlaskConical;
  if (value === 'biology' || value === '生物') return Dna;
  return Sparkles;
}

function watermarkClass(subject?: string) {
  const value = String(subject || '').trim().toLowerCase();
  if (value === 'math' || value === '数学') return 'is-math';
  if (value === 'physics' || value === '物理') return 'is-physics';
  if (value === 'chemistry' || value === '化学') return 'is-chemistry';
  if (value === 'biology' || value === '生物') return 'is-biology';
  return 'is-default';
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
    const imageKey = saveImagePayload(imageBase64);

    router.push({
      path: '/image/crop',
      query: {
        key: imageKey,
      },
    });
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

.record-row {
  position: relative;
  border-radius: 14px;
  overflow: hidden;
}

.delete-action {
  position: absolute;
  right: 0;
  top: 0;
  bottom: 0;
  width: 88px;
  border: 0;
  color: #fff;
  background: linear-gradient(180deg, #ef4444, #dc2626);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  font-size: 12px;
  font-weight: 600;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.2s ease;
}

.record-row.opened .delete-action {
  opacity: 1;
  pointer-events: auto;
}

.record-item {
  width: 100%;
  border: 1px solid rgba(148, 163, 184, 0.34);
  border-radius: 14px;
  background: #ffffff;
  padding: 14px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  text-align: left;
  cursor: pointer;
  transition: transform 0.2s ease, border-color 0.2s ease, background-color 0.2s ease, box-shadow 0.2s ease;
  position: relative;
  z-index: 1;
}

.record-main {
  min-width: 0;
  position: relative;
  z-index: 2;
  padding-left: 12px;
}

.subject-watermark {
  position: absolute;
  left: 12px;
  top: 50%;
  width: 40px;
  height: 40px;
  border-radius: 8px;
  display: grid;
  place-items: center;
  opacity: 0.15;
  pointer-events: none;
  z-index: 1;
}

.subject-watermark-icon {
  width: 72px;
  height: 72px;
}

.subject-watermark.is-math {
  color: #0f766e;
}

.subject-watermark.is-physics {
  color: #1d4ed8;
}

.subject-watermark.is-chemistry {
  color: #7c2d12;
}

.subject-watermark.is-biology {
  color: #166534;
}

.subject-watermark.is-default {
  color: #475569;
}

.record-item.dragging {
  transition: none;
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

.item-arrow {
  color: #94a3b8;
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
  background: rgba(30, 41, 59, 0.95);
}

.is-dark .record-item p {
  color: #cbd5e1;
}

.is-dark .subject-watermark {
  opacity: 0.20;
}

.is-dark .subject-watermark.is-math {
  color: #5eead4;
}

.is-dark .subject-watermark.is-physics {
  color: #93c5fd;
}

.is-dark .subject-watermark.is-chemistry {
  color: #fdba74;
}

.is-dark .subject-watermark.is-biology {
  color: #86efac;
}

.is-dark .subject-watermark.is-default {
  color: #cbd5e1;
}

@media (max-width: 640px) {
  .toolbar {
    grid-template-columns: 1fr;
  }
}
</style>
