<template>
  <section class="app-page ocr-page" :class="{ 'is-dark': resolvedTheme === 'dark' }">
    <header class="ocr-header">
      <h1>识别中...</h1>
      <p>大模型正在识别题目</p>
    </header>

    <Card class="progress-card">
      <CardContent class="progress-content">
        <div class="stages">
          <div
            v-for="(stage, index) in stages"
            :key="stage.key"
            class="stage-item"
            :class="{ 'stage-done': stageStatus[stage.key] === 'done', 'stage-active': stageStatus[stage.key] === 'active', 'stage-error': stageStatus[stage.key] === 'error' }"
          >
            <div class="stage-icon">
              <div v-if="stageStatus[stage.key] === 'done'" class="checkmark">✓</div>
              <div v-else-if="stageStatus[stage.key] === 'active'" class="spinner" />
              <div v-else class="circle" />
            </div>
            <div class="stage-text">
              <h3>{{ stage.label }}</h3>
              <p v-if="stageMessages[stage.key]" class="stage-message">{{ stageMessages[stage.key] }}</p>
            </div>
          </div>
        </div>

        <div v-if="finalData" class="final-result">
          <div class="result-item">
            <span class="label">学科：</span>
            <span class="value">{{ finalData.subject }}</span>
          </div>
          <div class="result-item">
            <span class="label">题型：</span>
            <span class="value">{{ finalData.question_type }}</span>
          </div>
          <div class="result-item">
            <span class="label">题目：</span>
            <span class="value">{{ finalData.title || '（自动生成）' }}</span>
          </div>
          <div class="result-item">
            <span class="label">标签数：</span>
            <span class="value">{{ (finalData.tags || []).length }}</span>
          </div>
        </div>

        <div v-if="errorMessage" class="error-box">
          <p>{{ errorMessage }}</p>
          <Button @click="retryIdentify" class="mt-2">重新识别</Button>
        </div>

        <div v-if="stageStatus.final === 'done'" class="success-actions">
          <Button @click="goToEditor" variant="default" class="w-full">编辑并保存</Button>
          <Button @click="goBackToList" variant="outline" class="w-full mt-2">返回列表</Button>
        </div>
      </CardContent>
    </Card>
  </section>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue';
import { useRouter } from 'vue-router';
import { useTheme } from '@/composables/useTheme';
import { generateLatexDraftByVisionStream } from '@/services/ai';
import { saveVisionDraftToStorage } from '@/services/ai';
import type { VisionStreamEvent, VisionLatexDraft } from '@/types/domain';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';

const router = useRouter();
const { resolvedTheme } = useTheme();

const stages = [
  { key: 'classify', label: '分类识别', icon: '🏷️' },
  { key: 'latex', label: '公式提取', icon: '∑' },
  { key: 'tags', label: '标签生成', icon: '🏷️' },
  { key: 'solve', label: '答案求解', icon: '✍️' },
  { key: 'final', label: '完成识别', icon: '✅' },
];

const stageStatus = reactive<Record<string, 'pending' | 'active' | 'done' | 'error'>>({
  classify: 'pending',
  latex: 'pending',
  tags: 'pending',
  solve: 'pending',
  final: 'pending',
});

const stageMessages = reactive<Record<string, string>>({});
const errorMessage = ref('');
const finalData = ref<any>(null);
const imageBase64 = ref('');

onMounted(async () => {
  // 获取从路由state传来的base64图片
  const state = router.currentRoute.value.query.data as string;
  if (!state) {
    errorMessage.value = '未获取到图片数据，请重新拍照';
    return;
  }

  imageBase64.value = state;
  startIdentify();
});

async function startIdentify() {
  try {
    errorMessage.value = '';
    finalData.value = null;
    resetStages();

    await generateLatexDraftByVisionStream(imageBase64.value, (evt: VisionStreamEvent) => {
      handleStreamEvent(evt);
    });

    // 流程完成
    stageStatus.final = 'done';
  } catch (error: any) {
    errorMessage.value = error?.message || '识别失败，请重试';
    stageStatus[getLastActiveStage()] = 'error';
  }
}

function handleStreamEvent(evt: VisionStreamEvent) {
  const stage = evt.stage as keyof typeof stageStatus;

  if (evt.error) {
    stageStatus[stage] = 'error';
    stageMessages[stage] = evt.error;
    return;
  }

  if (stage) {
    stageStatus[stage] = 'active';
  }

  // 更新进度消息
  switch (stage) {
    case 'classify':
      if (evt.subject) stageMessages.classify = `识别到: ${evt.subject}${evt.question_type ? ' · ' + evt.question_type : ''}`;
      break;
    case 'latex':
      if (evt.latex_question) stageMessages.latex = '已提取题目LaTeX';
      break;
    case 'tags':
      if (evt.tags && evt.tags.length) stageMessages.tags = `已生成 ${evt.tags.length} 个标签`;
      break;
    case 'solve':
      if (evt.latex_answer) stageMessages.solve = '已生成答案';
      break;
    case 'final':
      if (evt.done) {
        finalData.value = {
          subject: evt.subject,
          question_type: evt.question_type,
          title: evt.title,
          tags: evt.tags,
          latex_question: evt.latex_question,
          latex_answer: evt.latex_answer,
          latex_solution: evt.latex_solution,
        };
        // 保存到本地存储，以便编辑页面使用
        const draft: VisionLatexDraft = {
          latexQuestion: evt.latex_question || '',
          latexAnswer: evt.latex_answer || '',
          latexSolution: evt.latex_solution || '',
          tags: evt.tags || [],
          questionType: evt.question_type,
          subject: evt.subject,
          title: evt.title,
        };
        saveVisionDraftToStorage(draft);
        stageStatus.final = 'done';
        stageMessages.final = '识别完成！';
      }
      break;
  }

  // 当前stage完成时标记为done
  if (evt.done && stage) {
    stageStatus[stage] = 'done';
  }
}

function goToEditor() {
  router.push('/records/new');
}

function goBackToList() {
  router.push('/tabs/errors');
}

function retryIdentify() {
  startIdentify();
}

function resetStages() {
  for (const key of Object.keys(stageStatus)) {
    stageStatus[key] = 'pending';
    delete stageMessages[key];
  }
}

function getLastActiveStage(): string {
  const stageKeys = Object.keys(stageStatus);
  for (let i = stageKeys.length - 1; i >= 0; i--) {
    if (stageStatus[stageKeys[i]] === 'active') {
      return stageKeys[i];
    }
  }
  return 'classify';
}
</script>

<style scoped>
.ocr-page {
  min-height: 100vh;
  padding: 20px 16px;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.ocr-page.is-dark {
  background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
}

.ocr-header {
  text-align: center;
  color: #1e293b;
}

.ocr-page.is-dark .ocr-header {
  color: #f1f5f9;
}

.ocr-header h1 {
  font-size: 28px;
  font-weight: 600;
  margin: 0;
  margin-bottom: 8px;
}

.ocr-header p {
  color: #64748b;
  margin: 0;
}

.ocr-page.is-dark .ocr-header p {
  color: #cbd5e1;
}

.progress-card {
  margin: 0 auto;
  max-width: 400px;
  width: 100%;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.1);
}

.progress-content {
  padding: 24px;
}

.stages {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-bottom: 24px;
}

.stage-item {
  display: flex;
  gap: 12px;
  align-items: flex-start;
  transition: opacity 0.3s ease;
}

.stage-item.stage-done .stage-icon {
  background: #10b981;
}

.stage-item.stage-active .stage-icon {
  background: #3b82f6;
}

.stage-item.stage-error .stage-icon {
  background: #ef4444;
}

.stage-icon {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #e2e8f0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  color: white;
  position: relative;
  transition: background 0.3s ease;

  .ocr-page.is-dark & {
    background: #334155;
  }
}

.checkmark {
  font-size: 20px;
  animation: popIn 0.4s cubic-bezier(0.68, -0.55, 0.265, 1.55);
}

.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.circle {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #cbd5e1;

  .ocr-page.is-dark & {
    background: #475569;
  }
}

.stage-text h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 500;
  color: #1e293b;

  .ocr-page.is-dark & {
    color: #f1f5f9;
  }
}

.stage-text p {
  margin: 4px 0 0;
  font-size: 12px;
  color: #64748b;

  .ocr-page.is-dark & {
    color: #cbd5e1;
  }
}

.stage-message {
  display: inline-block;
}

.final-result {
  background: rgba(59, 130, 246, 0.1);
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 16px;

  .ocr-page.is-dark & {
    background: rgba(59, 130, 246, 0.2);
  }

  .result-item {
    display: flex;
    justify-content: space-between;
    font-size: 13px;
    padding: 6px 0;
    color: #1e293b;

    .ocr-page.is-dark & {
      color: #e2e8f0;
    }

    .label {
      font-weight: 500;
      color: #475569;

      .ocr-page.is-dark & {
        color: #cbd5e1;
      }
    }

    .value {
      text-align: right;
      max-width: 180px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }
}

.error-box {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 16px;
  color: #dc2626;

  .ocr-page.is-dark & {
    background: rgba(239, 68, 68, 0.15);
    color: #fca5a5;
  }

  p {
    margin: 0;
    font-size: 13px;
  }
}

.success-actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@keyframes popIn {
  0% {
    transform: scale(0) rotate(-45deg);
    opacity: 0;
  }
  50% {
    transform: scale(1.2) rotate(10deg);
  }
  100% {
    transform: scale(1) rotate(0);
  }
}
</style>
