// SPDX-License-Identifier: Apache-2.0
//
// Appearance preferences (theme, brand key color, density, AI summaries).
//
// Persisted to localStorage under individual keys — `ui.theme`, `ui.brand`,
// `ui.density`, `ui.aiSummaries` — exactly as the design doc specifies (§9.2.3 /
// §9.6). The pre-paint inline script in index.html reads `ui.theme` / `ui.brand`
// to prevent a flash of the wrong palette. Changes also sync to the server so
// preferences follow the user across devices.

import { create } from 'zustand';
import { applyBrand, applyTheme, type Theme } from '@/lib/theme';
import { api } from '@/lib/api';
import { authStore } from '@/stores/auth';
import type { Density, Preferences } from '@/lib/types';

export const DEFAULT_BRAND = '#ffd20a';

/** localStorage keys (mirrors the design doc). */
const LS = {
  theme: 'ui.theme',
  brand: 'ui.brand',
  density: 'ui.density',
  ai: 'ui.aiSummaries',
} as const;

function lsGet(key: string): string | null {
  try {
    return localStorage.getItem(key);
  } catch {
    return null;
  }
}
function lsSet(key: string, value: string): void {
  try {
    localStorage.setItem(key, value);
  } catch {
    /* storage unavailable (private mode / quota) — non-fatal */
  }
}

/** Preset key colors offered in settings (design doc §9.2.4). */
export const BRAND_PRESETS: ReadonlyArray<{ hex: string; name: string }> = [
  { hex: '#ffd20a', name: 'イエロー' },
  { hex: '#10a37f', name: 'エメラルド' },
  { hex: '#3b82f6', name: 'ブルー' },
  { hex: '#8b5cf6', name: 'バイオレット' },
  { hex: '#ef4444', name: 'レッド' },
  { hex: '#f97316', name: 'オレンジ' },
  { hex: '#ec4899', name: 'ピンク' },
  { hex: '#64748b', name: 'スレート' },
];

interface AppearanceState {
  theme: Theme;
  brand: string;
  density: Density;
  aiSummaries: boolean;
  setTheme: (t: Theme) => void;
  setBrand: (hex: string) => void;
  setDensity: (d: Density) => void;
  setAiSummaries: (on: boolean) => void;
  /** Apply server-stored preferences on login without re-syncing back. */
  hydrateFromServer: (prefs: Partial<Preferences>) => void;
}

let syncTimer: ReturnType<typeof setTimeout> | undefined;

/** Debounced best-effort persistence of appearance to the backend. */
function syncToServer(get: () => AppearanceState) {
  if (!authStore.getState().accessToken) return;
  clearTimeout(syncTimer);
  syncTimer = setTimeout(() => {
    const s = get();
    api
      .patch('/me/preferences', {
        theme: s.theme,
        brand_color: s.brand,
        density: s.density,
        ai_summaries: s.aiSummaries,
      })
      .catch(() => {
        /* preferences are non-critical; ignore transient failures */
      });
  }, 600);
}

function readInitial(): Pick<AppearanceState, 'theme' | 'brand' | 'density' | 'aiSummaries'> {
  const theme = lsGet(LS.theme);
  const density = lsGet(LS.density);
  return {
    theme: theme === 'light' || theme === 'dark' || theme === 'system' ? theme : 'system',
    brand: lsGet(LS.brand) ?? DEFAULT_BRAND,
    density: density === 'compact' ? 'compact' : 'comfortable',
    aiSummaries: lsGet(LS.ai) !== 'false',
  };
}

export const useAppearance = create<AppearanceState>((set, get) => ({
  ...readInitial(),
  setTheme: (t) => {
    set({ theme: t });
    lsSet(LS.theme, t);
    applyTheme(t);
    syncToServer(get);
  },
  setBrand: (hex) => {
    set({ brand: hex });
    lsSet(LS.brand, hex);
    applyBrand(hex);
    syncToServer(get);
  },
  setDensity: (d) => {
    set({ density: d });
    lsSet(LS.density, d);
    syncToServer(get);
  },
  setAiSummaries: (on) => {
    set({ aiSummaries: on });
    lsSet(LS.ai, String(on));
    syncToServer(get);
  },
  hydrateFromServer: (prefs) => {
    const next = {
      theme: prefs.theme ?? get().theme,
      brand: prefs.brand_color ?? get().brand,
      density: prefs.density ?? get().density,
      aiSummaries: prefs.ai_summaries ?? get().aiSummaries,
    };
    set(next);
    lsSet(LS.theme, next.theme);
    lsSet(LS.brand, next.brand);
    lsSet(LS.density, next.density);
    lsSet(LS.ai, String(next.aiSummaries));
    applyTheme(next.theme);
    applyBrand(next.brand);
  },
}));

/** Apply persisted appearance to the DOM at startup. */
export function initAppearance() {
  const s = useAppearance.getState();
  applyTheme(s.theme);
  applyBrand(s.brand);
}
