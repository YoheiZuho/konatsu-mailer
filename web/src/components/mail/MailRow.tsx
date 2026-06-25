// SPDX-License-Identifier: Apache-2.0

import clsx from 'clsx';
import type { Density } from '@/lib/types';
import type { EmailListItem } from '@/lib/types';
import { formatListDate } from '@/lib/format';
import { Avatar } from '@/components/common/Avatar';
import { LabelChip } from '@/components/common/LabelChip';
import { Icon } from '@/components/common/Icon';

interface MailRowProps {
  email: EmailListItem;
  selected: boolean;
  density: Density;
  aiOn: boolean;
  onSelect: () => void;
  onToggleStar: () => void;
}

export function MailRow({ email, selected, density, aiOn, onSelect, onToggleStar }: MailRowProps) {
  const unread = !email.is_read;
  const sender = email.sender_name || email.sender_addr;
  const summary = aiOn && email.ai_summary ? email.ai_summary : email.body_preview;
  const compact = density === 'compact';

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={onSelect}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          onSelect();
        }
      }}
      className={clsx(
        'group flex cursor-pointer items-start gap-3 border-b border-line/70 px-4 outline-none transition-colors',
        compact ? 'py-2' : 'py-3',
        selected ? '' : 'hover:bg-hover',
      )}
      style={
        selected
          ? { background: 'var(--brand-weak)', boxShadow: 'inset 3px 0 0 var(--brand)' }
          : undefined
      }
      aria-current={selected ? 'true' : undefined}
    >
      <button
        onClick={(e) => {
          e.stopPropagation();
          onToggleStar();
        }}
        className="flex flex-none pt-0.5 text-content-sub/60 hover:text-brand"
        aria-label={email.is_starred ? 'スターを外す' : 'スターを付ける'}
      >
        <Icon
          name="star"
          size={19}
          filled={email.is_starred}
          style={email.is_starred ? { color: 'var(--brand)' } : undefined}
        />
      </button>

      <Avatar name={sender} seed={email.sender_addr} size={compact ? 32 : 36} />

      <div className="flex min-w-0 flex-1 flex-col gap-0.5">
        <div className="flex items-center gap-2">
          {unread && (
            <span className="h-[7px] w-[7px] flex-none rounded-full" style={{ background: 'var(--brand)' }} />
          )}
          <span
            className={clsx(
              'min-w-0 flex-1 truncate text-[14px]',
              unread ? 'font-bold text-content' : 'font-medium text-content',
            )}
          >
            {sender}
          </span>
          {email.has_attachment && (
            <Icon name="attach_file" size={16} className="flex-none text-content-sub" />
          )}
          <span
            className={clsx('flex-none whitespace-nowrap text-[12px]', unread ? 'font-semibold' : '')}
            style={{ color: unread ? 'var(--brand-strong)' : 'var(--text-sub)' }}
          >
            {formatListDate(email.date_sent)}
          </span>
        </div>

        <div
          className={clsx(
            'truncate text-[13px]',
            unread ? 'font-semibold text-content' : 'text-content',
          )}
        >
          {email.subject || '(件名なし)'}
        </div>

        <div className="flex items-center gap-2">
          <div className="min-w-0 flex-1 truncate text-[12.5px] text-content-sub">{summary}</div>
          <div className="flex flex-none items-center gap-1.5">
            {email.labels.slice(0, 2).map((l) => (
              <LabelChip key={l.name} label={l} />
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
