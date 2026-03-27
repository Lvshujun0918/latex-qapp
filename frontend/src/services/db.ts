import type { ErrorRecord } from '@/types/domain';

const RECORDS_KEY = 'error_records';

function readRecords(): ErrorRecord[] {
  const raw = localStorage.getItem(RECORDS_KEY);
  return raw ? (JSON.parse(raw) as ErrorRecord[]) : [];
}

function writeRecords(records: ErrorRecord[]) {
  localStorage.setItem(RECORDS_KEY, JSON.stringify(records));
}

export const localDb = {
  listRecords(): ErrorRecord[] {
    return readRecords();
  },
  saveRecord(record: ErrorRecord): ErrorRecord {
    const records = readRecords();
    const idx = records.findIndex((item) => item.id === record.id);
    if (idx >= 0) {
      records[idx] = record;
    } else {
      records.unshift(record);
    }
    writeRecords(records);
    return record;
  },
};
