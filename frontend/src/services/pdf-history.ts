const PDF_HISTORY_PREFIX = 'pdfExportHistory';
const HISTORY_LIMIT = 40;

export interface PdfExportHistoryItem {
  jobId: string;
  pdfFileUrl: string;
  selectedCount: number;
  createdAt: string;
  source: 'errors' | 'review' | 'unknown';
}

export interface PdfHistoryUserIdentity {
  userId?: number | null;
  username?: string | null;
}

function getStorageKey(identity: PdfHistoryUserIdentity) {
  const idPart = identity.userId ? `id:${identity.userId}` : '';
  const namePart = (identity.username || '').trim().toLowerCase();
  const owner = idPart || (namePart ? `name:${namePart}` : 'guest');
  return `${PDF_HISTORY_PREFIX}:${owner}`;
}

function parseHistory(raw: string | null): PdfExportHistoryItem[] {
  if (!raw) {
    return [];
  }

  try {
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) {
      return [];
    }

    return parsed
      .map((item) => ({
        jobId: String(item?.jobId || '').trim(),
        pdfFileUrl: String(item?.pdfFileUrl || '').trim(),
        selectedCount: Number(item?.selectedCount || 0),
        createdAt: String(item?.createdAt || ''),
        source: item?.source === 'errors' || item?.source === 'review' ? item.source : 'unknown',
      }))
      .filter((item) => item.jobId && item.pdfFileUrl);
  } catch {
    return [];
  }
}

function inferCreatedAt(jobId: string): string {
  const match = String(jobId || '').match(/job-(\d{13})/);
  const ts = Number(match?.[1] || 0);
  if (!Number.isFinite(ts) || ts <= 0) {
    return new Date().toISOString();
  }
  return new Date(ts).toISOString();
}

export function getPdfExportHistory(identity: PdfHistoryUserIdentity): PdfExportHistoryItem[] {
  const key = getStorageKey(identity);
  const list = parseHistory(localStorage.getItem(key));
  return list.sort((a, b) => Date.parse(b.createdAt || '') - Date.parse(a.createdAt || ''));
}

export function upsertPdfExportHistory(
  identity: PdfHistoryUserIdentity,
  item: Partial<PdfExportHistoryItem> & Pick<PdfExportHistoryItem, 'jobId'>,
) {
  const key = getStorageKey(identity);
  const current = parseHistory(localStorage.getItem(key));
  const nextItem: PdfExportHistoryItem = {
    jobId: String(item.jobId || '').trim(),
    pdfFileUrl: String(item.pdfFileUrl || '').trim(),
    selectedCount: Math.max(0, Number(item.selectedCount || 0)),
    createdAt: String(item.createdAt || '').trim() || inferCreatedAt(item.jobId),
    source: item.source === 'errors' || item.source === 'review' ? item.source : 'unknown',
  };

  if (!nextItem.jobId || !nextItem.pdfFileUrl) {
    return;
  }

  const withoutDup = current.filter((entry) => entry.jobId !== nextItem.jobId);
  const next = [nextItem, ...withoutDup]
    .sort((a, b) => Date.parse(b.createdAt || '') - Date.parse(a.createdAt || ''))
    .slice(0, HISTORY_LIMIT);

  localStorage.setItem(key, JSON.stringify(next));
}

export function resolvePdfUrl(pdfFileUrl: string): string {
  if (/^https?:\/\//i.test(pdfFileUrl)) {
    return pdfFileUrl;
  }
  const base = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';
  return `${base}${pdfFileUrl}`;
}
