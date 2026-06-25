// SPDX-License-Identifier: Apache-2.0

import { useRef, useState } from 'react';
import type { ThreadMessage } from '@/lib/types';
import { streamDraft } from '@/lib/ai';
import { useSendEmail } from '@/hooks/mutations';
import { useUI } from '@/stores/ui';
import { Icon } from '@/components/common/Icon';
import { Spinner } from '@/components/common/Feedback';

interface QuickReplyProps {
  threadId: string;
  subject: string;
  /** The message being replied to (last in thread). */
  replyTo: ThreadMessage | undefined;
}

/** Inline reply bar with an "AI draft" action (design doc §5.7 / §9.3). */
export function QuickReply({ threadId, subject, replyTo }: QuickReplyProps) {
  const [body, setBody] = useState('');
  const [drafting, setDrafting] = useState(false);
  const abortRef = useRef<AbortController | null>(null);
  const send = useSendEmail();
  const openCompose = useUI((s) => s.openCompose);

  const recipient = replyTo?.from.name || replyTo?.from.addr || '';

  const generateDraft = async () => {
    if (drafting) {
      abortRef.current?.abort();
      setDrafting(false);
      return;
    }
    setDrafting(true);
    setBody('');
    const controller = new AbortController();
    abortRef.current = controller;
    await streamDraft(
      { mode: 'reply', thread_id: threadId, email_id: replyTo?.id, instruction: '' },
      {
        signal: controller.signal,
        onChunk: (text) => setBody((prev) => prev + text),
        onDone: () => setDrafting(false),
        onError: () => setDrafting(false),
      },
    );
  };

  const handleSend = () => {
    if (!body.trim() || !replyTo) return;
    send.mutate(
      {
        to: [replyTo.from.addr],
        subject: subject.startsWith('Re:') ? subject : `Re: ${subject}`,
        text: body,
        in_reply_to: replyTo.id,
        thread_id: threadId,
      },
      { onSuccess: () => setBody('') },
    );
  };

  const expand = () =>
    openCompose({
      mode: 'reply',
      to: replyTo?.from.addr ?? '',
      subject: subject.startsWith('Re:') ? subject : `Re: ${subject}`,
      body,
      inReplyTo: replyTo?.id,
      threadId,
    });

  return (
    <div className="flex-none border-t border-line px-4 py-3.5 sm:px-6">
      <div className="mx-auto flex max-w-3xl flex-col gap-2 rounded-2xl border border-line p-2 pl-4 shadow-sm focus-within:shadow-md">
        <div className="flex items-start gap-3 pt-1.5">
          <Icon name="reply" size={20} className="mt-1 flex-none text-content-sub" />
          <textarea
            value={body}
            onChange={(e) => setBody(e.target.value)}
            placeholder={recipient ? `${recipient} さんに返信…` : '返信を入力…'}
            rows={body ? 4 : 1}
            className="min-h-0 flex-1 resize-none bg-transparent py-1 text-[14px] text-content outline-none"
          />
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={generateDraft}
            className="flex h-8 items-center gap-1.5 rounded-full bg-surface-sub px-3 text-[12.5px] font-semibold text-content-sub transition hover:bg-hover"
          >
            {drafting ? <Spinner size={15} /> : <Icon name="auto_awesome" size={17} className="text-brand" />}
            {drafting ? '生成中…（クリックで停止）' : 'AIで返信案'}
          </button>
          <button
            onClick={expand}
            className="flex h-8 items-center gap-1.5 rounded-full px-2 text-[12.5px] font-medium text-content-sub hover:bg-hover"
            title="全画面で作成"
          >
            <Icon name="open_in_full" size={16} />
          </button>
          <div className="flex-1" />
          <button
            onClick={handleSend}
            disabled={!body.trim() || send.isPending}
            className="flex h-9 w-9 items-center justify-center rounded-full bg-brand transition hover:bg-brand-strong disabled:opacity-50"
            style={{ color: 'var(--on-brand)' }}
            aria-label="送信"
          >
            {send.isPending ? <Spinner size={17} /> : <Icon name="send" size={18} />}
          </button>
        </div>
      </div>
    </div>
  );
}
