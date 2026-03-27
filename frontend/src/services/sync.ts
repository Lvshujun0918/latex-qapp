import { Network } from '@capacitor/network';

export async function isOnline() {
  const status = await Network.getStatus();
  return status.connected;
}

export async function syncNow() {
  const online = await isOnline();
  if (!online) {
    return { ok: false, message: '当前离线，无法同步' };
  }
  return { ok: true, message: '同步任务已触发' };
}
