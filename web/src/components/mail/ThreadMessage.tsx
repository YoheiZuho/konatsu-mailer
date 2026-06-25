// SPDX-License-Identifier: Apache-2.0

import { useMemo, useRef, useState } from 'react';
import type { ThreadMessage as ThreadMessageType } from '@/lib/types';
import { formatFullDate } from '@/lib/format';
import { sanitizeEmailHtml, loadRemoteImages } from '@/lib/sanitize';
import { Avatar } from '@/components/common/Avatar';
import { Icon } from '@/components/common/Icon';

interface ThreadMessageProps {
  message: ThreadMessageType;
  /** Collapsed by default for older messages in a thread. */
  defaultCollapsed?: boolean;
  onReply?: () => void;
}

export function ThreadMessage({ message, defaultCollapsed = false, onReply }: ThreadMessageProps) {
  const [collapsed, setCollapsed] = useState(defaultCollapsed);
  const senderName = message.from.name || message.from.addr;

  if (collapsed) {
    return (
      <button
        onClick={() => setCollapsed(false)}
        className="mb-3 flex w-full items-center gap-3 rounded-lg border border-line px-4 py-3 text-left transition hover:bg-hover"
      >
        <Avatar name={senderName} seed={message.from.addr} size={32} />
        <div className="min-w-0 flex-1">
          <span className="text-[13.5px] font-semibold text-content">{senderName}</span>
          {message.text && (
            <span className="ml-2 truncate text-[13px] text-content-sub">
              {message.text.slice(0, 80)}
            </span>
          )}
        </div>
        <span className="flex-none text-[12px] text-content-sub">{formatFullDate(message.date)}</span>
        <Icon name="expand_more" size={20} className="flex-none text-content-sub" />
      </button>
    );
  }

  return (
    <article className="mb-5 border-b border-line pb-5 last:border-b-0">
      <header className="mb-4 flex items-start gap-3.5">
        <Avatar name={senderName} seed={message.from.addr} size={44} />
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <span className="text-[15px] font-semibold text-content">{senderName}</span>
            <span className="text-[13px] text-content-sub">&lt;{message.from.addr}&gt;</span>
          </div>
          <div className="mt-0.5 truncate text-[12.5px] text-content-sub">
            宛先: {message.to.map((a) => a.name || a.addr).join(', ') || '—'}
          </div>
        </div>
        <div className="flex flex-none items-center gap-1">
          <span className="text-[12.5px] text-content-sub">{formatFullDate(message.date)}</span>
          {onReply && (
            <button className="icon-btn-sm" onClick={onReply} aria-label="この投稿に返信">
              <Icon name="reply" size={19} />
            </button>
          )}
        </div>
      </header>

      <MessageBody html={message.html} text={message.text} />
    </article>
  );
}

function MessageBody({ html, text }: { html?: string | null; text?: string | null }) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [imagesLoaded, setImagesLoaded] = useState(false);

  const safeHtml = useMemo(() => (html ? sanitizeEmailHtml(html) : null), [html]);
  const showImageBanner =
    !!safeHtml && /data-blocked-src/.test(safeHtml) && !imagesLoaded;

  if (safeHtml) {
    return (
      <div>
        {showImageBanner && (
          <div className="mb-3 flex items-center gap-2 rounded-lg bg-surface-sub px-3 py-2 text-[12.5px] text-content-sub">
            <Icon name="hide_image" size={18} />
            <span className="flex-1">プライバシー保護のため画像をブロックしました。</span>
            <button
              className="font-semibold text-content hover:underline"
              onClick={() => {
                if (containerRef.current) {
                  loadRemoteImages(containerRef.current);
                  setImagesLoaded(true);
                }
              }}
            >
              画像を表示
            </button>
          </div>
        )}
        <div
          ref={containerRef}
          className="email-html max-w-none text-[14.5px] leading-relaxed text-content [&_a]:text-brand-strong [&_a]:underline"
          // Content is sanitized by DOMPurify in sanitizeEmailHtml().
          dangerouslySetInnerHTML={{ __html: safeHtml }}
        />
      </div>
    );
  }

  return (
    <div className="whitespace-pre-wrap text-[14.5px] leading-relaxed text-content">
      {text || '(本文がありません)'}
    </div>
  );
}
