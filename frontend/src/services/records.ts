import { apiClient } from '@/services/api';
import type { ErrorRecord } from '@/types/domain';

export interface SaveRecordPayload {
  subject: string;
  question_type?: string;
  difficulty?: number;
  title?: string;
  latex_source: string;
  latex_answer?: string;
  question_tags?: string[];
  mistake_reason?: string;
}

export async function listRecords(): Promise<ErrorRecord[]> {
  const { data } = await apiClient.get('/api/records');
  return (data?.data ?? []) as ErrorRecord[];
}

export async function createRecord(payload: SaveRecordPayload): Promise<ErrorRecord> {
  const { data } = await apiClient.post('/api/records', payload);
  return data?.data as ErrorRecord;
}

export async function updateRecord(id: number, payload: SaveRecordPayload): Promise<ErrorRecord> {
  const { data } = await apiClient.put(`/api/records/${id}`, payload);
  return data?.data as ErrorRecord;
}

export async function getRecord(id: number): Promise<ErrorRecord> {
  const { data } = await apiClient.get(`/api/records/${id}`);
  return data?.data as ErrorRecord;
}
