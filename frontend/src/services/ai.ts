import { apiClient } from '@/services/api';
import { Camera, CameraResultType, CameraSource } from '@capacitor/camera';
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
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify({ image_base64: imageBase64, include_solution: false }),
  });

  if (!res.ok || !res.body) {
    throw new Error(`SSE request failed: ${res.status}`);
  }

  const reader = res.body.getReader();
  const decoder = new TextDecoder('utf-8');
  let buffer = '';
  let finalDraft: VisionLatexDraft | null = null;

  while (true) {
    const { done, value } = await reader.read();
    if (done) {
      break;
    }

    buffer += decoder.decode(value, { stream: true });
    const chunks = buffer.split('\n\n');
    buffer = chunks.pop() || '';

    for (const chunk of chunks) {
      const dataLine = chunk
        .split('\n')
        .find((line) => line.startsWith('data:'));

      if (!dataLine) {
        continue;
      }

      const raw = dataLine.slice(5).trim();
      if (!raw) {
        continue;
      }

      try {
        const evt = JSON.parse(raw) as VisionStreamEvent;
        onEvent(evt);

        if (evt.error) {
          throw new Error(evt.error);
        }

        if (evt.stage === 'final' || evt.done) {
          finalDraft = {
            latexQuestion: evt.latex_question ?? '',
            latexAnswer: evt.latex_answer ?? '',
            latexSolution: evt.latex_solution ?? '',
            tags: evt.tags ?? inferTags(evt.latex_question ?? ''),
            subject: evt.subject ?? 'math',
            title: evt.title ?? mapTitleFromType(evt.question_type ?? inferQuestionType(evt.latex_question ?? '')),
            questionType: evt.question_type ?? inferQuestionType(evt.latex_question ?? ''),
          };
        }
      } catch (err: any) {
        throw new Error(err?.message || 'failed to parse SSE event');
      }
    }
  }

  if (!finalDraft) {
    throw new Error('SSE finished without final event');
  }

  return finalDraft;
}

export async function generateSolutionByLatex(payload: {
  latexQuestion: string;
  questionType?: string;
  subject?: string;
}): Promise<{ latexAnswer: string; latexSolution: string }> {
  const { data } = await apiClient.post('/api/ai/solve', {
    latex_question: payload.latexQuestion,
    question_type: payload.questionType ?? 'unknown',
    subject: payload.subject ?? 'math',
  });

  const result = data?.data ?? {};
  return {
    latexAnswer: result?.latex_answer ?? '',
    latexSolution: result?.latex_solution ?? '',
  };
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
  const rawContent = typeof payload?.raw_content === 'string' ? payload.raw_content : '';
  const parsedLatex = payload?.latex_question || extractLatexCode(rawContent);
  const parsedType = payload?.question_type || inferQuestionType(parsedLatex);
  return {
    latexQuestion: parsedLatex ?? '',
    latexAnswer: payload?.latex_answer || '',
    latexSolution: payload?.latex_solution || '',
    tags: Array.isArray(payload?.tags) ? payload.tags : inferTags(parsedLatex),
    subject: payload?.subject ?? 'math',
    title: payload?.title ?? mapTitleFromType(parsedType),
    questionType: parsedType,
  };
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
    return 'choice';
  }
  if (content.includes('\\fillin')) {
    return 'fill_blank';
  }
  if (content.includes('\\begin{enumerate}')) {
    return 'essay';
  }
  return 'unknown';
}

function inferTags(latex: string): string[] {
  return ['exam-zh', inferQuestionType(latex)];
}

function mapTitleFromType(questionType: string): string {
  switch (questionType) {
    case 'choice':
      return '选择题';
    case 'fill_blank':
      return '填空题';
    case 'essay':
      return '大题';
    default:
      return '识别题目';
  }
}
