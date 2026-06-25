// SPDX-License-Identifier: Apache-2.0

interface PriorityBadgeProps {
  priority: number; // 1..5
}

/** Maps an AI priority (1..5) to a label + tint. 4–5 use the high-priority accent. */
export function PriorityBadge({ priority }: PriorityBadgeProps) {
  const { label, high } = describePriority(priority);
  if (!high) {
    return (
      <span className="inline-flex h-[18px] items-center rounded-[5px] bg-surface-sub px-2 text-[10.5px] font-semibold text-content-sub">
        重要度 {label}
      </span>
    );
  }
  return (
    <span
      className="inline-flex h-[18px] items-center rounded-[5px] px-2 text-[10.5px] font-semibold"
      style={{
        background: 'color-mix(in srgb, var(--prio-high) 14%, var(--surface))',
        color: 'var(--prio-high)',
      }}
    >
      重要度 {label}
    </span>
  );
}

export function describePriority(priority: number): { label: string; high: boolean } {
  if (priority >= 5) return { label: '緊急', high: true };
  if (priority >= 4) return { label: '高', high: true };
  if (priority === 3) return { label: '中', high: false };
  return { label: '低', high: false };
}
