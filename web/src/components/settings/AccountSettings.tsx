// SPDX-License-Identifier: Apache-2.0

import { useState } from 'react';
import { useAccounts } from '@/hooks/queries';
import { useCreateAccount, useDeleteAccount } from '@/hooks/mutations';
import { Button, Field, TextInput, Toggle } from '@/components/common/Form';
import { CenteredSpinner, EmptyState, InlineError } from '@/components/common/Feedback';
import { Icon } from '@/components/common/Icon';
import { ApiRequestError } from '@/lib/api';
import type { AccountInput } from '@/lib/types';

const blankAccount: AccountInput = {
  email: '',
  imap_host: '',
  imap_port: 993,
  imap_use_tls: true,
  smtp_host: '',
  smtp_port: 587,
  smtp_use_starttls: true,
  auth_user: '',
  password: '',
  is_active: true,
};

export function AccountSettings() {
  const { data: accounts, isLoading, isError, error } = useAccounts();
  const createAccount = useCreateAccount();
  const deleteAccount = useDeleteAccount();
  const [adding, setAdding] = useState(false);
  const [form, setForm] = useState<AccountInput>(blankAccount);
  const [formError, setFormError] = useState<string | null>(null);

  const set = <K extends keyof AccountInput>(key: K, value: AccountInput[K]) =>
    setForm((f) => ({ ...f, [key]: value }));

  const submit = () => {
    setFormError(null);
    createAccount.mutate(
      { ...form, auth_user: form.auth_user || form.email },
      {
        onSuccess: () => {
          setForm(blankAccount);
          setAdding(false);
        },
        onError: (e) =>
          setFormError(e instanceof ApiRequestError ? e.message : 'アカウントの追加に失敗しました。'),
      },
    );
  };

  if (isLoading) return <CenteredSpinner />;

  return (
    <div className="flex flex-col gap-4">
      {isError && <InlineError message={error?.message ?? 'アカウントの取得に失敗しました'} />}

      {(accounts ?? []).length === 0 && !adding && (
        <EmptyState
          icon="alternate_email"
          title="メールアカウントがありません"
          description="IMAP/SMTP アカウントを追加すると同期が始まります。"
        />
      )}

      {(accounts ?? []).map((a) => (
        <div
          key={a.id}
          className="flex items-center gap-3 rounded-xl border border-line p-4"
        >
          <Icon name="alternate_email" size={22} className="text-content-sub" />
          <div className="min-w-0 flex-1">
            <div className="truncate text-[14px] font-semibold text-content">{a.email}</div>
            <div className="truncate text-[12px] text-content-sub">
              {a.imap_host}:{a.imap_port} · {a.smtp_host}:{a.smtp_port}
            </div>
          </div>
          <span
            className="rounded-full px-2.5 py-1 text-[11px] font-semibold"
            style={
              a.is_active
                ? { background: 'var(--brand-weak)', color: 'var(--text)' }
                : { background: 'var(--surface-sub)', color: 'var(--text-sub)' }
            }
          >
            {a.is_active ? '有効' : '無効'}
          </span>
          <button
            className="icon-btn-sm"
            onClick={() => deleteAccount.mutate(a.id)}
            aria-label="アカウントを削除"
          >
            <Icon name="delete" size={19} />
          </button>
        </div>
      ))}

      {adding ? (
        <div className="flex flex-col gap-4 rounded-xl border border-line p-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field label="メールアドレス">
              <TextInput
                type="email"
                value={form.email}
                onChange={(e) => set('email', e.target.value)}
                placeholder="you@example.com"
              />
            </Field>
            <Field label="ログインID" hint="メールアドレスと同じ場合は空欄で可">
              <TextInput
                value={form.auth_user}
                onChange={(e) => set('auth_user', e.target.value)}
                placeholder="（任意）"
              />
            </Field>
            <Field label="IMAP ホスト">
              <TextInput
                value={form.imap_host}
                onChange={(e) => set('imap_host', e.target.value)}
                placeholder="imap.example.com"
              />
            </Field>
            <Field label="IMAP ポート">
              <TextInput
                type="number"
                value={form.imap_port}
                onChange={(e) => set('imap_port', Number(e.target.value))}
              />
            </Field>
            <Field label="SMTP ホスト">
              <TextInput
                value={form.smtp_host}
                onChange={(e) => set('smtp_host', e.target.value)}
                placeholder="smtp.example.com"
              />
            </Field>
            <Field label="SMTP ポート">
              <TextInput
                type="number"
                value={form.smtp_port}
                onChange={(e) => set('smtp_port', Number(e.target.value))}
              />
            </Field>
            <Field label="パスワード">
              <TextInput
                type="password"
                value={form.password ?? ''}
                onChange={(e) => set('password', e.target.value)}
                autoComplete="off"
              />
            </Field>
          </div>
          <div className="flex flex-wrap gap-5">
            <label className="flex items-center gap-2.5 text-[13px] text-content">
              <Toggle checked={form.imap_use_tls} onChange={(v) => set('imap_use_tls', v)} />
              IMAP TLS
            </label>
            <label className="flex items-center gap-2.5 text-[13px] text-content">
              <Toggle
                checked={form.smtp_use_starttls}
                onChange={(v) => set('smtp_use_starttls', v)}
              />
              SMTP STARTTLS
            </label>
          </div>

          {formError && <InlineError message={formError} />}

          <div className="flex gap-2">
            <Button variant="primary" onClick={submit} loading={createAccount.isPending}>
              追加
            </Button>
            <Button variant="ghost" onClick={() => setAdding(false)}>
              キャンセル
            </Button>
          </div>
        </div>
      ) : (
        <Button variant="secondary" className="self-start" onClick={() => setAdding(true)}>
          <Icon name="add" size={18} />
          アカウントを追加
        </Button>
      )}
    </div>
  );
}
