// SPDX-License-Identifier: Apache-2.0
//
// TanStack Query read hooks. Query keys follow design doc §9.4.

import {
  useInfiniteQuery,
  useQuery,
  type InfiniteData,
} from '@tanstack/react-query';
import { useShallow } from 'zustand/react/shallow';
import { api } from '@/lib/api';
import { useUI, type MailView } from '@/stores/ui';
import type {
  Account,
  EmailListItem,
  EmailListResponse,
  FoldersResponse,
  Label,
  LLMConfig,
  Preferences,
  ThreadDetail,
} from '@/lib/types';

export interface EmailFilter {
  folder: string;
  view: MailView;
  label: string | null;
  q: string;
  unread: boolean;
}

export const queryKeys = {
  emails: (f: EmailFilter) => ['emails', f] as const,
  email: (id: string) => ['email', id] as const,
  folders: ['folders'] as const,
  labels: ['labels'] as const,
  accounts: ['accounts'] as const,
  llmConfigs: ['llm-configs'] as const,
  preferences: ['preferences'] as const,
};

function buildEmailQuery(f: EmailFilter, cursor?: string | null): string {
  const params = new URLSearchParams();
  if (f.label) {
    params.set('label', f.label);
  } else if (f.view === 'starred') {
    params.set('starred', 'true');
  } else if (f.view === 'important') {
    params.set('important', 'true');
  } else {
    params.set('folder', f.folder);
  }
  if (f.q.trim()) params.set('q', f.q.trim());
  if (f.unread) params.set('unread', 'true');
  params.set('limit', '50');
  if (cursor) params.set('cursor', cursor);
  return params.toString();
}

/** Infinite (keyset-paginated) list of emails for the active filter. */
export function useEmails() {
  const filter: EmailFilter = useUI(
    useShallow((s) => ({
      folder: s.folder,
      view: s.view,
      label: s.labelFilter,
      q: s.search,
      unread: s.unreadOnly,
    })),
  );

  return useInfiniteQuery<
    EmailListResponse,
    Error,
    InfiniteData<EmailListResponse>,
    ReturnType<typeof queryKeys.emails>,
    string | null
  >({
    queryKey: queryKeys.emails(filter),
    initialPageParam: null,
    queryFn: ({ pageParam }) =>
      api.get<EmailListResponse>(`/emails?${buildEmailQuery(filter, pageParam)}`),
    getNextPageParam: (last) => last.next_cursor ?? undefined,
    // Keep the previous list visible while a new filter loads (no flash).
    placeholderData: (prev) => prev,
  });
}

/** Flatten infinite pages into a single array of rows. */
export function flattenEmails(data?: InfiniteData<EmailListResponse>): EmailListItem[] {
  return data?.pages.flatMap((p) => p.items) ?? [];
}

export function useThread(id: string | null) {
  return useQuery({
    queryKey: id ? queryKeys.email(id) : ['email', 'none'],
    queryFn: () => api.get<ThreadDetail>(`/emails/${id}`),
    enabled: !!id,
  });
}

export function useFolders() {
  return useQuery({
    queryKey: queryKeys.folders,
    queryFn: () => api.get<FoldersResponse>('/folders'),
    staleTime: 30_000,
  });
}

export function useLabels() {
  return useQuery({
    queryKey: queryKeys.labels,
    queryFn: () => api.get<{ items: Label[] }>('/labels').then((r) => r.items ?? []),
  });
}

export function useAccounts() {
  return useQuery({
    queryKey: queryKeys.accounts,
    queryFn: () => api.get<{ items: Account[] }>('/accounts').then((r) => r.items ?? []),
  });
}

export function useLLMConfigs() {
  return useQuery({
    queryKey: queryKeys.llmConfigs,
    queryFn: () => api.get<{ items: LLMConfig[] }>('/llm-configs').then((r) => r.items ?? []),
  });
}

export function usePreferences() {
  return useQuery({
    queryKey: queryKeys.preferences,
    queryFn: () => api.get<Preferences>('/me/preferences'),
  });
}
