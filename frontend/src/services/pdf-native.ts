import { Capacitor } from '@capacitor/core';
import { FileViewer } from '@capacitor/file-viewer';
import { Directory, Filesystem } from '@capacitor/filesystem';

function sanitizeFileName(name: string) {
  return name.replace(/[\\/:*?"<>|\s]+/g, '-').replace(/-+/g, '-').replace(/^-|-$/g, '');
}

function toBase64(arrayBuffer: ArrayBuffer) {
  const bytes = new Uint8Array(arrayBuffer);
  let binary = '';
  const chunkSize = 0x8000;

  for (let i = 0; i < bytes.length; i += chunkSize) {
    const chunk = bytes.subarray(i, i + chunkSize);
    binary += String.fromCharCode(...chunk);
  }

  return btoa(binary);
}

async function ensureDocumentPermission() {
  if (Capacitor.getPlatform() !== 'android') {
    return;
  }

  const status = await Filesystem.checkPermissions();
  if (status.publicStorage === 'granted') {
    return;
  }

  const next = await Filesystem.requestPermissions();
  if (next.publicStorage !== 'granted') {
    throw new Error('未获得文件存储权限，无法保存 PDF。');
  }
}

export async function saveRemotePdfToDevice(url: string, fileBaseName: string) {
  if (!url) {
    throw new Error('PDF 地址为空');
  }

  await ensureDocumentPermission();

  const response = await fetch(url, {
    headers: {
      Authorization: `Bearer ${localStorage.getItem('accessToken') || ''}`,
    },
  });

  if (!response.ok) {
    throw new Error(`下载 PDF 失败（${response.status}）`);
  }

  const blob = await response.blob();
  const buffer = await blob.arrayBuffer();
  const base64 = toBase64(buffer);

  const safeName = sanitizeFileName(fileBaseName || 'latex-qapp-pdf') || 'latex-qapp-pdf';
  const fileName = `${safeName}.pdf`;
  const path = `latex-qapp/pdfs/${fileName}`;

  await Filesystem.writeFile({
    path,
    data: base64,
    directory: Directory.Documents,
    recursive: true,
  });

  const { uri } = await Filesystem.getUri({
    path,
    directory: Directory.Documents,
  });

  return {
    path,
    uri,
    fileName,
  };
}

export async function openPdfFromLocalUri(uri: string) {
  if (!uri) {
    throw new Error('文件路径为空');
  }

  try {
    await FileViewer.openDocumentFromLocalPath({ path: uri });
  } catch {
    const withoutFileScheme = uri.replace(/^file:\/\//i, '');
    await FileViewer.openDocumentFromLocalPath({ path: withoutFileScheme });
  }
}
