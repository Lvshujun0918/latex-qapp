<template>
  <section class="app-page ocr-page" :class="{ 'is-dark': resolvedTheme === 'dark' }">
    <header class="ocr-header">
      <div class="ocr-badge">
        <ScanSearch class="h-4 w-4" />
        <span>Vision Pipeline</span>
      </div>
      <h1>识别进行中</h1>
      <p>每个阶段完成后会自动勾选，完成后可直接进入编辑页。</p>
    </header>

    <Card class="progress-card">
      <CardContent class="progress-content">
        <div class="stages">
          <div
            v-for="stage in stages"
            :key="stage.key"
            class="stage-item"
            :class="{
              'stage-done': stageStatus[stage.key] === 'done',
              'stage-active': stageStatus[stage.key] === 'active',
              'stage-error': stageStatus[stage.key] === 'error',
            }"
          >
            <div class="stage-icon">
              <Check v-if="stageStatus[stage.key] === 'done'" class="h-5 w-5" />
              <Loader2 v-else-if="stageStatus[stage.key] === 'active'" class="h-5 w-5 spinner" />
              <AlertCircle v-else-if="stageStatus[stage.key] === 'error'" class="h-5 w-5" />
              <component :is="stage.icon" v-else class="h-5 w-5 stage-default-icon" />
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
            <span class="value">{{ formatSubject(finalData.subject) }}</span>
          </div>
          <div class="result-item">
            <span class="label">题型：</span>
            <span class="value">{{ formatQuestionType(finalData.question_type) }}</span>
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
import { AlertCircle, Check, Loader2, ScanSearch, Sparkles, Tags, FunctionSquare, BrainCircuit } from 'lucide-vue-next';
import { useTheme } from '@/composables/useTheme';
import { generateLatexDraftByVisionStream } from '@/services/ai';
import { saveVisionDraftToStorage } from '@/services/ai';
import type { VisionStreamEvent, VisionLatexDraft } from '@/types/domain';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';

const router = useRouter();
const { resolvedTheme } = useTheme();

const stages = [
  { key: 'classify', label: '分类识别', icon: Sparkles },
  { key: 'latex', label: '公式提取', icon: FunctionSquare },
  { key: 'tags', label: '标签生成', icon: Tags },
  { key: 'solve', label: '答案求解', icon: BrainCircuit },
  { key: 'final', label: '完成识别', icon: Check },
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
  if (!stageStatus[stage]) {
    return;
  }

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

function formatSubject(subject?: string) {
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
</script>

<style scoped>
.ocr-page {
  min-height: 100vh;
  padding: 20px 16px 28px;
  background: radial-gradient(circle at 10% 10%, rgba(14, 116, 144, 0.12), transparent 45%),
    radial-gradient(circle at 90% 0%, rgba(14, 165, 233, 0.16), transparent 40%),
    linear-gradient(160deg, #f5f9ff 0%, #eef3ff 100%);
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.ocr-page.is-dark {
  background: radial-gradient(circle at 15% 10%, rgba(14, 165, 233, 0.2), transparent 40%),
    radial-gradient(circle at 90% 5%, rgba(59, 130, 246, 0.2), transparent 40%),
    linear-gradient(155deg, #0f172a 0%, #111827 100%);
}

.ocr-header {
  text-align: left;
  color: #1e293b;
}

.ocr-page.is-dark .ocr-header {
  color: #f1f5f9;
}

.ocr-badge {
  width: fit-content;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  color: #0c4a6e;
  background: rgba(14, 165, 233, 0.15);
  margin-bottom: 8px;
}

.ocr-page.is-dark .ocr-badge {
  color: #bae6fd;
  background: rgba(14, 165, 233, 0.22);
}

.ocr-header h1 {
  font-size: 26px;
  font-weight: 700;
  margin: 0;
  margin-bottom: 6px;
}

.ocr-header p {
  color: #64748b;
  margin: 0;
  font-size: 13px;
}

.ocr-page.is-dark .ocr-header p {
  color: #cbd5e1;
}

.progress-card {
  margin: 0 auto;
  max-width: 460px;
  width: 100%;
  box-shadow: 0 14px 38px rgba(15, 23, 42, 0.12);
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
  align-items: center;
  transition: transform 0.22s ease, opacity 0.22s ease;
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

.stage-item.stage-active {
  transform: translateX(2px);
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
}

.ocr-page.is-dark .stage-icon {
  background: #334155;
}

.stage-default-icon {
  color: #64748b;
}

.spinner {
  animation: spin 1s linear infinite;
}

.stage-text h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 500;
  color: #1e293b;
}

.ocr-page.is-dark .stage-text h3 {
  color: #f1f5f9;
}

.stage-text p {
  margin: 4px 0 0;
  font-size: 12px;
  color: #64748b;
}

.ocr-page.is-dark .stage-text p {
  color: #cbd5e1;
}

.stage-message {
  display: inline-block;
}

.final-result {
  background: rgba(59, 130, 246, 0.1);
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 16px;
}

.ocr-page.is-dark .final-result {
  background: rgba(59, 130, 246, 0.2);
}

.result-item {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  padding: 6px 0;
  color: #1e293b;
}

.ocr-page.is-dark .result-item {
  color: #e2e8f0;
}

.result-item .label {
  font-weight: 500;
  color: #475569;
}

.ocr-page.is-dark .result-item .label {
  color: #cbd5e1;
}

.result-item .value {
  text-align: right;
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.error-box {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 16px;
  color: #dc2626;
}

.ocr-page.is-dark .error-box {
  background: rgba(239, 68, 68, 0.15);
  color: #fca5a5;
}

.error-box p {
  margin: 0;
  font-size: 13px;
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

@media (max-width: 640px) {
  .progress-content {
    padding: 18px;
  }
}
</style>
