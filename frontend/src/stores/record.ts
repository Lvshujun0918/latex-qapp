import { defineStore } from 'pinia';
import { computed, ref } from 'vue';
import { localDb } from '@/services/db';
import type { ErrorRecord } from '@/types/domain';

export const useRecordStore = defineStore('record', () => {
  const records = ref<ErrorRecord[]>(localDb.listRecords());
  const pendingCount = computed(() => records.value.filter((r) => r.syncStatus !== 'synced').length);

  function reload() {
    records.value = localDb.listRecords();
  }

  function save(record: ErrorRecord) {
    localDb.saveRecord(record);
    reload();
  }

  return {
    records,
    pendingCount,
    reload,
    save,
  };
});
