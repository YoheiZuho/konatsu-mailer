// SPDX-License-Identifier: Apache-2.0
//
// A single mail-list row. Two layouts (design 案A / 案B):
//   - variant "wide"   : full-width classic rows (with star), comfortable/compact
//   - variant "column" : narrow 3-line rows for the reading-pane layout (no star)

import clsx from 'clsx';
import type { Density, EmailListItem } from '@/lib/types';
import { formatListDate } from '@/lib/format';
import { Avatar } from '@/components/common/Avatar';
import { LabelChip } from '@/components/common/LabelChip';
import { Icon } from '@/components/common/Icon';

interface MailRowProps {
  email: EmailListItem;
  selected: boolean;
  variant: 'wide' | 'column';
  density: Density;
  aiOn: boolean;
  onSelect: () => void;
  onToggleStar: () => void;
}

export function MailRow(props: MailRowProps) {
  if (props.variant === 'column') return <ColumnRow {...props} />;
  if (props.density === 'compact') return <WideCompactRow {...props} />;
  return <WideComfortableRow {...props} />;
}

function rowInteract(onSelect: () => void) {
  return {
    role: 'button' as const,
    tabIndex: 0,
    onClick: onSelect,
    onKeyDown: (e: React.KeyboardEvent) => {
      if (e.key === 'Enter' || e.key === ' ') {
        e.preventDefault();
        onSelect();
      }
    },
  };
}

function summaryOf(email: EmailListItem, aiOn: boolean): string {
  return (aiOn && email.ai_summary) || email.body_preview || '';
}

function StarButton({ starred, onToggle, size = 19 }: { starred: boolean; onToggle: () => void; size?: number }) {
  return (
    <button
      onClick={(e) => {
        e.stopPropagation();
        onToggle();
      }}
      className="flex flex-none text-content-sub/50 hover:text-brand"
      aria-label={starred ? 'スターを外す' : 'スターを付ける'}
    >
      <Icon name="star" size={size} filled={starred} style={starred ? { color: 'var(--brand)' } : undefined} />
    </button>
  );
}

// ── 案B: narrow column row ──────────────────────────────────────────────────
function ColumnRow({ email, selected, aiOn, onSelect }: MailRowProps) {
  const unread = !email.is_read;
  const sender = email.sender_name || email.sender_addr;
  return (
    <div
      {...rowInteract(onSelect)}
      className={clsx(
        'flex cursor-pointer items-start gap-[11px] border-b border-line/70 px-4 py-[11px] outline-none transition-colors',
        !selected && 'hover:bg-hover',
      )}
      style={selected ? { background: 'var(--brand-weak)', boxShadow: 'inset 3px 0 0 var(--brand)' } : undefined}
      aria-current={selected ? 'true' : undefined}
    >
      <Avatar name={sender} seed={email.sender_addr} size={34} />
      <div className="flex min-w-0 flex-1 flex-col gap-0.5">
        <div className="flex items-center gap-2">
          {unread && <span className="h-[7px] w-[7px] flex-none rounded-full" style={{ background: 'var(--brand)' }} />}
          <span className={clsx('min-w-0 flex-1 truncate text-[14px]', unread ? 'font-bold' : 'font-medium', 'text-content')}>
            {sender}
          </span>
          <span
            className={clsx('flex-none whitespace-nowrap text-[12px]', unread && 'font-semibold')}
            style={{ color: unread ? 'var(--brand-strong)' : 'var(--text-sub)' }}
          >
            {formatListDate(email.date_sent)}
          </span>
        </div>
        <div className={clsx('truncate text-[13px] text-content', unread && 'font-semibold')}>
          {email.subject || '(件名なし)'}
        </div>
        <div className="flex items-center gap-2">
          <div className="min-w-0 flex-1 truncate text-[12.5px] text-content-sub">{summaryOf(email, aiOn)}</div>
          <div className="flex flex-none items-center gap-1.5">
            {email.has_attachment && <Icon name="attach_file" size={15} className="text-content-sub" />}
            {email.labels.slice(0, 2).map((l) => (
              <LabelChip key={l.name} label={l} />
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

// ── 案A: full-width comfortable row ─────────────────────────────────────────
function WideComfortableRow({ email, aiOn, onSelect, onToggleStar }: MailRowProps) {
  const unread = !email.is_read;
  const sender = email.sender_name || email.sender_addr;
  return (
    <div
      {...rowInteract(onSelect)}
      className="group flex cursor-pointer items-start gap-[13px] border-b border-line/70 px-[18px] pb-3 pt-[11px] outline-none transition-colors hover:bg-hover hover:shadow-[inset_3px_0_0_var(--brand)]"
    >
      <div className="flex flex-none items-center gap-2 pt-1">
        <DecorCheckbox />
        <StarButton starred={email.is_starred} onToggle={onToggleStar} />
      </div>
      <Avatar name={sender} seed={email.sender_addr} size={38} />
      <div className="flex min-w-0 flex-1 flex-col gap-[3px]">
        <div className="flex items-center gap-2">
          {unread && <span className="h-[7px] w-[7px] flex-none rounded-full" style={{ background: 'var(--brand)' }} />}
          <span className={clsx('min-w-0 flex-1 truncate text-[14px] text-content', unread ? 'font-bold' : 'font-medium')}>
            {sender}
          </span>
          {email.has_attachment && <Icon name="attach_file" size={16} className="flex-none text-content-sub" />}
          <span
            className={clsx('flex-none whitespace-nowrap text-[12px]', unread && 'font-semibold')}
            style={{ color: unread ? 'var(--brand-strong)' : 'var(--text-sub)' }}
          >
            {formatListDate(email.date_sent)}
          </span>
        </div>
        <div className="flex items-center gap-2.5">
          <div className="min-w-0 flex-1 truncate text-[13.5px]">
            <span className={clsx('text-content', unread && 'font-semibold')}>{email.subject || '(件名なし)'}</span>
            {summaryOf(email, aiOn) && <span className="text-content-sub"> — {summaryOf(email, aiOn)}</span>}
          </div>
          <div className="flex flex-none items-center gap-1.5">
            {email.labels.slice(0, 3).map((l) => (
              <LabelChip key={l.name} label={l} />
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

// ── 案A: full-width compact row ─────────────────────────────────────────────
function WideCompactRow({ email, aiOn, onSelect, onToggleStar }: MailRowProps) {
  const unread = !email.is_read;
  const sender = email.sender_name || email.sender_addr;
  return (
    <div
      {...rowInteract(onSelect)}
      className="group flex h-[42px] cursor-pointer items-center gap-2.5 border-b border-line/70 px-[18px] outline-none transition-colors hover:bg-hover hover:shadow-[inset_3px_0_0_var(--brand)]"
    >
      <DecorCheckbox size={18} />
      <StarButton starred={email.is_starred} onToggle={onToggleStar} size={18} />
      {unread && <span className="h-[7px] w-[7px] flex-none rounded-full" style={{ background: 'var(--brand)' }} />}
      <span className={clsx('w-[168px] flex-none truncate text-[14px] text-content', unread ? 'font-bold' : 'font-medium')}>
        {sender}
      </span>
      <div className="min-w-0 flex-1 truncate text-[13.5px]">
        <span className={clsx('text-content', unread && 'font-semibold')}>{email.subject || '(件名なし)'}</span>
        {summaryOf(email, aiOn) && <span className="text-content-sub"> — {summaryOf(email, aiOn)}</span>}
      </div>
      <div className="flex flex-none items-center gap-1.5">
        {email.labels.slice(0, 2).map((l) => (
          <LabelChip key={l.name} label={l} />
        ))}
      </div>
      {email.has_attachment && <Icon name="attach_file" size={16} className="flex-none text-content-sub" />}
      <span
        className={clsx('flex-none whitespace-nowrap text-[12px]', unread && 'font-semibold')}
        style={{ color: unread ? 'var(--brand-strong)' : 'var(--text-sub)' }}
      >
        {formatListDate(email.date_sent)}
      </span>
    </div>
  );
}

// DecorCheckbox is a visual-only checkbox matching the design (bulk-select is
// not yet wired). It does not toggle selection.
function DecorCheckbox({ size = 19 }: { size?: number }) {
  return (
    <span onClick={(e) => e.stopPropagation()} className="flex flex-none cursor-default text-content-sub/45">
      <Icon name="check_box_outline_blank" size={size} />
    </span>
  );
}
