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
    const { data } = await apiClient.post('/api/ai/vision/latex', {
      image_base64: imageBase64,
    });

    return {
      latexQuestion: data?.latex_question ?? '',
      latexAnswer: data?.latex_answer ?? '',
      tags: Array.isArray(data?.tags) ? data.tags : [],
      subject: data?.subject,
      title: data?.title,
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
