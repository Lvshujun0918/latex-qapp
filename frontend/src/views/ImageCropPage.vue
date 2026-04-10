<template>
  <section class="app-page app-inner-page page-wrap pt-8 crop-page" :class="{ 'is-dark': resolvedTheme === 'dark' }">
    <header class="app-page-header page-header">
      <Button variant="outline" size="icon-sm" class="app-header-back" @click="goBack" aria-label="返回上一级"><</Button>
      <span class="app-kicker">Crop Stage</span>
      <h1>裁剪题目</h1>
      <p>拖动裁剪框，把题目内容尽量完整框住后再进入识别。</p>
    </header>

    <Alert v-if="errorMessage" variant="destructive">
      <AlertTitle>无法裁剪</AlertTitle>
      <AlertDescription>{{ errorMessage }}</AlertDescription>
    </Alert>

    <Card v-if="hasImage" class="crop-card app-page-shell">
      <CardContent class="crop-content !p-0">
        <div class="crop-body">
        <div ref="stageEl" class="crop-stage">
          <img ref="imageEl" :src="imageUrl" alt="待裁剪图片" class="crop-image" @load="handleImageLoad" />

          <div
            v-if="cropRect"
            class="crop-box"
            :style="cropStyle"
            @pointerdown.stop.prevent="startDrag"
          >
            <div class="crop-hint">拖拽移动</div>
            <span class="crop-corner crop-corner-tl" />
            <span class="crop-corner crop-corner-tr" />
            <span class="crop-corner crop-corner-bl" />
            <span class="crop-corner crop-corner-br" />
          </div>
        </div>

        <div class="crop-meta">
          <div>
            <p class="meta-label">说明</p>
            <p class="meta-text">框住题干、选项和小问，不要裁掉题号。</p>
          </div>
          <div>
            <p class="meta-label">当前裁剪比例</p>
            <p class="meta-text">{{ cropRatioText }}</p>
          </div>
        </div>

        <div class="crop-actions">
          <Button variant="outline" :disabled="processing" @click="resetCrop">重置裁剪</Button>
          <Button variant="outline" :disabled="processing" @click="goBack">返回</Button>
          <Button :disabled="processing || !cropRect" @click="confirmCrop">
            {{ processing ? '裁剪中...' : '确认裁剪并识别' }}
          </Button>
        </div>
        </div>
      </CardContent>
    </Card>

    <Card v-else class="crop-card app-soft-card">
      <CardContent class="crop-empty">未获取到图片数据，请重新拍照或从相册导入。</CardContent>
    </Card>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useTheme } from '@/composables/useTheme';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Card, CardContent } from '@/components/ui/card';

const route = useRoute();
const router = useRouter();
const { resolvedTheme } = useTheme();

const imageBase64 = computed(() => String(route.query.data ?? '').trim());
const hasImage = computed(() => imageBase64.value.length > 0);
const imageUrl = computed(() => (hasImage.value ? `data:image/jpeg;base64,${imageBase64.value}` : ''));
const errorMessage = ref('');
const processing = ref(false);
const stageEl = ref<HTMLElement | null>(null);
const imageEl = ref<HTMLImageElement | null>(null);
const cropRect = ref<{ x: number; y: number; width: number; height: number } | null>(null);
const dragState = ref<{ offsetX: number; offsetY: number } | null>(null);
const dragPointerId = ref<number | null>(null);
const dragTarget = ref<HTMLElement | null>(null);

const cropStyle = computed(() => {
  if (!cropRect.value) {
    return {};
  }

  return {
    left: `${cropRect.value.x}px`,
    top: `${cropRect.value.y}px`,
    width: `${cropRect.value.width}px`,
    height: `${cropRect.value.height}px`,
  };
});

const cropRatioText = computed(() => {
  if (!cropRect.value || !imageEl.value) {
    return '未初始化';
  }

  const naturalWidth = imageEl.value.naturalWidth || 1;
  const naturalHeight = imageEl.value.naturalHeight || 1;
  const displayedWidth = imageEl.value.clientWidth || 1;
  const displayedHeight = imageEl.value.clientHeight || 1;
  const scaleX = naturalWidth / displayedWidth;
  const scaleY = naturalHeight / displayedHeight;
  const sourceWidth = Math.round(cropRect.value.width * scaleX);
  const sourceHeight = Math.round(cropRect.value.height * scaleY);
  return `${sourceWidth} × ${sourceHeight}`;
});

onMounted(() => {
  if (!hasImage.value) {
    errorMessage.value = '未获取到图片数据，请重新拍照';
  }
});

onBeforeUnmount(() => {
  stopDragging();
});

function handleImageLoad() {
  initializeCropRect();
}

function initializeCropRect() {
  const el = imageEl.value;
  if (!el) {
    return;
  }

  const width = el.clientWidth;
  const height = el.clientHeight;
  if (!width || !height) {
    return;
  }

  const cropWidth = Math.max(180, Math.floor(width * 0.88));
  const cropHeight = Math.max(180, Math.floor(height * 0.72));
  cropRect.value = {
    x: Math.floor((width - cropWidth) / 2),
    y: Math.floor((height - cropHeight) / 2),
    width: cropWidth,
    height: cropHeight,
  };
}

function resetCrop() {
  initializeCropRect();
}

function startDrag(event: PointerEvent) {
  if (!cropRect.value || !stageEl.value) {
    return;
  }

  const target = event.currentTarget as HTMLElement | null;
  if (!target) {
    return;
  }

  event.preventDefault();

  dragState.value = {
    offsetX: event.clientX - cropRect.value.x,
    offsetY: event.clientY - cropRect.value.y,
  };
  dragPointerId.value = event.pointerId;
  dragTarget.value = target;

  if (target.setPointerCapture) {
    target.setPointerCapture(event.pointerId);
  }

  window.addEventListener('pointermove', onPointerMove, { passive: false });
  window.addEventListener('pointerup', stopDragging, { once: true });
  window.addEventListener('pointercancel', stopDragging, { once: true });
}

function onPointerMove(event: PointerEvent) {
  if (!dragState.value || !cropRect.value || !stageEl.value) {
    return;
  }

  if (dragPointerId.value !== null && event.pointerId !== dragPointerId.value) {
    return;
  }

  event.preventDefault();

  const stageRect = stageEl.value.getBoundingClientRect();
  const nextX = clamp(event.clientX - stageRect.left - dragState.value.offsetX, 0, stageRect.width - cropRect.value.width);
  const nextY = clamp(event.clientY - stageRect.top - dragState.value.offsetY, 0, stageRect.height - cropRect.value.height);
  cropRect.value = { ...cropRect.value, x: nextX, y: nextY };
}

function stopDragging() {
  dragState.value = null;
  if (dragTarget.value && dragPointerId.value !== null && dragTarget.value.hasPointerCapture?.(dragPointerId.value)) {
    dragTarget.value.releasePointerCapture?.(dragPointerId.value);
  }
  dragPointerId.value = null;
  dragTarget.value = null;
  window.removeEventListener('pointermove', onPointerMove);
  window.removeEventListener('pointercancel', stopDragging);
}

async function confirmCrop() {
  const img = imageEl.value;
  const rect = cropRect.value;
  if (!img || !rect) {
    errorMessage.value = '裁剪框尚未初始化';
    return;
  }

  processing.value = true;
  try {
    const naturalWidth = img.naturalWidth;
    const naturalHeight = img.naturalHeight;
    const displayedWidth = img.clientWidth;
    const displayedHeight = img.clientHeight;
    if (!naturalWidth || !naturalHeight || !displayedWidth || !displayedHeight) {
      throw new Error('图片尚未加载完成');
    }

    const scaleX = naturalWidth / displayedWidth;
    const scaleY = naturalHeight / displayedHeight;

    const sourceX = Math.max(0, Math.round(rect.x * scaleX));
    const sourceY = Math.max(0, Math.round(rect.y * scaleY));
    const sourceWidth = Math.max(1, Math.round(rect.width * scaleX));
    const sourceHeight = Math.max(1, Math.round(rect.height * scaleY));

    const canvas = document.createElement('canvas');
    canvas.width = sourceWidth;
    canvas.height = sourceHeight;

    const context = canvas.getContext('2d');
    if (!context) {
      throw new Error('无法创建裁剪画布');
    }

    context.drawImage(img, sourceX, sourceY, sourceWidth, sourceHeight, 0, 0, sourceWidth, sourceHeight);

    const croppedBase64 = canvas.toDataURL('image/jpeg', 0.95).split(',')[1] || '';
    if (!croppedBase64) {
      throw new Error('裁剪失败');
    }

    router.push({
      path: '/ocr/progress',
      query: { data: croppedBase64 },
    });
  } catch (error: any) {
    errorMessage.value = error?.message || '裁剪失败，请重试';
  } finally {
    processing.value = false;
  }
}

function clamp(value: number, min: number, max: number) {
  return Math.max(min, Math.min(max, value));
}

function goBack() {
  if (window.history.length > 1) {
    router.back();
    return;
  }

  router.replace('/tabs/errors');
}
</script>

<style scoped>
.page-wrap {
  gap: 14px;
}

.crop-page {
  --crop-page-gutter: 14px;
}

.crop-card {
  overflow: hidden;
}

.crop-content {
  display: grid;
  gap: 14px;
}

.crop-body {
  display: grid;
  gap: 14px;
  padding: 16px;
}

.crop-stage {
  position: relative;
  width: 100%;
  display: inline-block;
  border-radius: 16px;
  overflow: hidden;
  background: linear-gradient(180deg, rgba(15, 23, 42, 0.08), rgba(15, 23, 42, 0.04));
}

.crop-image {
  display: block;
  width: 100%;
  height: auto;
  user-select: none;
  -webkit-user-drag: none;
}

.crop-box {
  position: absolute;
  border: 2px solid rgba(59, 130, 246, 0.95);
  background: rgba(59, 130, 246, 0.08);
  box-shadow: 0 0 0 9999px rgba(15, 23, 42, 0.38);
  border-radius: 14px;
  cursor: grab;
  touch-action: none;
  user-select: none;
  -webkit-user-select: none;
}

.crop-box:active {
  cursor: grabbing;
}

.crop-hint {
  position: absolute;
  left: 10px;
  top: -28px;
  padding: 2px 8px;
  border-radius: 999px;
  background: rgba(15, 23, 42, 0.82);
  color: #fff;
  font-size: 11px;
  white-space: nowrap;
}

.crop-corner {
  position: absolute;
  width: 12px;
  height: 12px;
  border-radius: 999px;
  background: #fff;
  border: 2px solid #2563eb;
}

.crop-corner-tl {
  left: -7px;
  top: -7px;
}

.crop-corner-tr {
  right: -7px;
  top: -7px;
}

.crop-corner-bl {
  left: -7px;
  bottom: -7px;
}

.crop-corner-br {
  right: -7px;
  bottom: -7px;
}

.crop-meta {
  display: grid;
  gap: 8px;
}

.meta-label {
  margin: 0 0 2px;
  font-size: 12px;
  color: #64748b;
}

.meta-text {
  margin: 0;
  font-size: 13px;
  color: #0f172a;
}

.crop-actions {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.crop-empty {
  color: #475569;
  text-align: center;
  padding: 18px 12px;
}

@media (max-width: 640px) {
  .crop-page {
    padding-left: var(--crop-page-gutter);
    padding-right: var(--crop-page-gutter);
  }

  .crop-body {
    padding: 12px;
    gap: 12px;
  }

  .crop-stage {
    border-radius: 14px;
  }

  .crop-image {
    max-height: 72dvh;
    object-fit: contain;
  }

  .crop-box {
    border-width: 3px;
    border-radius: 12px;
  }

  .crop-hint {
    top: -24px;
    left: 8px;
    font-size: 10px;
  }

  .crop-corner {
    width: 14px;
    height: 14px;
  }

  .meta-text {
    font-size: 12px;
    line-height: 1.45;
  }

  .crop-actions {
    grid-template-columns: 1fr;
  }
}
</style>
