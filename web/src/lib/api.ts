// SPDX-License-Identifier: Apache-2.0
//
// Thin REST client with bearer auth and transparent one-shot token refresh.

import { authStore } from '@/stores/auth';
import type { AuthTokens } from '@/lib/types';

export const API_BASE = (import.meta.env.VITE_API_BASE_URL as string | undefined) ?? '/api';

/** Error thrown for non-2xx responses, carrying the parsed API error shape. */
export class ApiRequestError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
  ) {
    super(message);
    this.name = 'ApiRequestError';
  }
}

interface RequestOptions extends Omit<RequestInit, 'body'> {
  body?: unknown;
  /** Internal: prevents infinite refresh recursion. */
  _retry?: boolean;
}

// Coalesce concurrent refreshes into a single in-flight request.
let refreshInFlight: Promise<boolean> | null = null;

async function refreshTokens(): Promise<boolean> {
  const { refreshToken, setTokens, clear } = authStore.getState();
  if (!refreshToken) return false;

  if (!refreshInFlight) {
    refreshInFlight = (async () => {
      try {
        const res = await fetch(`${API_BASE}/auth/refresh`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refresh_token: refreshToken }),
        });
        if (!res.ok) {
          clear();
          return false;
        }
        const tokens = (await res.json()) as AuthTokens;
        setTokens(tokens);
        return true;
      } catch {
        return false;
      } finally {
        refreshInFlight = null;
      }
    })();
  }
  return refreshInFlight;
}

async function request<T>(method: string, path: string, opts: RequestOptions = {}): Promise<T> {
  const { body, headers, _retry, ...rest } = opts;
  const token = authStore.getState().accessToken;

  const finalHeaders = new Headers(headers);
  if (token) finalHeaders.set('Authorization', `Bearer ${token}`);
  if (body !== undefined && !(body instanceof FormData)) {
    finalHeaders.set('Content-Type', 'application/json');
  }

  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers: finalHeaders,
    body:
      body === undefined
        ? undefined
        : body instanceof FormData
          ? body
          : JSON.stringify(body),
    ...rest,
  });

  if (res.status === 401 && !_retry) {
    const refreshed = await refreshTokens();
    if (refreshed) return request<T>(method, path, { ...opts, _retry: true });
  }

  // A 401 that survives a refresh means the session is unusable (expired, or the
  // user no longer exists after a DB reset) — clear it so the app returns to login.
  if (res.status === 401 && !path.startsWith('/auth/')) {
    authStore.getState().clear();
  }

  if (!res.ok) {
    let code = 'http_error';
    let message = res.statusText;
    try {
      const data = await res.json();
      if (data?.error) {
        code = data.error.code ?? code;
        message = data.error.message ?? message;
      }
    } catch {
      // non-JSON error body; keep status text
    }
    throw new ApiRequestError(res.status, code, message);
  }

  if (res.status === 204) return undefined as T;
  const contentType = res.headers.get('Content-Type') ?? '';
  if (!contentType.includes('application/json')) return undefined as T;
  return (await res.json()) as T;
}

export const api = {
  get: <T>(path: string, opts?: RequestOptions) => request<T>('GET', path, opts),
  post: <T>(path: string, body?: unknown, opts?: RequestOptions) =>
    request<T>('POST', path, { ...opts, body }),
  patch: <T>(path: string, body?: unknown, opts?: RequestOptions) =>
    request<T>('PATCH', path, { ...opts, body }),
  put: <T>(path: string, body?: unknown, opts?: RequestOptions) =>
    request<T>('PUT', path, { ...opts, body }),
  delete: <T>(path: string, opts?: RequestOptions) => request<T>('DELETE', path, opts),
};

/** Build an absolute URL to an API path (used for SSE / WebSocket). */
export function apiUrl(path: string): string {
  if (API_BASE.startsWith('http')) return `${API_BASE}${path}`;
  return `${window.location.origin}${API_BASE}${path}`;
}
