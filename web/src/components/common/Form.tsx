// SPDX-License-Identifier: Apache-2.0
//
// Lightweight, theme-aware form primitives shared by the settings panels.

import clsx from 'clsx';
import { Spinner } from '@/components/common/Feedback';

export function Field({
  label,
  hint,
  children,
}: {
  label: string;
  hint?: string;
  children: React.ReactNode;
}) {
  return (
    <label className="flex flex-col gap-1.5">
      <span className="text-[12.5px] font-medium text-content-sub">{label}</span>
      {children}
      {hint && <span className="text-[11.5px] text-content-sub/80">{hint}</span>}
    </label>
  );
}

const inputClass =
  'h-10 rounded-lg border border-line bg-surface px-3 text-[14px] text-content outline-none transition-colors focus:border-brand focus:ring-2 focus:ring-brand/30';

export function TextInput(props: React.InputHTMLAttributes<HTMLInputElement>) {
  return <input {...props} className={clsx(inputClass, props.className)} />;
}

export function Select(props: React.SelectHTMLAttributes<HTMLSelectElement>) {
  return <select {...props} className={clsx(inputClass, 'pr-8', props.className)} />;
}

export function Toggle({
  checked,
  onChange,
  label,
}: {
  checked: boolean;
  onChange: (v: boolean) => void;
  label?: string;
}) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      aria-label={label}
      onClick={() => onChange(!checked)}
      className={clsx(
        'relative inline-flex h-6 w-11 flex-none items-center rounded-full transition-colors',
        checked ? '' : 'bg-line',
      )}
      style={checked ? { background: 'var(--brand)' } : undefined}
    >
      <span
        className={clsx(
          'inline-block h-5 w-5 transform rounded-full bg-white shadow transition-transform',
          checked ? 'translate-x-[22px]' : 'translate-x-0.5',
        )}
      />
    </button>
  );
}

type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger';

export function Button({
  variant = 'secondary',
  loading,
  children,
  className,
  ...rest
}: React.ButtonHTMLAttributes<HTMLButtonElement> & { variant?: ButtonVariant; loading?: boolean }) {
  const base =
    'inline-flex h-10 items-center justify-center gap-2 rounded-full px-5 text-[14px] font-semibold transition-colors disabled:opacity-60';
  const variants: Record<ButtonVariant, string> = {
    primary: '',
    secondary: 'border border-line bg-surface text-content hover:bg-hover',
    ghost: 'text-content-sub hover:bg-hover',
    danger: 'border border-line text-content hover:bg-hover',
  };
  return (
    <button
      {...rest}
      disabled={rest.disabled || loading}
      className={clsx(base, variants[variant], className)}
      style={
        variant === 'primary'
          ? { background: 'var(--brand)', color: 'var(--on-brand)' }
          : variant === 'danger'
            ? { color: 'var(--prio-high)' }
            : undefined
      }
    >
      {loading && <Spinner size={16} />}
      {children}
    </button>
  );
}

export function SegmentedControl<T extends string>({
  options,
  value,
  onChange,
}: {
  options: ReadonlyArray<{ value: T; label: string }>;
  value: T;
  onChange: (v: T) => void;
}) {
  return (
    <div className="inline-flex rounded-lg bg-surface-sub p-1">
      {options.map((o) => {
        const active = o.value === value;
        return (
          <button
            key={o.value}
            type="button"
            onClick={() => onChange(o.value)}
            className={clsx(
              'rounded-md px-4 py-1.5 text-[13px] font-semibold transition-colors',
              active ? 'bg-surface text-content shadow-sm' : 'text-content-sub hover:text-content',
            )}
          >
            {o.label}
          </button>
        );
      })}
    </div>
  );
}
