// SPDX-License-Identifier: Apache-2.0
//
// Email body translation via the backend's LibreTranslate proxy.

import { useMutation, useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { TranslateConfig, TranslateLanguage, TranslateResult } from '@/lib/types';

/** Whether translation is configured server-side, and the default target. */
export function useTranslateConfig() {
  return useQuery({
    queryKey: ['translate', 'config'],
    queryFn: () => api.get<TranslateConfig>('/translate/config'),
    staleTime: 5 * 60_000,
  });
}

/** Supported languages (only fetched when translation is enabled). */
export function useTranslateLanguages(enabled: boolean) {
  return useQuery({
    queryKey: ['translate', 'languages'],
    queryFn: () => api.get<TranslateLanguage[]>('/translate/languages'),
    enabled,
    staleTime: 60 * 60_000,
  });
}

export function useTranslate() {
  return useMutation({
    mutationFn: (input: { q: string; target: string; source?: string; format?: 'text' | 'html' }) =>
      api.post<TranslateResult>('/translate', input),
  });
}

/** Common fallback targets when the /languages list is unavailable. */
export const COMMON_TARGETS: ReadonlyArray<TranslateLanguage> = [
  { code: 'ja', name: '日本語' },
  { code: 'en', name: 'English' },
  { code: 'zh', name: '中文' },
  { code: 'ko', name: '한국어' },
  { code: 'es', name: 'Español' },
  { code: 'fr', name: 'Français' },
  { code: 'de', name: 'Deutsch' },
  { code: 'pt', name: 'Português' },
  { code: 'ru', name: 'Русский' },
];
