import type { ApiResponse, Consumer, LoginResponse, Role, Secret } from '../types';

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
  token?: string;
}

async function api<T>(path: string, opts: ApiOptions): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (opts.token) {
    headers['Authorization'] = `Bearer ${opts.token}`;
  }

  const res = await fetch(path, {
    method: opts.method ?? 'GET',
    headers,
    body: opts.body === undefined ? undefined : JSON.stringify(opts.body),
  });

  const payload = (await res.json().catch(() => null)) as ApiResponse<T> | null;
  const message = payload?.message ?? res.statusText;

  if (!res.ok) {
    throw new ApiError(res.status, message);
  }

  return payload?.data as T;
}

export function login(username: string, password: string): Promise<LoginResponse> {
  return api<LoginResponse>('/auth/login', {
    method: 'POST',
    body: { username, password },
  });
}

export function fetchConsumers(token: string): Promise<Consumer[]> {
  return api<Consumer[]>('/management/consumers', { token });
}

export function createConsumer(token: string, body: { name: string; apikey: string; roleId: string }): Promise<Consumer> {
  return api<Consumer>('/management/consumers', { method: 'POST', token, body });
}

export function updateConsumer(
  token: string,
  id: string,
  patch: { name?: string; apikey?: string; roleId?: string; active?: boolean },
): Promise<Consumer> {
  return api<Consumer>(`/management/consumers/${id}`, { method: 'PUT', token, body: patch });
}

export function deleteConsumer(token: string, id: string): Promise<void> {
  return api<void>(`/management/consumers/${id}`, { method: 'DELETE', token });
}

export function fetchRoles(token: string): Promise<Role[]> {
  return api<Role[]>('/management/roles', { token });
}

export function fetchSecrets(token: string, consumerId: string): Promise<Secret[]> {
  return api<Secret[]>(`/management/secrets?consumerId=${encodeURIComponent(consumerId)}`, { token });
}

export function createSecret(
  token: string,
  consumerId: string,
  key: string,
  value: string,
  isSecret: boolean,
): Promise<Secret> {
  return api<Secret>('/management/secrets', {
    method: 'POST',
    token,
    body: { consumerId, key, value, isSecret },
  });
}

export function updateSecret(
  token: string,
  id: string,
  patch: { key?: string; value?: string; isSecret?: boolean },
): Promise<Secret> {
  return api<Secret>(`/management/secrets/${id}`, {
    method: 'PUT',
    token,
    body: patch,
  });
}

export function deleteSecret(token: string, id: string): Promise<void> {
  return api<void>(`/management/secrets/${id}`, { method: 'DELETE', token });
}