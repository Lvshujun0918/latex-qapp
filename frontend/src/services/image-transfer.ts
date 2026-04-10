const IMAGE_PAYLOAD_PREFIX = 'vision_image_payload:';

const inMemoryImagePayload = new Map<string, string>();

function safeSessionStorageSet(key: string, value: string) {
  try {
    sessionStorage.setItem(key, value);
  } catch {
    // Ignore storage quota and private mode errors; in-memory fallback still works.
  }
}

function safeSessionStorageGet(key: string): string | null {
  try {
    return sessionStorage.getItem(key);
  } catch {
    return null;
  }
}

function safeSessionStorageRemove(key: string) {
  try {
    sessionStorage.removeItem(key);
  } catch {
    // Ignore cleanup errors.
  }
}

function createImagePayloadKey() {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID();
  }

  return `img_${Date.now()}_${Math.random().toString(36).slice(2, 10)}`;
}

export function saveImagePayload(base64: string): string {
  const normalized = String(base64 || '').trim();
  if (!normalized) {
    throw new Error('图片数据为空，无法继续处理');
  }

  const key = createImagePayloadKey();
  const storageKey = `${IMAGE_PAYLOAD_PREFIX}${key}`;
  inMemoryImagePayload.set(storageKey, normalized);
  safeSessionStorageSet(storageKey, normalized);
  return key;
}

export function loadImagePayload(key: string): string | null {
  const normalizedKey = String(key || '').trim();
  if (!normalizedKey) {
    return null;
  }

  const storageKey = `${IMAGE_PAYLOAD_PREFIX}${normalizedKey}`;
  const memoryValue = inMemoryImagePayload.get(storageKey);
  if (memoryValue) {
    return memoryValue;
  }

  const storageValue = safeSessionStorageGet(storageKey);
  if (storageValue) {
    inMemoryImagePayload.set(storageKey, storageValue);
    return storageValue;
  }

  return null;
}

export function removeImagePayload(key: string) {
  const normalizedKey = String(key || '').trim();
  if (!normalizedKey) {
    return;
  }

  const storageKey = `${IMAGE_PAYLOAD_PREFIX}${normalizedKey}`;
  inMemoryImagePayload.delete(storageKey);
  safeSessionStorageRemove(storageKey);
}
