<template>
  <section class="app-page app-inner-page page-wrap pt-8 crop-page" :class="{ 'is-dark': resolvedTheme === 'dark' }">
    <header class="app-page-header page-header">
      <Button variant="outline" size="icon-sm" class="app-header-back mr-4" @click="goBack" aria-label="返回上一级"><</Button>
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
          </div>

          <div class="crop-zoom" aria-label="缩放控制">
            <Button type="button" variant="outline" size="sm" :disabled="processing" @click="zoomOut">-</Button>
            <input
              class="zoom-range"
              type="range"
              min="1"
              max="3"
              step="0.01"
              :value="zoomValue"
              :disabled="processing || !cropperReady"
              @input="onZoomInput"
            />
            <Button type="button" variant="outline" size="sm" :disabled="processing" @click="zoomIn">+</Button>
          </div>

          <div class="crop-meta">
            <div>
              <p class="meta-label">说明</p>
              <p class="meta-text">单指拖动画面，双指可缩放；确保题干、选项和小问都在框内。</p>
            </div>
            <div>
              <p class="meta-label">当前裁剪尺寸</p>
              <p class="meta-text">{{ cropRatioText }}</p>
            </div>
          </div>

          <div class="crop-actions">
            <Button variant="outline" :disabled="processing || !cropperReady" @click="resetCrop">重置裁剪</Button>
            <Button variant="outline" :disabled="processing" @click="goBack">返回</Button>
            <Button :disabled="processing || !cropperReady" @click="confirmCrop">
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
import Cropper from 'cropperjs';
import 'cropperjs/dist/cropper.css';
import { computed, onBeforeUnmount, onMounted, ref, shallowRef } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useTheme } from '@/composables/useTheme';
import { loadImagePayload, saveImagePayload } from '@/services/image-transfer';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Card, CardContent } from '@/components/ui/card';

const route = useRoute();
const router = useRouter();
const { resolvedTheme } = useTheme();

const imageBase64 = ref('');
const hasImage = computed(() => imageBase64.value.length > 0);
const imageUrl = computed(() => (hasImage.value ? `data:image/jpeg;base64,${imageBase64.value}` : ''));
const errorMessage = ref('');
const processing = ref(false);
const stageEl = ref<HTMLElement | null>(null);
const imageEl = ref<HTMLImageElement | null>(null);
const cropper = shallowRef<Cropper | null>(null);
const zoomValue = ref(1);
const cropData = ref<{ width: number; height: number } | null>(null);
const cropperReady = ref(false);

const cropRatioText = computed(() => {
  if (!cropData.value) {
    return '未初始化';
  }

  return `${Math.round(cropData.value.width)} × ${Math.round(cropData.value.height)}`;
});

onMounted(() => {
  const key = String(route.query.key ?? '').trim();
  const legacyData = String(route.query.data ?? '').trim();

  if (key) {
    const payload = loadImagePayload(key);
    if (payload) {
      imageBase64.value = payload;
    }
  }

  if (!imageBase64.value && legacyData) {
    imageBase64.value = legacyData;
    const migratedKey = saveImagePayload(legacyData);
    router.replace({
      path: '/image/crop',
      query: { key: migratedKey },
    });
  }

  if (!imageBase64.value) {
    errorMessage.value = '未获取到图片数据，请重新拍照';
  }
});

onBeforeUnmount(() => {
  destroyCropper();
});

function handleImageLoad() {
  initializeCropper();
}

function initializeCropper() {
  const image = imageEl.value;
  if (!image) {
    return;
  }

  destroyCropper();
  cropperReady.value = false;
  zoomValue.value = 1;
  cropData.value = null;

  cropper.value = new Cropper(image, {
    viewMode: 1,
    dragMode: 'move',
    autoCropArea: 0.92,
    responsive: true,
    restore: false,
    background: false,
    guides: true,
    center: true,
    highlight: false,
    movable: true,
    zoomable: true,
    cropBoxMovable: true,
    cropBoxResizable: true,
    toggleDragModeOnDblclick: false,
    minContainerWidth: 280,
    minContainerHeight: 360,
    ready() {
      cropperReady.value = true;
      syncZoomFromCropper();
      syncCropData();
    },
    crop() {
      syncCropData();
    },
    zoom() {
      syncZoomFromCropper();
    },
  });
}

function destroyCropper() {
  cropper.value?.destroy();
  cropper.value = null;
  cropperReady.value = false;
}

function syncCropData() {
  const instance = cropper.value;
  if (!instance) {
    return;
  }

  const data = instance.getData(true);
  cropData.value = {
    width: data.width || 0,
    height: data.height || 0,
  };
}

function syncZoomFromCropper() {
  const instance = cropper.value;
  if (!instance) {
    return;
  }

  const imageData = instance.getImageData();
  const naturalWidth = imageData.naturalWidth || 1;
  const ratio = clamp(imageData.width / naturalWidth, 1, 3);
  zoomValue.value = Number(ratio.toFixed(2));
}

function onZoomInput(event: Event) {
  const target = event.target as HTMLInputElement | null;
  if (!target) {
    return;
  }

  const nextValue = clamp(Number(target.value), 1, 3);
  zoomValue.value = nextValue;
  cropper.value?.zoomTo(nextValue);
}

function resetCrop() {
  cropper.value?.reset();
  syncZoomFromCropper();
  syncCropData();
}

function zoomIn() {
  const next = clamp(zoomValue.value + 0.1, 1, 3);
  zoomValue.value = Number(next.toFixed(2));
  cropper.value?.zoomTo(zoomValue.value);
}

function zoomOut() {
  const next = clamp(zoomValue.value - 0.1, 1, 3);
  zoomValue.value = Number(next.toFixed(2));
  cropper.value?.zoomTo(zoomValue.value);
}

async function confirmCrop() {
  const instance = cropper.value;
  if (!instance || !cropperReady.value) {
    errorMessage.value = '裁剪器尚未初始化';
    return;
  }

  processing.value = true;
  try {
    const canvas = instance.getCroppedCanvas({
      fillColor: '#ffffff',
      imageSmoothingEnabled: true,
      imageSmoothingQuality: 'high',
    });
    if (!canvas) {
      throw new Error('无法生成裁剪结果');
    }

    const croppedBase64 = canvas.toDataURL('image/jpeg', 0.95).split(',')[1] || '';
    if (!croppedBase64) {
      throw new Error('裁剪失败');
    }

    const croppedKey = saveImagePayload(croppedBase64);

    router.push({
      path: '/ocr/progress',
      query: { key: croppedKey },
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
  --crop-page-gutter: 10px;
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
  padding: 12px;
}

.crop-stage {
  position: relative;
  width: 100%;
  min-height: min(76dvh, 620px);
  border-radius: 18px;
  overflow: hidden;
  background: linear-gradient(180deg, rgba(15, 23, 42, 0.12), rgba(15, 23, 42, 0.06));
  touch-action: none;
}

.crop-image {
  width: 100%;
  height: min(76dvh, 620px);
  object-fit: contain;
  user-select: none;
  opacity: 0;
  -webkit-user-drag: none;
}

.crop-zoom {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 10px;
}

.zoom-range {
  width: 100%;
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

.crop-stage :deep(.cropper-view-box),
.crop-stage :deep(.cropper-face) {
  border-radius: 12px;
}

.crop-stage :deep(.cropper-line),
.crop-stage :deep(.cropper-point) {
  background: #2563eb;
}

.crop-stage :deep(.cropper-point.point-se) {
  width: 16px;
  height: 16px;
}

@media (max-width: 640px) {
  .crop-page {
    padding-left: var(--crop-page-gutter);
    padding-right: var(--crop-page-gutter);
  }

  .crop-body {
    padding: 10px;
    gap: 10px;
  }

  .crop-stage {
    min-height: 72dvh;
    border-radius: 14px;
  }

  .crop-image {
    height: 72dvh;
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
