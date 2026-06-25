// SPDX-License-Identifier: Apache-2.0

import { useEffect, useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { api, ApiRequestError } from '@/lib/api';
import { useAuth } from '@/stores/auth';
import type { AuthConfig, AuthTokens, MailAccountSetup, RegisterInput } from '@/lib/types';
import { Icon } from '@/components/common/Icon';
import { Spinner } from '@/components/common/Feedback';
import { Toggle } from '@/components/common/Form';

type Mode = 'login' | 'register';

const blankAccount: MailAccountSetup = {
  email: '',
  imap_host: '',
  imap_port: 993,
  imap_use_tls: true,
  smtp_host: '',
  smtp_port: 587,
  smtp_use_starttls: true,
  auth_user: '',
  password: '',
};

export function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const setTokens = useAuth((s) => s.setTokens);
  const setEmailStore = useAuth((s) => s.setEmail);

  const [mode, setMode] = useState<Mode>('login');
  const [allowRegistration, setAllowRegistration] = useState(true);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [displayName, setDisplayName] = useState('');

  const [setupAccount, setSetupAccount] = useState(false);
  const [account, setAccount] = useState<MailAccountSetup>(blankAccount);

  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const redirectTo = (location.state as { from?: string } | null)?.from ?? '/';

  // Find out whether self-service registration is open (env-gated server side).
  useEffect(() => {
    api
      .get<AuthConfig>('/auth/config')
      .then((cfg) => setAllowRegistration(cfg.allow_registration))
      .catch(() => setAllowRegistration(true));
  }, []);

  const setAcc = <K extends keyof MailAccountSetup>(key: K, value: MailAccountSetup[K]) =>
    setAccount((a) => ({ ...a, [key]: value }));

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setBusy(true);
    try {
      if (mode === 'login') {
        const tokens = await api.post<AuthTokens>('/auth/login', { email, password });
        setTokens(tokens);
        setEmailStore(email);
      } else {
        const body: RegisterInput = { email, password, display_name: displayName };
        if (setupAccount && account.imap_host && account.smtp_host && account.password) {
          body.mail_account = {
            ...account,
            email: account.email || email,
            auth_user: account.auth_user || undefined,
          };
        }
        const tokens = await api.post<AuthTokens>('/auth/register', body);
        setTokens(tokens);
        setEmailStore(email);
      }
      navigate(redirectTo, { replace: true });
    } catch (err) {
      setError(
        err instanceof ApiRequestError
          ? err.message
          : 'ネットワークエラーが発生しました。時間をおいて再度お試しください。',
      );
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="flex min-h-full items-center justify-center bg-bg px-4 py-10">
      <div className="w-full max-w-sm">
        <div className="mb-8 flex items-center justify-center gap-3">
          <div
            className="flex h-11 w-11 items-center justify-center rounded-xl bg-brand"
            style={{ color: 'var(--on-brand)' }}
          >
            <Icon name="mail" size={24} />
          </div>
          <div className="text-2xl font-semibold text-content">konatsu</div>
        </div>

        <div className="rounded-2xl border border-line bg-surface p-7 shadow-fab">
          <h1 className="mb-1 text-lg font-semibold text-content">
            {mode === 'login' ? 'ログイン' : 'アカウント作成'}
          </h1>
          <p className="mb-6 text-[13px] text-content-sub">
            {mode === 'login'
              ? 'メールアドレスとパスワードを入力してください。'
              : '新しいアカウントを作成します。'}
          </p>

          <form onSubmit={submit} className="flex flex-col gap-4">
            {mode === 'register' && (
              <Field
                label="表示名"
                type="text"
                value={displayName}
                onChange={setDisplayName}
                autoComplete="name"
                placeholder="山田 太郎"
              />
            )}
            <Field
              label="メールアドレス"
              type="email"
              value={email}
              onChange={setEmail}
              required
              autoComplete="email"
              placeholder="you@example.com"
            />
            <Field
              label="パスワード"
              type="password"
              value={password}
              onChange={setPassword}
              required
              autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
              placeholder={mode === 'register' ? '8文字以上' : '••••••••'}
            />

            {mode === 'register' && (
              <div className="rounded-xl border border-line">
                <label className="flex cursor-pointer items-center gap-3 p-3.5">
                  <Toggle checked={setupAccount} onChange={setSetupAccount} />
                  <span className="flex-1">
                    <span className="block text-[13.5px] font-medium text-content">
                      メールアカウントを設定
                    </span>
                    <span className="block text-[12px] text-content-sub">
                      IMAP/SMTP を今すぐ登録（後で設定画面からも追加できます）
                    </span>
                  </span>
                </label>

                {setupAccount && (
                  <div className="flex flex-col gap-3 border-t border-line p-3.5">
                    <Field
                      label="メールアドレス（アカウント）"
                      type="email"
                      value={account.email}
                      onChange={(v) => setAcc('email', v)}
                      placeholder="空欄ならログイン用と同じ"
                    />
                    <div className="grid grid-cols-3 gap-2">
                      <div className="col-span-2">
                        <Field
                          label="IMAP ホスト"
                          type="text"
                          value={account.imap_host}
                          onChange={(v) => setAcc('imap_host', v)}
                          placeholder="imap.example.com"
                        />
                      </div>
                      <Field
                        label="ポート"
                        type="number"
                        value={String(account.imap_port)}
                        onChange={(v) => setAcc('imap_port', Number(v))}
                      />
                    </div>
                    <div className="grid grid-cols-3 gap-2">
                      <div className="col-span-2">
                        <Field
                          label="SMTP ホスト"
                          type="text"
                          value={account.smtp_host}
                          onChange={(v) => setAcc('smtp_host', v)}
                          placeholder="smtp.example.com"
                        />
                      </div>
                      <Field
                        label="ポート"
                        type="number"
                        value={String(account.smtp_port)}
                        onChange={(v) => setAcc('smtp_port', Number(v))}
                      />
                    </div>
                    <Field
                      label="ログインID"
                      type="text"
                      value={account.auth_user ?? ''}
                      onChange={(v) => setAcc('auth_user', v)}
                      placeholder="空欄ならメールアドレスと同じ"
                    />
                    <Field
                      label="メールパスワード"
                      type="password"
                      value={account.password}
                      onChange={(v) => setAcc('password', v)}
                      autoComplete="off"
                    />
                    <div className="flex flex-wrap gap-4 pt-1">
                      <label className="flex items-center gap-2 text-[12.5px] text-content">
                        <Toggle checked={account.imap_use_tls} onChange={(v) => setAcc('imap_use_tls', v)} />
                        IMAP TLS
                      </label>
                      <label className="flex items-center gap-2 text-[12.5px] text-content">
                        <Toggle
                          checked={account.smtp_use_starttls}
                          onChange={(v) => setAcc('smtp_use_starttls', v)}
                        />
                        SMTP STARTTLS
                      </label>
                    </div>
                  </div>
                )}
              </div>
            )}

            {error && (
              <div className="text-[13px]" style={{ color: 'var(--prio-high)' }} role="alert">
                {error}
              </div>
            )}

            <button
              type="submit"
              disabled={busy}
              className="mt-1 flex h-11 items-center justify-center gap-2 rounded-full bg-brand text-[15px] font-semibold transition-colors hover:bg-brand-strong disabled:opacity-60"
              style={{ color: 'var(--on-brand)' }}
            >
              {busy && <Spinner size={18} />}
              {mode === 'login' ? 'ログイン' : '作成する'}
            </button>
          </form>

          {allowRegistration && (
            <div className="mt-6 text-center text-[13px] text-content-sub">
              {mode === 'login' ? 'アカウントをお持ちでないですか？' : 'すでにアカウントがありますか？'}{' '}
              <button
                type="button"
                className="font-semibold text-content underline-offset-2 hover:underline"
                onClick={() => {
                  setMode(mode === 'login' ? 'register' : 'login');
                  setError(null);
                }}
              >
                {mode === 'login' ? '新規登録' : 'ログイン'}
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

interface FieldProps {
  label: string;
  type: string;
  value: string;
  onChange: (v: string) => void;
  required?: boolean;
  autoComplete?: string;
  placeholder?: string;
}

function Field({ label, type, value, onChange, required, autoComplete, placeholder }: FieldProps) {
  return (
    <label className="flex flex-col gap-1.5">
      <span className="text-[12.5px] font-medium text-content-sub">{label}</span>
      <input
        type={type}
        value={value}
        required={required}
        autoComplete={autoComplete}
        placeholder={placeholder}
        onChange={(e) => onChange(e.target.value)}
        className="h-11 rounded-lg border border-line bg-surface px-3.5 text-[14px] text-content outline-none transition-colors focus:border-brand focus:ring-2 focus:ring-brand/30"
      />
    </label>
  );
}
