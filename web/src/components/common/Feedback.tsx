// SPDX-License-Identifier: Apache-2.0
//
// Small shared feedback primitives: spinner, empty state, inline error.

import { Icon } from '@/components/common/Icon';

export function Spinner({ size = 22 }: { size?: number }) {
  return (
    <span
      className="inline-block animate-spin rounded-full border-2 border-line border-t-brand"
      style={{ width: size, height: size }}
      role="status"
      aria-label="読み込み中"
    />
  );
}

export function CenteredSpinner() {
  return (
    <div className="flex h-full w-full items-center justify-center p-10">
      <Spinner size={28} />
    </div>
  );
}

interface EmptyStateProps {
  icon: string;
  title: string;
  description?: string;
  action?: React.ReactNode;
}

export function EmptyState({ icon, title, description, action }: EmptyStateProps) {
  return (
    <div className="flex h-full w-full flex-col items-center justify-center gap-3 p-10 text-center">
      <Icon name={icon} size={44} className="text-content-sub opacity-50" />
      <div className="text-[15px] font-semibold text-content">{title}</div>
      {description && (
        <div className="max-w-xs text-[13px] leading-relaxed text-content-sub">{description}</div>
      )}
      {action}
    </div>
  );
}

export function InlineError({ message }: { message: string }) {
  return (
    <div
      className="flex items-center gap-2 rounded-lg px-3 py-2 text-[13px]"
      style={{ background: 'color-mix(in srgb, var(--prio-high) 12%, var(--surface))', color: 'var(--prio-high)' }}
      role="alert"
    >
      <Icon name="error" size={18} />
      <span>{message}</span>
    </div>
  );
}
