// SPDX-License-Identifier: Apache-2.0

import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useUI } from '@/stores/ui';
import { useAppearance } from '@/stores/appearance';
import { useAuth } from '@/stores/auth';
import { resolveTheme } from '@/lib/theme';
import { Icon } from '@/components/common/Icon';
import type { SyncState } from '@/lib/types';

const SYNC_LABEL: Record<SyncState, string> = {
  connected: '同期中',
  reconnecting: '再接続中',
  down: '切断',
};

export function TopBar() {
  const navigate = useNavigate();
  const toggleSidebar = useUI((s) => s.toggleSidebar);
  const setSettingsOpen = useUI((s) => s.setSettingsOpen);
  const syncStatus = useUI((s) => s.syncStatus);
  const search = useUI((s) => s.search);
  const setSearch = useUI((s) => s.setSearch);

  const theme = useAppearance((s) => s.theme);
  const setTheme = useAppearance((s) => s.setTheme);
  const clearAuth = useAuth((s) => s.clear);

  // Debounce the search input into the store to avoid refetch-per-keystroke.
  const [local, setLocal] = useState(search);
  useEffect(() => setLocal(search), [search]);
  useEffect(() => {
    const t = setTimeout(() => setSearch(local), 300);
    return () => clearTimeout(t);
  }, [local, setSearch]);

  const isDark = resolveTheme(theme) === 'dark';

  return (
    <header className="flex h-16 flex-none items-center gap-2 border-b border-line bg-surface px-2 sm:gap-3 sm:px-3.5">
      <button className="icon-btn" onClick={toggleSidebar} aria-label="メニューを開閉">
        <Icon name="menu" />
      </button>

      <div className="flex w-auto flex-none items-center gap-2.5 sm:w-[180px]">
        <div
          className="flex h-8 w-8 flex-none items-center justify-center rounded-[9px] bg-brand"
          style={{ color: 'var(--on-brand)' }}
        >
          <Icon name="mail" size={20} />
        </div>
        <div className="hidden text-[19px] font-semibold text-content sm:block">konatsu</div>
      </div>

      <label className="flex h-11 max-w-[660px] flex-1 items-center gap-3 rounded-3xl bg-surface-sub px-4 transition focus-within:bg-surface focus-within:shadow-[0_1px_5px_rgba(0,0,0,.13)]">
        <Icon name="search" size={22} className="text-content-sub" />
        <input
          value={local}
          onChange={(e) => setLocal(e.target.value)}
          placeholder="メールを検索"
          className="min-w-0 flex-1 bg-transparent text-[15px] text-content outline-none"
          aria-label="メールを検索"
        />
        {local && (
          <button
            className="flex items-center text-content-sub hover:text-content"
            onClick={() => setLocal('')}
            aria-label="検索をクリア"
          >
            <Icon name="close" size={20} />
          </button>
        )}
      </label>

      <div className="hidden flex-1 sm:block" />

      <SyncBadge state={syncStatus} />

      <button
        className="icon-btn"
        onClick={() => setTheme(isDark ? 'light' : 'dark')}
        aria-label={isDark ? 'ライトモードに切替' : 'ダークモードに切替'}
      >
        <Icon name={isDark ? 'light_mode' : 'dark_mode'} />
      </button>
      <button className="icon-btn" onClick={() => setSettingsOpen(true)} aria-label="設定">
        <Icon name="settings" />
      </button>
      <button
        className="flex h-9 w-9 flex-none items-center justify-center rounded-full bg-brand text-[14px] font-semibold"
        style={{ color: 'var(--on-brand)' }}
        title="ログアウト"
        aria-label="ログアウト"
        onClick={() => {
          clearAuth();
          navigate('/login', { replace: true });
        }}
      >
        <Icon name="logout" size={18} />
      </button>
    </header>
  );
}

function SyncBadge({ state }: { state: SyncState }) {
  const dotColor =
    state === 'connected'
      ? 'var(--brand)'
      : state === 'reconnecting'
        ? 'oklch(0.7 0.13 80)'
        : 'var(--prio-high)';
  return (
    <div
      className="hidden h-8 flex-none items-center gap-2 rounded-2xl px-3 sm:flex"
      style={{ background: 'var(--brand-weak)' }}
      title={`同期状態: ${SYNC_LABEL[state]}`}
    >
      <span
        className={state !== 'connected' ? 'animate-pulse' : undefined}
        style={{ width: 8, height: 8, borderRadius: '50%', background: dotColor }}
      />
      <span className="text-[12.5px] font-semibold" style={{ color: 'var(--text)' }}>
        {SYNC_LABEL[state]}
      </span>
    </div>
  );
}
