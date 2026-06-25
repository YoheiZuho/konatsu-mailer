// SPDX-License-Identifier: Apache-2.0

import { useEffect, useRef } from 'react';
import clsx from 'clsx';
import { useVirtualizer } from '@tanstack/react-virtual';

import { useEmails, flattenEmails } from '@/hooks/queries';
import { useToggleStar, useMarkRead } from '@/hooks/mutations';
import { useUI } from '@/stores/ui';
import { useAppearance } from '@/stores/appearance';
import { MailRow } from '@/components/mail/MailRow';
import { Icon } from '@/components/common/Icon';
import { CenteredSpinner, EmptyState, InlineError, Spinner } from '@/components/common/Feedback';
import type { Density } from '@/lib/types';

interface MailListPaneProps {
  selectedId: string | null;
  onSelect: (id: string) => void;
  /** "wide" = full-width classic list (案A); "column" = narrow list beside the reading pane (案B). */
  variant: 'wide' | 'column';
  className?: string;
  style?: React.CSSProperties;
}

function rowHeight(variant: 'wide' | 'column', density: Density): number {
  if (variant === 'column') return 86;
  return density === 'compact' ? 42 : 78;
}

export function MailListPane({ selectedId, onSelect, variant, className, style }: MailListPaneProps) {
  const density = useAppearance((s) => s.density);
  const setDensity = useAppearance((s) => s.setDensity);
  const aiOn = useAppearance((s) => s.aiSummaries);
  const unreadOnly = useUI((s) => s.unreadOnly);
  const setUnreadOnly = useUI((s) => s.setUnreadOnly);
  const labelFilter = useUI((s) => s.labelFilter);
  const setLabelFilter = useUI((s) => s.setLabelFilter);

  const query = useEmails();
  const emails = flattenEmails(query.data);

  const toggleStar = useToggleStar();
  const markRead = useMarkRead();

  const parentRef = useRef<HTMLDivElement>(null);
  const virtualizer = useVirtualizer({
    count: emails.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => rowHeight(variant, density),
    overscan: 8,
  });
  const items = virtualizer.getVirtualItems();

  useEffect(() => {
    const last = items.at(-1);
    if (last && last.index >= emails.length - 1 && query.hasNextPage && !query.isFetchingNextPage) {
      query.fetchNextPage();
    }
  }, [items, emails.length, query]);

  const handleSelect = (id: string, isRead: boolean) => {
    if (!isRead) markRead.mutate({ id, is_read: true });
    onSelect(id);
  };

  return (
    <section className={clsx('flex min-w-0 flex-col bg-surface', className)} style={style}>
      {variant === 'wide' ? (
        <WideToolbar
          density={density}
          setDensity={setDensity}
          count={emails.length}
          fetching={query.isFetching && !query.isFetchingNextPage}
          onRefresh={() => query.refetch()}
        />
      ) : (
        <ColumnToolbar
          unreadOnly={unreadOnly}
          setUnreadOnly={setUnreadOnly}
          labelFilter={labelFilter}
          setLabelFilter={setLabelFilter}
          fetching={query.isFetching && !query.isFetchingNextPage}
          onRefresh={() => query.refetch()}
        />
      )}

      <div ref={parentRef} className="min-h-0 flex-1 overflow-y-auto">
        {query.isLoading ? (
          <CenteredSpinner />
        ) : query.isError ? (
          <div className="p-4">
            <InlineError message={query.error?.message ?? 'メールの取得に失敗しました'} />
          </div>
        ) : emails.length === 0 ? (
          <EmptyState
            icon="inbox"
            title="メールがありません"
            description="アカウントを追加して同期を開始するか、フィルタ条件を変更してください。"
          />
        ) : (
          <div style={{ height: virtualizer.getTotalSize(), position: 'relative', width: '100%' }}>
            {items.map((vi) => {
              const email = emails[vi.index];
              return (
                <div
                  key={email.id}
                  data-index={vi.index}
                  ref={virtualizer.measureElement}
                  style={{ position: 'absolute', top: 0, left: 0, width: '100%', transform: `translateY(${vi.start}px)` }}
                >
                  <MailRow
                    email={email}
                    selected={email.id === selectedId}
                    variant={variant}
                    density={density}
                    aiOn={aiOn}
                    onSelect={() => handleSelect(email.id, email.is_read)}
                    onToggleStar={() => toggleStar.mutate({ id: email.id, is_starred: !email.is_starred })}
                  />
                </div>
              );
            })}
          </div>
        )}
        {query.isFetchingNextPage && (
          <div className="flex justify-center py-4">
            <Spinner size={20} />
          </div>
        )}
      </div>
    </section>
  );
}

// 案B toolbar: filter pills + refresh.
function ColumnToolbar({
  unreadOnly,
  setUnreadOnly,
  labelFilter,
  setLabelFilter,
  fetching,
  onRefresh,
}: {
  unreadOnly: boolean;
  setUnreadOnly: (v: boolean) => void;
  labelFilter: string | null;
  setLabelFilter: (v: string | null) => void;
  fetching: boolean;
  onRefresh: () => void;
}) {
  return (
    <div className="flex h-12 flex-none items-center gap-2 border-b border-line px-4">
      <FilterPill active={!unreadOnly && !labelFilter} onClick={() => { setUnreadOnly(false); setLabelFilter(null); }}>
        すべて
      </FilterPill>
      <FilterPill active={unreadOnly} onClick={() => setUnreadOnly(!unreadOnly)}>
        未読
      </FilterPill>
      <FilterPill active={labelFilter === '重要'} onClick={() => setLabelFilter(labelFilter === '重要' ? null : '重要')}>
        重要
      </FilterPill>
      <div className="flex-1" />
      <RefreshButton fetching={fetching} onRefresh={onRefresh} />
    </div>
  );
}

// 案A toolbar: select-all (visual), refresh, count, density toggle.
function WideToolbar({
  density,
  setDensity,
  count,
  fetching,
  onRefresh,
}: {
  density: Density;
  setDensity: (d: Density) => void;
  count: number;
  fetching: boolean;
  onRefresh: () => void;
}) {
  return (
    <div className="flex h-12 flex-none items-center gap-1 border-b border-line pl-[18px] pr-4">
      <span className="flex cursor-default text-content-sub/60">
        <Icon name="check_box_outline_blank" size={20} />
      </span>
      <span className="mr-1.5 flex cursor-default text-content-sub/60">
        <Icon name="expand_more" size={18} />
      </span>
      <RefreshButton fetching={fetching} onRefresh={onRefresh} />
      <div className="flex-1" />
      {count > 0 && <span className="mr-2 text-[13px] text-content-sub">{count} 件</span>}
      <div className="mx-2 h-[22px] w-px bg-line" />
      <DensityToggle density={density} setDensity={setDensity} />
    </div>
  );
}

function DensityToggle({ density, setDensity }: { density: Density; setDensity: (d: Density) => void }) {
  return (
    <div className="flex items-center rounded-[9px] bg-surface-sub p-[3px]">
      <DensityBtn active={density === 'comfortable'} icon="view_agenda" onClick={() => setDensity('comfortable')} label="ゆったり" />
      <DensityBtn active={density === 'compact'} icon="density_small" onClick={() => setDensity('compact')} label="コンパクト" />
    </div>
  );
}

function DensityBtn({ active, icon, onClick, label }: { active: boolean; icon: string; onClick: () => void; label: string }) {
  return (
    <button
      onClick={onClick}
      aria-label={label}
      aria-pressed={active}
      className={clsx(
        'flex h-[30px] w-[38px] items-center justify-center rounded-[7px] transition-colors',
        active ? 'bg-surface text-content shadow-sm' : 'text-content-sub hover:text-content',
      )}
    >
      <Icon name={icon} size={19} />
    </button>
  );
}

function RefreshButton({ fetching, onRefresh }: { fetching: boolean; onRefresh: () => void }) {
  if (fetching) return <Spinner size={18} />;
  return (
    <button className="icon-btn-sm" onClick={onRefresh} aria-label="更新">
      <Icon name="refresh" size={19} />
    </button>
  );
}

function FilterPill({ active, onClick, children }: { active: boolean; onClick: () => void; children: React.ReactNode }) {
  return (
    <button
      onClick={onClick}
      className={clsx(
        'flex h-[30px] items-center rounded-full px-3.5 text-[12.5px] font-semibold transition-colors',
        active ? '' : 'bg-surface-sub text-content-sub hover:bg-hover',
      )}
      style={active ? { background: 'var(--brand)', color: 'var(--on-brand)' } : undefined}
    >
      {children}
    </button>
  );
}
