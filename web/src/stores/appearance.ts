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
import type { AiFilters, Density, Preferences } from '@/lib/types';

export const DEFAULT_BRAND = '#ffd20a';

/** Default AI analysis filters (true = skip AI summary for that category). */
export const DEFAULT_AI_FILTERS: AiFilters = {
  promotions: true,
  social: true,
  newsletters: true,
  automated: true,
};

/** localStorage keys (mirrors the design doc). */
const LS = {
  theme: 'ui.theme',
  brand: 'ui.brand',
  density: 'ui.density',
  ai: 'ui.aiSummaries',
  translateTarget: 'ui.translateTarget',
  sidebarWidth: 'ui.sidebarWidth',
  listWidth: 'ui.listWidth',
  aiFilters: 'ui.aiFilters',
} as const;

function readAiFilters(): AiFilters {
  const raw = lsGet(LS.aiFilters);
  if (!raw) return { ...DEFAULT_AI_FILTERS };
  try {
    return { ...DEFAULT_AI_FILTERS, ...(JSON.parse(raw) as Partial<AiFilters>) };
  } catch {
    return { ...DEFAULT_AI_FILTERS };
  }
}

export const SIDEBAR_WIDTH = { default: 256, min: 200, max: 400 } as const;
export const LIST_WIDTH = { default: 430, min: 320, max: 680 } as const;

function clampNum(raw: string | null, def: number, min: number, max: number): number {
  const n = raw ? parseInt(raw, 10) : NaN;
  if (Number.isNaN(n)) return def;
  return Math.min(Math.max(n, min), max);
}

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
  /** Preferred target language for email translation (ISO code). */
  translateTarget: string;
  /** Per-category AI analysis filters (true = skip AI summary). */
  aiFilters: AiFilters;
  /** Resizable pane widths (px). */
  sidebarWidth: number;
  listWidth: number;
  setTheme: (t: Theme) => void;
  setBrand: (hex: string) => void;
  setDensity: (d: Density) => void;
  setAiSummaries: (on: boolean) => void;
  setAiFilters: (f: AiFilters) => void;
  setTranslateTarget: (code: string) => void;
  setSidebarWidth: (w: number) => void;
  setListWidth: (w: number) => void;
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
        ai_filters: s.aiFilters,
      })
      .catch(() => {
        /* preferences are non-critical; ignore transient failures */
      });
  }, 600);
}

function readInitial(): Pick<
  AppearanceState,
  'theme' | 'brand' | 'density' | 'aiSummaries' | 'aiFilters' | 'translateTarget' | 'sidebarWidth' | 'listWidth'
> {
  const theme = lsGet(LS.theme);
  const density = lsGet(LS.density);
  return {
    theme: theme === 'light' || theme === 'dark' || theme === 'system' ? theme : 'system',
    brand: lsGet(LS.brand) ?? DEFAULT_BRAND,
    density: density === 'compact' ? 'compact' : 'comfortable',
    aiSummaries: lsGet(LS.ai) !== 'false',
    aiFilters: readAiFilters(),
    translateTarget: lsGet(LS.translateTarget) ?? 'ja',
    sidebarWidth: clampNum(lsGet(LS.sidebarWidth), SIDEBAR_WIDTH.default, SIDEBAR_WIDTH.min, SIDEBAR_WIDTH.max),
    listWidth: clampNum(lsGet(LS.listWidth), LIST_WIDTH.default, LIST_WIDTH.min, LIST_WIDTH.max),
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
  setAiFilters: (f) => {
    set({ aiFilters: f });
    lsSet(LS.aiFilters, JSON.stringify(f));
    syncToServer(get);
  },
  setTranslateTarget: (code) => {
    set({ translateTarget: code });
    lsSet(LS.translateTarget, code);
  },
  setSidebarWidth: (w) => {
    const v = Math.min(Math.max(Math.round(w), SIDEBAR_WIDTH.min), SIDEBAR_WIDTH.max);
    set({ sidebarWidth: v });
    lsSet(LS.sidebarWidth, String(v));
  },
  setListWidth: (w) => {
    const v = Math.min(Math.max(Math.round(w), LIST_WIDTH.min), LIST_WIDTH.max);
    set({ listWidth: v });
    lsSet(LS.listWidth, String(v));
  },
  hydrateFromServer: (prefs) => {
    const aiFilters = prefs.ai_filters
      ? { ...DEFAULT_AI_FILTERS, ...prefs.ai_filters }
      : get().aiFilters;
    const next = {
      theme: prefs.theme ?? get().theme,
      brand: prefs.brand_color ?? get().brand,
      density: prefs.density ?? get().density,
      aiSummaries: prefs.ai_summaries ?? get().aiSummaries,
      aiFilters,
    };
    set(next);
    lsSet(LS.theme, next.theme);
    lsSet(LS.brand, next.brand);
    lsSet(LS.density, next.density);
    lsSet(LS.ai, String(next.aiSummaries));
    lsSet(LS.aiFilters, JSON.stringify(next.aiFilters));
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
