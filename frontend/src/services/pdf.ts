import { apiClient } from '@/services/api';

export async function exportPdfByRecordIds(recordIds: number[]) {
  const { data } = await apiClient.post('/api/pdf/export', { record_ids: recordIds });
  return data;
}

export async function getPdfJob(jobId: string | number) {
  const { data } = await apiClient.get(`/api/pdf/jobs/${jobId}`);
  return data;
}
