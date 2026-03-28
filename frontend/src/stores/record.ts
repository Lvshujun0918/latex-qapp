import { defineStore } from 'pinia';
import { ref } from 'vue';
import { createRecord, listRecords, updateRecord, type SaveRecordPayload } from '@/services/records';
import type { ErrorRecord } from '@/types/domain';

export const useRecordStore = defineStore('record', () => {
  const records = ref<ErrorRecord[]>([]);
  const loading = ref(false);

  async function reload() {
    loading.value = true;
    try {
      records.value = await listRecords();
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
    } finally {
      loading.value = false;
    }
  }

  async function updateById(id: number, payload: SaveRecordPayload) {
    loading.value = true;
    try {
      const updated = await updateRecord(id, payload);
      records.value = records.value.map((item) => (item.id === id ? updated : item));
      return updated;
    } finally {
      loading.value = false;
    }
  }

  return {
    records,
    loading,
    reload,
    save,
    updateById,
  };
});
