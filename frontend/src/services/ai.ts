import { apiClient } from '@/services/api';

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
