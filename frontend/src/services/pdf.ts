import { apiClient } from '@/services/api';

export async function exportPdf(filterPayload: Record<string, unknown>) {
  const { data } = await apiClient.post('/api/pdf/export', { filter_payload: filterPayload });
  return data;
}

export async function getPdfJob(jobId: string | number) {
  const { data } = await apiClient.get(`/api/pdf/jobs/${jobId}`);
  return data;
}
