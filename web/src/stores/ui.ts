// SPDX-License-Identifier: Apache-2.0
//
// Transient UI state (not persisted): current folder/label filter, search,
// selection, compose dialog, settings dialog, and live sync status.

import { create } from 'zustand';
import type { SyncState } from '@/lib/types';

export type Folder = 'INBOX' | 'STARRED' | 'IMPORTANT' | 'SENT' | 'DRAFTS' | 'SPAM' | 'TRASH';

export interface ComposeState {
  open: boolean;
  mode: 'compose' | 'reply' | 'forward';
  to: string;
  cc: string;
  bcc: string;
  subject: string;
  body: string;
  inReplyTo?: string;
  threadId?: string;
}

const emptyCompose: ComposeState = {
  open: false,
  mode: 'compose',
  to: '',
  cc: '',
  bcc: '',
  subject: '',
  body: '',
};

interface UIState {
  folder: Folder;
  labelFilter: string | null;
  search: string;
  unreadOnly: boolean;
  sidebarOpen: boolean;
  settingsOpen: boolean;
  compose: ComposeState;
  syncStatus: SyncState;

  setFolder: (f: Folder) => void;
  setLabelFilter: (label: string | null) => void;
  setSearch: (q: string) => void;
  setUnreadOnly: (v: boolean) => void;
  toggleSidebar: () => void;
  setSettingsOpen: (v: boolean) => void;
  setSyncStatus: (s: SyncState) => void;

  openCompose: (init?: Partial<ComposeState>) => void;
  updateCompose: (patch: Partial<ComposeState>) => void;
  closeCompose: () => void;
}

export const useUI = create<UIState>((set) => ({
  folder: 'INBOX',
  labelFilter: null,
  search: '',
  unreadOnly: false,
  sidebarOpen: true,
  settingsOpen: false,
  compose: emptyCompose,
  syncStatus: 'connected',

  setFolder: (f) => set({ folder: f, labelFilter: null }),
  setLabelFilter: (label) => set({ labelFilter: label, folder: 'INBOX' }),
  setSearch: (q) => set({ search: q }),
  setUnreadOnly: (v) => set({ unreadOnly: v }),
  toggleSidebar: () => set((s) => ({ sidebarOpen: !s.sidebarOpen })),
  setSettingsOpen: (v) => set({ settingsOpen: v }),
  setSyncStatus: (s) => set({ syncStatus: s }),

  openCompose: (init) =>
    set({ compose: { ...emptyCompose, ...init, open: true } }),
  updateCompose: (patch) => set((s) => ({ compose: { ...s.compose, ...patch } })),
  closeCompose: () => set({ compose: emptyCompose }),
}));
