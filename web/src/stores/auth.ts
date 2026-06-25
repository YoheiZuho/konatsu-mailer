// SPDX-License-Identifier: Apache-2.0
//
// Authentication state: JWT access/refresh tokens persisted to localStorage.

import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { AuthTokens } from '@/lib/types';

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  expiresAt: string | null;
  setTokens: (tokens: AuthTokens) => void;
  clear: () => void;
}

export const useAuth = create<AuthState>()(
  persist(
    (set) => ({
      accessToken: null,
      refreshToken: null,
      expiresAt: null,
      setTokens: (t) =>
        set({
          accessToken: t.access_token,
          refreshToken: t.refresh_token,
          expiresAt: t.expires_at,
        }),
      clear: () => set({ accessToken: null, refreshToken: null, expiresAt: null }),
    }),
    { name: 'auth' },
  ),
);

/** Non-reactive selector for use outside React (e.g. the fetch client). */
export const authStore = useAuth;

export function isAuthenticated(): boolean {
  return !!useAuth.getState().accessToken;
}
