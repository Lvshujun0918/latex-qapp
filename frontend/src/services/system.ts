import { apiClient } from '@/services/api';

export interface BackendRuntimeInfo {
  status: string;
  version: string;
  go_version: string;
  started_at: string;
  server_time: string;
  uptime_seconds: number;
}

export interface BackendRuntimeResult {
  data: BackendRuntimeInfo;
  latencyMs: number;
}

export async function fetchBackendRuntime(): Promise<BackendRuntimeResult> {
  const start = performance.now();
  const { data } = await apiClient.get('/api/system/runtime');
  const end = performance.now();

  return {
    data,
    latencyMs: Math.max(0, Math.round(end - start)),
  };
}
