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

interface MailListPaneProps {
  selectedId: string | null;
  onSelect: (id: string) => void;
  className?: string;
}

export function MailListPane({ selectedId, onSelect, className }: MailListPaneProps) {
  const density = useAppearance((s) => s.density);
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
  const rowHeight = density === 'compact' ? 56 : 84;
  const virtualizer = useVirtualizer({
    count: emails.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => rowHeight,
    overscan: 8,
  });

  // Infinite scroll: fetch the next page as the last row scrolls into view.
  const items = virtualizer.getVirtualItems();
  useEffect(() => {
    const last = items.at(-1);
    if (!last) return;
    if (
      last.index >= emails.length - 1 &&
      query.hasNextPage &&
      !query.isFetchingNextPage
    ) {
      query.fetchNextPage();
    }
  }, [items, emails.length, query]);

  const handleSelect = (id: string, isRead: boolean) => {
    if (!isRead) markRead.mutate({ id, is_read: true });
    onSelect(id);
  };

  return (
    <section className={clsx('flex min-w-0 flex-col border-r border-line bg-surface', className)}>
      {/* Toolbar */}
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
        {query.isFetching && !query.isFetchingNextPage ? (
          <Spinner size={18} />
        ) : (
          <button className="icon-btn-sm" onClick={() => query.refetch()} aria-label="更新">
            <Icon name="refresh" size={19} />
          </button>
        )}
      </div>

      {/* List */}
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

function FilterPill({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
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
