import { defineStore } from 'pinia';
import { computed, ref } from 'vue';
import { localDb } from '@/services/db';
import { createRecord, listRecords, type SaveRecordPayload } from '@/services/records';
import type { ErrorRecord } from '@/types/domain';

export const useRecordStore = defineStore('record', () => {
  const records = ref<ErrorRecord[]>(localDb.listRecords());
  const pendingCount = computed(() => records.value.filter((r) => r.syncStatus !== 'synced').length);
  const loading = ref(false);

  async function reload() {
    loading.value = true;
    try {
      records.value = await listRecords();
    } catch {
      records.value = localDb.listRecords();
    } finally {
      loading.value = false;
    }
  }

  async function save(payload: SaveRecordPayload) {
    loading.value = true;
    try {
      const created = await createRecord(payload);
      records.value = [created, ...records.value.filter((item) => item.id !== created.id)];
      return created;
    } catch {
      const now = Date.now();
      const fallback = localDb.saveRecord({
        id: now,
        userId: 0,
        subject: payload.subject,
        questionType: payload.question_type,
        difficulty: payload.difficulty ?? 3,
        title: payload.title,
        latexSource: payload.latex_source,
        latexAnswer: payload.latex_answer,
        questionTags: payload.question_tags,
        latexVersion: 1,
        latexRenderStatus: 'pending',
        mistakeReason: payload.mistake_reason,
        masteryLevel: 0,
        reviewCount: 0,
        syncStatus: 'pending',
        localVersion: now,
        serverVersion: 0,
        createdAt: new Date(now).toISOString(),
        updatedAt: new Date(now).toISOString(),
      });
      records.value = [fallback, ...records.value.filter((item) => item.id !== fallback.id)];
      return fallback;
    } finally {
      loading.value = false;
    }
  }

  return {
    records,
    pendingCount,
    loading,
    reload,
    save,
  };
});
