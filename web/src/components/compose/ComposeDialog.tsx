// SPDX-License-Identifier: Apache-2.0

import { useRef, useState } from 'react';
import { useUI } from '@/stores/ui';
import { useSendEmail } from '@/hooks/mutations';
import { streamDraft } from '@/lib/ai';
import { parseAddressList } from '@/lib/email';
import { ApiRequestError } from '@/lib/api';
import { Icon } from '@/components/common/Icon';
import { Spinner, InlineError } from '@/components/common/Feedback';

/** Full compose window (design doc §9.3 ComposeDialog). */
export function ComposeDialog() {
  const compose = useUI((s) => s.compose);
  const update = useUI((s) => s.updateCompose);
  const close = useUI((s) => s.closeCompose);
  const send = useSendEmail();

  const [showCc, setShowCc] = useState(!!compose.cc || !!compose.bcc);
  const [drafting, setDrafting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  const generateDraft = async () => {
    if (drafting) {
      abortRef.current?.abort();
      setDrafting(false);
      return;
    }
    setError(null);
    setDrafting(true);
    update({ body: '' });
    const controller = new AbortController();
    abortRef.current = controller;
    await streamDraft(
      {
        mode: compose.mode === 'compose' ? 'compose' : 'reply',
        thread_id: compose.threadId,
        email_id: compose.inReplyTo,
        instruction: compose.subject,
        context: compose.body,
      },
      {
        signal: controller.signal,
        onChunk: (text) => update({ body: (useUI.getState().compose.body ?? '') + text }),
        onDone: () => setDrafting(false),
        onError: (e) => {
          setError(e.message);
          setDrafting(false);
        },
      },
    );
  };

  const handleSend = () => {
    const to = parseAddressList(compose.to);
    if (to.length === 0) {
      setError('宛先を入力してください。');
      return;
    }
    setError(null);
    send.mutate(
      {
        to,
        cc: parseAddressList(compose.cc),
        bcc: parseAddressList(compose.bcc),
        subject: compose.subject,
        text: compose.body,
        in_reply_to: compose.inReplyTo,
        thread_id: compose.threadId,
      },
      {
        onSuccess: () => close(),
        onError: (e) =>
          setError(e instanceof ApiRequestError ? e.message : '送信に失敗しました。'),
      },
    );
  };

  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center p-0 sm:items-end sm:justify-end sm:p-6">
      {/* Backdrop only on mobile (full screen); desktop is a docked window. */}
      <div className="absolute inset-0 bg-black/30 sm:hidden" onClick={close} />
      <div className="relative flex h-full w-full flex-col overflow-hidden bg-surface shadow-compose sm:h-[620px] sm:w-[560px] sm:rounded-xl">
        {/* Header */}
        <div className="flex h-12 flex-none items-center gap-2 border-b border-line bg-surface-sub px-4">
          <span className="flex-1 text-[14px] font-semibold text-content">
            {compose.mode === 'compose' ? '新規メッセージ' : '返信'}
          </span>
          <button className="icon-btn-sm" onClick={close} aria-label="閉じる">
            <Icon name="close" size={20} />
          </button>
        </div>

        {/* Recipients */}
        <div className="flex items-center gap-2 border-b border-line px-4" style={{ height: 46 }}>
          <span className="w-9 flex-none text-[13px] text-content-sub">宛先</span>
          <input
            value={compose.to}
            onChange={(e) => update({ to: e.target.value })}
            className="flex-1 bg-transparent text-[14px] text-content outline-none"
            placeholder="recipient@example.com"
            autoFocus
          />
          {!showCc && (
            <button
              className="text-[13px] text-content-sub hover:text-content"
              onClick={() => setShowCc(true)}
            >
              Cc / Bcc
            </button>
          )}
        </div>
        {showCc && (
          <>
            <Row label="Cc" value={compose.cc} onChange={(v) => update({ cc: v })} />
            <Row label="Bcc" value={compose.bcc} onChange={(v) => update({ bcc: v })} />
          </>
        )}

        {/* Subject */}
        <div className="flex items-center border-b border-line px-4" style={{ height: 46 }}>
          <input
            value={compose.subject}
            onChange={(e) => update({ subject: e.target.value })}
            className="flex-1 bg-transparent text-[14px] font-medium text-content outline-none"
            placeholder="件名"
          />
        </div>

        {/* Body */}
        <div className="relative min-h-0 flex-1">
          <button
            onClick={generateDraft}
            className="absolute right-4 top-3.5 z-10 flex h-8 items-center gap-1.5 rounded-2xl border px-3 text-[12.5px] font-semibold transition"
            style={{
              borderColor: 'color-mix(in srgb, var(--brand) 35%, transparent)',
              background: 'var(--brand-weak)',
              color: 'var(--text)',
            }}
          >
            {drafting ? <Spinner size={15} /> : <Icon name="auto_awesome" size={17} className="text-brand" />}
            {drafting ? '生成中…' : 'AIで下書きを生成'}
          </button>
          <textarea
            value={compose.body}
            onChange={(e) => update({ body: e.target.value })}
            className="h-full w-full resize-none bg-transparent px-4 pb-4 pt-14 text-[14.5px] leading-relaxed text-content outline-none"
            placeholder="本文を入力…"
          />
        </div>

        {error && (
          <div className="px-4 pb-2">
            <InlineError message={error} />
          </div>
        )}

        {/* Footer */}
        <div className="flex flex-none items-center gap-2 border-t border-line px-4 py-3">
          <button
            onClick={handleSend}
            disabled={send.isPending}
            className="flex h-10 items-center gap-2 rounded-full bg-brand pl-6 pr-3 text-[14.5px] font-semibold transition hover:bg-brand-strong disabled:opacity-60"
            style={{ color: 'var(--on-brand)' }}
          >
            {send.isPending ? <Spinner size={18} /> : null}
            送信
            <Icon name="send" size={19} />
          </button>
          <div className="flex-1" />
          <button className="icon-btn-sm" onClick={close} aria-label="破棄">
            <Icon name="delete" size={20} />
          </button>
        </div>
      </div>
    </div>
  );
}

function Row({ label, value, onChange }: { label: string; value: string; onChange: (v: string) => void }) {
  return (
    <div className="flex items-center gap-2 border-b border-line px-4" style={{ height: 44 }}>
      <span className="w-9 flex-none text-[13px] text-content-sub">{label}</span>
      <input
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="flex-1 bg-transparent text-[14px] text-content outline-none"
      />
    </div>
  );
}
