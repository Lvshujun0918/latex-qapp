import { apiClient } from '@/services/api';
import { Camera, CameraResultType, CameraSource } from '@capacitor/camera';
import type { VisionLatexDraft } from '@/types/domain';

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
  try {
    const response = await apiClient.post('/api/ai/vision/latex', {
      image_base64: imageBase64,
    });

    const payload = response?.data?.data ?? response?.data ?? {};
    const rawContent = typeof payload?.raw_content === 'string' ? payload.raw_content : '';
    const parsedLatex = payload?.latex_question || extractLatexCode(rawContent);
    const parsedAnswer = payload?.latex_answer || '';
    const parsedTags = Array.isArray(payload?.tags) ? payload.tags : inferTags(parsedLatex);
    const parsedType = payload?.question_type || inferQuestionType(parsedLatex);

    return {
      latexQuestion: parsedLatex ?? '',
      latexAnswer: parsedAnswer,
      tags: parsedTags,
      subject: payload?.subject ?? 'math',
      title: payload?.title ?? mapTitleFromType(parsedType),
    };
  } catch {
    return {
      latexQuestion: '',
      latexAnswer: '',
      tags: [],
    };
  }
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
