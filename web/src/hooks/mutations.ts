// SPDX-License-Identifier: Apache-2.0
//
// TanStack Query mutation hooks with optimistic updates (design doc §9.5).

import {
  useMutation,
  useQueryClient,
  type InfiniteData,
} from '@tanstack/react-query';
import { api } from '@/lib/api';
import { queryKeys } from '@/hooks/queries';
import type {
  Account,
  AccountInput,
  EmailListResponse,
  Label,
  LLMConfig,
  LLMConfigInput,
  LLMTestResult,
  SendEmailInput,
} from '@/lib/types';

/** Patch a single email row across every cached email list page. */
function patchEmailInLists(
  qc: ReturnType<typeof useQueryClient>,
  id: string,
  patch: Partial<EmailListResponse['items'][number]>,
) {
  qc.setQueriesData<InfiniteData<EmailListResponse>>({ queryKey: ['emails'] }, (data) => {
    if (!data) return data;
    return {
      ...data,
      pages: data.pages.map((page) => ({
        ...page,
        items: page.items.map((item) => (item.id === id ? { ...item, ...patch } : item)),
      })),
    };
  });
}

export function useMarkRead() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, is_read }: { id: string; is_read: boolean }) =>
      api.patch(`/emails/${id}/read`, { is_read }),
    onMutate: async ({ id, is_read }) => {
      await qc.cancelQueries({ queryKey: ['emails'] });
      patchEmailInLists(qc, id, { is_read });
    },
    onError: (_e, { id, is_read }) => patchEmailInLists(qc, id, { is_read: !is_read }),
  });
}

export function useToggleStar() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, is_starred }: { id: string; is_starred: boolean }) =>
      api.patch(`/emails/${id}/star`, { is_starred }),
    onMutate: async ({ id, is_starred }) => {
      await qc.cancelQueries({ queryKey: ['emails'] });
      patchEmailInLists(qc, id, { is_starred });
    },
    onError: (_e, { id, is_starred }) => patchEmailInLists(qc, id, { is_starred: !is_starred }),
  });
}

export function useAssignLabels() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, add, remove }: { id: string; add?: string[]; remove?: string[] }) =>
      api.post(`/emails/${id}/labels`, { add, remove }),
    onSuccess: (_d, { id }) => {
      qc.invalidateQueries({ queryKey: queryKeys.email(id) });
      qc.invalidateQueries({ queryKey: ['emails'] });
    },
  });
}

export function useReanalyze() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.post(`/emails/${id}/reanalyze`),
    onMutate: async (id) => {
      patchEmailInLists(qc, id, { analysis_status: 'pending' });
    },
  });
}

export function useSendEmail() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: SendEmailInput) => api.post('/emails/send', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['emails'] }),
  });
}

// --- Labels CRUD ---

export function useCreateLabel() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: Pick<Label, 'name' | 'color'>) => api.post<Label>('/labels', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.labels }),
  });
}

export function useUpdateLabel() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...patch }: Partial<Label> & { id: string }) =>
      api.patch<Label>(`/labels/${id}`, patch),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.labels }),
  });
}

export function useDeleteLabel() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete(`/labels/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.labels }),
  });
}

// --- Accounts CRUD ---

export function useCreateAccount() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: AccountInput) => api.post<Account>('/accounts', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.accounts }),
  });
}

export function useUpdateAccount() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...patch }: Partial<AccountInput> & { id: string }) =>
      api.patch<Account>(`/accounts/${id}`, patch),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.accounts }),
  });
}

export function useDeleteAccount() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete(`/accounts/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.accounts }),
  });
}

// --- LLM configs CRUD + connection test ---

export function useCreateLLMConfig() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: LLMConfigInput) => api.post<LLMConfig>('/llm-configs', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.llmConfigs }),
  });
}

export function useUpdateLLMConfig() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...patch }: Partial<LLMConfigInput> & { id: string }) =>
      api.patch<LLMConfig>(`/llm-configs/${id}`, patch),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.llmConfigs }),
  });
}

export function useDeleteLLMConfig() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete(`/llm-configs/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.llmConfigs }),
  });
}

export function useTestLLMConfig() {
  return useMutation({
    mutationFn: (id: string) => api.post<LLMTestResult>(`/llm-configs/${id}/test`),
  });
}
