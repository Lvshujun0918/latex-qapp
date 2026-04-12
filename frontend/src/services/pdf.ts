import { apiClient } from '@/services/api';

export async function exportPdfByRecordIds(recordIds: number[]) {
  const { data } = await apiClient.post('/api/pdf/export', { record_ids: recordIds });
  return data;
}

export async function listPdfJobs() {
  const { data } = await apiClient.get('/api/pdf/jobs');
  return data;
}

export async function getPdfJob(jobId: string | number) {
  const { data } = await apiClient.get(`/api/pdf/jobs/${jobId}`);
  return data;
}

export async function updatePdfQuestionReview(jobId: string | number, recordId: number, isCorrect: boolean) {
  const { data } = await apiClient.post(`/api/pdf/jobs/${jobId}/questions/${recordId}/review`, {
    is_correct: isCorrect,
  });
  return data;
}
