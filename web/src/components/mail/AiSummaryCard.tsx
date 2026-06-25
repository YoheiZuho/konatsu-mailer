// SPDX-License-Identifier: Apache-2.0

import { Icon } from '@/components/common/Icon';
import { describePriority } from '@/components/common/PriorityBadge';

interface AiSummaryCardProps {
  summary: string;
  priority?: number | null;
}

/** Highlighted AI summary card shown above a thread (design doc §9.3). */
export function AiSummaryCard({ summary, priority }: AiSummaryCardProps) {
  const prio = priority ? describePriority(priority) : null;
  return (
    <div
      className="mb-6 flex gap-3 rounded-xl border p-4"
      style={{
        background: 'var(--brand-weak)',
        borderColor: 'color-mix(in srgb, var(--brand) 30%, transparent)',
      }}
    >
      <Icon name="auto_awesome" size={22} className="flex-none text-brand" />
      <div className="flex-1">
        <div className="mb-1.5 flex items-center gap-2.5">
          <span className="font-mono text-[10.5px] tracking-[0.08em] text-content-sub">AI 要約</span>
          {prio && (
            <span
              className="inline-flex h-[18px] items-center rounded-[5px] px-2 text-[10.5px] font-semibold"
              style={
                prio.high
                  ? {
                      background: 'color-mix(in srgb, var(--prio-high) 16%, var(--surface))',
                      color: 'var(--prio-high)',
                    }
                  : { background: 'var(--surface-sub)', color: 'var(--text-sub)' }
              }
            >
              重要度 {prio.label}
            </span>
          )}
        </div>
        <p className="text-[14px] leading-relaxed text-content">{summary}</p>
      </div>
    </div>
  );
}
