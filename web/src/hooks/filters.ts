// SPDX-License-Identifier: Apache-2.0
//
// Message-filter (auto-classification rule) hooks.

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { FilterRule, FilterRuleInput } from '@/lib/types';

const KEY = ['filters'] as const;

export function useFilters() {
  return useQuery({
    queryKey: KEY,
    queryFn: () => api.get<{ items: FilterRule[] }>('/filters').then((r) => r.items ?? []),
  });
}

export function useCreateFilter() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: FilterRuleInput) => api.post<FilterRule>('/filters', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEY }),
  });
}

export function useUpdateFilter() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...input }: FilterRuleInput & { id: string }) =>
      api.patch<FilterRule>(`/filters/${id}`, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEY }),
  });
}

export function useDeleteFilter() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete(`/filters/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEY }),
  });
}
