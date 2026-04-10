import { apiClient } from '@/services/api';
import { Camera, CameraResultType, CameraSource } from '@capacitor/camera';
import { Capacitor } from '@capacitor/core';
import type { VisionLatexDraft, VisionStreamEvent } from '@/types/domain';

const DRAFT_STORAGE_KEY = 'draft_record_from_camera';

export async function analyzeLatex(recordId: number, latexSource: string) {
  const { data } = await apiClient.post('/api/ai/analyze', {
    record_id: recordId,
    latex_source: latexSource,
  });
  return data;
}

export async function getAnalysis(recordId: number) {
  const { data } = await apiClient.get(`/api/ai/analysis/${recordId}`);
  return data;
}

export async function capturePhotoAsBase64(): Promise<string> {
	if (!Capacitor.isNativePlatform()) {
		return pickImageFromFileAsBase64(true);
	}

  const photo = await Camera.getPhoto({
    quality: 85,
    source: CameraSource.Camera,
    resultType: CameraResultType.Base64,
    allowEditing: false,
  });

  if (!photo.base64String) {
    throw new Error('未获取到照片数据');
  }

  return photo.base64String;
}

export async function pickImageAsBase64(source: 'camera' | 'album' | 'file' = 'camera'): Promise<string> {
  if (!Capacitor.isNativePlatform()) {
    return pickImageFromFileAsBase64(source === 'camera');
  }

  const nativeSource = source === 'camera' ? CameraSource.Camera : CameraSource.Photos;
  const photo = await Camera.getPhoto({
    quality: 85,
    source: nativeSource,
    resultType: CameraResultType.Base64,
    allowEditing: false,
  });

  if (!photo.base64String) {
    throw new Error('未获取到照片数据');
  }

  return photo.base64String;
}

export async function generateLatexDraftByVision(imageBase64: string): Promise<VisionLatexDraft> {
  const response = await apiClient.post('/api/ai/vision/latex', {
    image_base64: imageBase64,
    include_solution: false,
  });

  const payload = response?.data?.data ?? response?.data ?? {};
  return normalizeDraftFromPayload(payload);
}

export async function generateLatexDraftByVisionStream(
  imageBase64: string,
  onEvent: (evt: VisionStreamEvent) => void,
): Promise<VisionLatexDraft> {
  const token = localStorage.getItem('accessToken') || '';
  const url = `${apiClient.defaults.baseURL}/api/ai/vision/latex/stream`;

  const res = await fetch(url, {
    method: 'POST',
    headers: {
      Accept: 'text/event-stream',
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify({ image_base64: imageBase64, include_solution: false }),
  });

  if (!res.ok || !res.body) {
    throw new Error(`SSE request failed: ${res.status}`);
  }

  let finalDraft: VisionLatexDraft | null = null;

  await readSSEStream(res, (raw) => {
    try {
      const evt = JSON.parse(raw) as VisionStreamEvent;
      onEvent(evt);

      if (evt.error) {
        throw new Error(evt.error);
      }

      if (evt.stage === 'final' || evt.done) {
        const subject = normalizeSubjectLabel(evt.subject);
        const questionType = normalizeQuestionTypeLabel(evt.question_type);
        const latexQuestion = assembleLatexFromQuestionJson(evt.question_json);
        finalDraft = {
          questionJson: evt.question_json,
          latexSource: String(evt.latex_source ?? ''),
          latexQuestion,
          latexAnswer: evt.latex_answer ?? '',
          latexSolution: evt.latex_solution ?? '',
          tags: evt.tags ?? inferTags(latexQuestion),
          subject,
          title: evt.title ?? mapTitleFromType(questionType),
          questionType,
        };
      }
    } catch (err: any) {
      throw new Error(err?.message || 'failed to parse SSE event');
    }
  });

  if (!finalDraft) {
    throw new Error('SSE finished without final event');
  }

  return finalDraft;
}

export async function generateSolutionByLatex(payload: {
  latexQuestion: string;
  latexSource?: string;
  questionType?: string;
  subject?: string;
}): Promise<{ latexAnswer: string; latexSolution: string }> {
  const { data } = await apiClient.post('/api/ai/solve', {
    latex_question: payload.latexQuestion,
    latex_source: payload.latexSource,
    question_type: normalizeQuestionTypeLabel(payload.questionType),
    subject: normalizeSubjectLabel(payload.subject),
  });

  const result = data?.data ?? {};
  return {
    latexAnswer: result?.latex_answer ?? '',
    latexSolution: result?.latex_solution ?? '',
  };
}

export async function generateSolutionByLatexStream(
  payload: { latexQuestion?: string; latexSource?: string; questionType?: string; subject?: string },
  onEvent: (evt: { stage: string; latex_answer?: string; latex_solution?: string; done?: boolean; error?: string }) => void,
): Promise<{ latexAnswer: string; latexSolution: string }> {
  const token = localStorage.getItem('accessToken') || '';
  const url = `${apiClient.defaults.baseURL}/api/ai/solve/stream`;

  const res = await fetch(url, {
    method: 'POST',
    headers: {
      Accept: 'text/event-stream',
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify({
      latex_question: payload.latexQuestion,
      latex_source: payload.latexSource,
      question_type: normalizeQuestionTypeLabel(payload.questionType),
      subject: normalizeSubjectLabel(payload.subject),
    }),
  });

  if (!res.ok || !res.body) {
    throw new Error(`SSE request failed: ${res.status}`);
  }

  let latexAnswer = '';
  let latexSolution = '';

  await readSSEStream(res, (raw) => {
    const evt = JSON.parse(raw) as { stage: string; latex_answer?: string; latex_solution?: string; done?: boolean; error?: string };
    onEvent(evt);

    if (evt.error) {
      throw new Error(evt.error);
    }

    if (evt.latex_answer) {
      latexAnswer = evt.latex_answer;
    }
    if (evt.latex_solution) {
      latexSolution = evt.latex_solution;
    }
  });

  return { latexAnswer, latexSolution };
}

export function saveVisionDraftToStorage(draft: VisionLatexDraft) {
  localStorage.setItem(DRAFT_STORAGE_KEY, JSON.stringify(draft));
}

export function loadVisionDraftFromStorage(): VisionLatexDraft | null {
  const raw = localStorage.getItem(DRAFT_STORAGE_KEY);
  if (!raw) {
    return null;
  }

  try {
    return JSON.parse(raw) as VisionLatexDraft;
  } catch {
    return null;
  }
}

export function clearVisionDraftStorage() {
  localStorage.removeItem(DRAFT_STORAGE_KEY);
}

function normalizeDraftFromPayload(payload: any): VisionLatexDraft {
  const questionJson = payload?.question_json;
  const parsedLatex = assembleLatexFromQuestionJson(questionJson);
  const parsedType = normalizeQuestionTypeLabel(payload?.question_type);
  return {
    questionJson,
    latexSource: String(payload?.latex_source ?? ''),
    latexQuestion: parsedLatex ?? '',
    latexAnswer: payload?.latex_answer || '',
    latexSolution: payload?.latex_solution || '',
    tags: Array.isArray(payload?.tags) ? payload.tags : inferTags(parsedLatex),
    subject: normalizeSubjectLabel(payload?.subject),
    title: payload?.title ?? mapTitleFromType(parsedType),
    questionType: parsedType,
  };
}

function assembleLatexFromQuestionJson(questionJson: any): string {
  if (!questionJson || typeof questionJson !== 'object') {
    return '';
  }

  const questionType = normalizeQuestionTypeLabel(questionJson.question_type);
  const stem = String(questionJson.stem ?? '').trim();
  if (!stem) {
    return '';
  }

  if (questionType === '选择') {
    const options = Array.isArray(questionJson.options)
      ? questionJson.options.map((it: any) => String(it ?? '').trim()).filter(Boolean)
      : [];
    if (!options.length) {
      return stem;
    }
    return `${stem}\n\\begin{choices}\n${options.map((opt: string) => `\\item ${opt}`).join('\n')}\n\\end{choices}`;
  }

  if (questionType === '解答') {
    const subQuestions = Array.isArray(questionJson.sub_questions)
      ? questionJson.sub_questions.map((it: any) => String(it ?? '').trim()).filter(Boolean)
      : [];
    if (!subQuestions.length) {
      return stem;
    }
    return `${stem}\n\\begin{enumerate}\n${subQuestions.map((sq: string) => `\\item ${sq}`).join('\n')}\n\\end{enumerate}`;
  }

  return stem;
}

function extractLatexCode(text: string): string {
  if (!text) {
    return '';
  }

  const match = text.match(/```latex\s*([\s\S]*?)```/i);
  if (match && match[1]) {
    return match[1].trim();
  }
  return text.trim();
}

function inferQuestionType(latex: string): string {
  const content = latex.toLowerCase();
  if (content.includes('\\begin{choices}')) {
    return '选择';
  }
  if (content.includes('\\fillin')) {
    return '填空';
  }
  if (content.includes('\\begin{enumerate}')) {
    return '解答';
  }
  return '未知';
}

function inferTags(latex: string): string[] {
  return ['exam-zh', inferQuestionType(latex)];
}

function mapTitleFromType(questionType: string): string {
  switch (questionType) {
    case '选择':
      return '选择题';
    case '填空':
      return '填空题';
    case '解答':
      return '大题';
    default:
      return '识别题目';
  }
}

function normalizeSubjectLabel(input?: string): string {
  const value = String(input || '').trim().toLowerCase();
  if (value === 'math' || value === '数学') return '数学';
  if (value === 'physics' || value === '物理') return '物理';
  if (value === 'chemistry' || value === '化学') return '化学';
  if (value === 'biology' || value === '生物') return '生物';
  if (!value || value === 'unknown' || value === '未知') return '未知';
  return String(input || '未知');
}

function normalizeQuestionTypeLabel(input?: string): string {
  const value = String(input || '').trim().toLowerCase();
  if (['choice', '选择', '选择题', 'single_choice', 'multiple_choice'].includes(value)) return '选择';
  if (['fill_blank', '填空', '填空题'].includes(value)) return '填空';
  if (['essay', '解答', '解答题', 'subjective', '大题'].includes(value)) return '解答';
  if (!value || value === 'unknown' || value === '未知') return '未知';
  return String(input || '未知');
}

function pickImageFromFileAsBase64(preferCamera: boolean): Promise<string> {
  return new Promise((resolve, reject) => {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = 'image/*';
    if (preferCamera) {
      input.setAttribute('capture', 'environment');
    }

    input.onchange = async () => {
      const file = input.files?.[0];
      if (!file) {
        reject(new Error('未选择图片'));
        return;
      }

      try {
        const data = await fileToBase64(file);
        resolve(data);
      } catch (error) {
        reject(error);
      }
    };

    input.click();
  });
}

function fileToBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => {
      const result = String(reader.result || '');
      const idx = result.indexOf(',');
      resolve(idx >= 0 ? result.slice(idx + 1) : result);
    };
    reader.onerror = () => reject(new Error('读取文件失败'));
    reader.readAsDataURL(file);
  });
}

type SSEMessage = {
  event: string;
  data: string;
};

async function readSSEStream(response: Response, onMessage: (data: string, message: SSEMessage) => void | Promise<void>) {
  if (!response.body) {
    throw new Error('SSE response body is empty');
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder('utf-8');
  let buffer = '';
  let eventName = 'message';
  let dataLines: string[] = [];

  const emitMessage = async () => {
    if (!dataLines.length) {
      eventName = 'message';
      return;
    }

    const data = dataLines.join('\n').trim();
    const message: SSEMessage = { event: eventName, data };

    dataLines = [];
    eventName = 'message';

    if (!data) {
      return;
    }

    await onMessage(data, message);
  };

  while (true) {
    const { done, value } = await reader.read();
    if (done) {
      break;
    }

    buffer += decoder.decode(value, { stream: true });

    let lineEnd = buffer.indexOf('\n');
    while (lineEnd >= 0) {
      let line = buffer.slice(0, lineEnd);
      buffer = buffer.slice(lineEnd + 1);

      if (line.endsWith('\r')) {
        line = line.slice(0, -1);
      }

      if (line === '') {
        await emitMessage();
      } else if (line.startsWith(':')) {
        // Keep-alive/comment line.
      } else if (line.startsWith('event:')) {
        eventName = line.slice(6).trim() || 'message';
      } else if (line.startsWith('data:')) {
        dataLines.push(line.slice(5).trimStart());
      }

      lineEnd = buffer.indexOf('\n');
    }
  }

  buffer += decoder.decode();
  if (buffer.trim()) {
    const trailing = buffer.replace(/\r/g, '').split('\n');
    for (const line of trailing) {
      if (line.startsWith('event:')) {
        eventName = line.slice(6).trim() || 'message';
      } else if (line.startsWith('data:')) {
        dataLines.push(line.slice(5).trimStart());
      }
    }
  }

  await emitMessage();
}
