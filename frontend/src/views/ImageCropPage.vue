<template>
    <section class="app-page app-inner-page page-wrap pt-8 crop-page" :class="{ 'is-dark': resolvedTheme === 'dark' }">
        <header class="app-page-header page-header">
            <Button variant="outline" size="icon-sm" class="app-header-back mr-4" @click="goBack" aria-label="返回上一级">
                <
            </Button>
            <h1>裁剪题目</h1>
            <p>拖动裁剪框，把题目内容尽量完整框住后再进入识别。</p>
        </header>

        <Alert v-if="errorMessage" variant="destructive">
            <AlertTitle>无法裁剪</AlertTitle>
            <AlertDescription>{{ errorMessage }}</AlertDescription>
        </Alert>

        <div class="crop-body">
            <div ref="stageEl" class="crop-stage">
                <img ref="imageEl" :src="imageUrl" alt="待裁剪图片" class="crop-image" @load="handleImageLoad" />

                <div class="crop-stage-tip" v-if="cropperReady">
                    <span>拖动边框完成裁剪</span>
                    <strong>{{ zoomPercentText }}</strong>
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
const imageEl = ref<HTMLImageElement | null>(null);
const cropper = shallowRef<Cropper | null>(null);
const zoomValue = ref(1);
const cropData = ref<{ width: number; height: number } | null>(null);
const cropperReady = ref(false);

const zoomPercentText = computed(() => `${Math.round(zoomValue.value * 100)}%`);

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

function resetCrop() {
    cropper.value?.reset();
    syncZoomFromCropper();
    syncCropData();
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
    gap: 12px;
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

.crop-stage-tip {
    position: absolute;
    left: 10px;
    right: 10px;
    top: 10px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    border-radius: 12px;
    background: rgba(15, 23, 42, 0.58);
    color: #f8fafc;
    padding: 8px 10px;
    font-size: 12px;
    z-index: 2;
    backdrop-filter: blur(8px);
}

.crop-image {
    width: 100%;
    height: min(76dvh, 620px);
    object-fit: contain;
    user-select: none;
    opacity: 0;
    -webkit-user-drag: none;
}

.crop-controls {
    border-radius: 14px;
    padding: 10px;
    display: grid;
    gap: 10px;
}

.crop-zoom-head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
}

.zoom-value {
    margin: 0;
    font-size: 13px;
    font-weight: 700;
    color: #2563eb;
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
    grid-template-columns: 1.5fr 1fr;
}

.meta-item {
    border: 1px solid rgba(148, 163, 184, 0.22);
    border-radius: 12px;
    background: rgba(248, 250, 252, 0.72);
    padding: 9px 10px;
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

.is-dark .crop-controls {
    border-color: rgba(148, 163, 184, 0.24);
    background: rgba(15, 23, 42, 0.52);
}

.is-dark .meta-item {
    border-color: rgba(148, 163, 184, 0.28);
    background: rgba(15, 23, 42, 0.46);
}

.is-dark .meta-label {
    color: #94a3b8;
}

.is-dark .meta-text {
    color: #e2e8f0;
}

.is-dark .zoom-value {
    color: #7dd3fc;
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

    .crop-meta {
        grid-template-columns: 1fr;
    }

    .crop-actions {
        position: sticky;
        bottom: calc(8px + env(safe-area-inset-bottom, 0px));
        background: rgba(255, 255, 255, 0.86);
        border: 1px solid rgba(148, 163, 184, 0.24);
        border-radius: 14px;
        padding: 8px;
        backdrop-filter: blur(10px);
        z-index: 3;
    }

    .is-dark .crop-actions {
        background: rgba(15, 23, 42, 0.82);
        border-color: rgba(148, 163, 184, 0.26);
    }
}
</style>
