import type { ApiResponse, Consumer, ConsumerSecrets, Secret } from '../types';

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

interface ApiOptions {
  method?: string;
  body?: unknown;
  apikey: string;
}

async function api<T>(path: string, opts: ApiOptions): Promise<T> {
  const res = await fetch(path, {
    method: opts.method ?? 'GET',
    headers: {
      'Content-Type': 'application/json',
      'x-apikey': opts.apikey,
    },
    body: opts.body === undefined ? undefined : JSON.stringify(opts.body),
  });

  const payload = (await res.json().catch(() => null)) as ApiResponse<T> | null;
  const message = payload?.message ?? res.statusText;

  if (!res.ok) {
    throw new ApiError(res.status, message);
  }

  return payload?.data as T;
}

export function fetchGroupedSecrets(apikey: string): Promise<ConsumerSecrets[]> {
  return api<ConsumerSecrets[]>('/management/secrets', { apikey });
}

export function fetchConsumers(apikey: string): Promise<Consumer[]> {
  return api<Consumer[]>('/management/consumers', { apikey });
}

export function createSecret(
  apikey: string,
  consumerId: string,
  key: string,
  value: string,
): Promise<Secret> {
  return api<Secret>('/management/secrets', {
    method: 'POST',
    apikey,
    body: { consumerId, key, value },
  });
}

export function updateSecret(
  apikey: string,
  id: string,
  patch: { key?: string; value?: string },
): Promise<Secret> {
  return api<Secret>(`/management/secrets/${id}`, {
    method: 'PUT',
    apikey,
    body: patch,
  });
}

export function deleteSecret(apikey: string, id: string): Promise<void> {
  return api<void>(`/management/secrets/${id}`, { method: 'DELETE', apikey });
}
