import { apiClient } from '@/services/api';
import type { ErrorRecord } from '@/types/domain';

export interface SaveRecordPayload {
  subject: string;
  question_type?: string;
  title?: string;
  latex_source: string;
  latex_answer?: string;
  question_tags?: string[];
  mistake_reason?: string;
}

export function toSavePayload(record: ErrorRecord): SaveRecordPayload {
  return {
    subject: record.subject,
    question_type: record.questionType,
    title: record.title,
    latex_source: record.latexSource,
    latex_answer: record.latexAnswer,
    question_tags: record.questionTags,
    mistake_reason: record.mistakeReason,
  };
}

export async function listRecords(): Promise<ErrorRecord[]> {
  const { data } = await apiClient.get('/api/records');
  const rows = data?.data ?? [];
  if (!Array.isArray(rows)) {
    return [];
  }
  return rows.map(normalizeRecord);
}

export async function createRecord(payload: SaveRecordPayload): Promise<ErrorRecord> {
  const { data } = await apiClient.post('/api/records', payload);
  return normalizeRecord(data?.data ?? {});
}

export async function updateRecord(id: number, payload: SaveRecordPayload): Promise<ErrorRecord> {
  const { data } = await apiClient.put(`/api/records/${id}`, payload);
  return normalizeRecord(data?.data ?? {});
}

export async function getRecord(id: number): Promise<ErrorRecord> {
  const { data } = await apiClient.get(`/api/records/${id}`);
  return normalizeRecord(data?.data ?? {});
}

export async function deleteRecord(id: number): Promise<void> {
  await apiClient.delete(`/api/records/${id}`);
}

function normalizeRecord(raw: any): ErrorRecord {
  const id = Number(raw?.id ?? 0);
  const userId = Number(raw?.userId ?? raw?.user_id ?? 0);

  return {
    id,
    userId,
    subject: String(raw?.subject ?? ''),
    questionType: raw?.questionType ?? raw?.question_type,
    title: raw?.title,
    latexSource: String(raw?.latexSource ?? raw?.latex_source ?? ''),
    latexAnswer: raw?.latexAnswer ?? raw?.latex_answer,
    questionTags: raw?.questionTags ?? raw?.question_tags,
    latexVersion: Number(raw?.latexVersion ?? raw?.latex_version ?? 1),
    latexRenderStatus: (raw?.latexRenderStatus ?? raw?.latex_render_status ?? 'pending') as ErrorRecord['latexRenderStatus'],
    mistakeReason: raw?.mistakeReason ?? raw?.mistake_reason,
    masteryLevel: Number(raw?.masteryLevel ?? raw?.mastery_level ?? 0),
    reviewCount: Number(raw?.reviewCount ?? raw?.review_count ?? 0),
    createdAt: String(raw?.createdAt ?? raw?.created_at ?? ''),
    updatedAt: String(raw?.updatedAt ?? raw?.updated_at ?? ''),
  };
}
