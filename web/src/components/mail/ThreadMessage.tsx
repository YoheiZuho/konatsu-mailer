// SPDX-License-Identifier: Apache-2.0

import { useMemo, useRef, useState } from 'react';
import type { ThreadMessage as ThreadMessageType } from '@/lib/types';
import { formatFullDate } from '@/lib/format';
import { sanitizeEmailHtml, loadRemoteImages, htmlToText } from '@/lib/sanitize';
import { useAppearance } from '@/stores/appearance';
import { useTranslate } from '@/hooks/translate';
import { ApiRequestError } from '@/lib/api';
import { Avatar } from '@/components/common/Avatar';
import { Icon } from '@/components/common/Icon';
import { Spinner } from '@/components/common/Feedback';

interface ThreadMessageProps {
  message: ThreadMessageType;
  /** Collapsed by default for older messages in a thread. */
  defaultCollapsed?: boolean;
  /** Show the "翻訳" action (translation configured server-side). */
  translationEnabled?: boolean;
  onReply?: () => void;
}

export function ThreadMessage({
  message,
  defaultCollapsed = false,
  translationEnabled = false,
  onReply,
}: ThreadMessageProps) {
  const [collapsed, setCollapsed] = useState(defaultCollapsed);
  const senderName = message.from.name || message.from.addr;

  const target = useAppearance((s) => s.translateTarget);
  const translate = useTranslate();
  const [translated, setTranslated] = useState<string | null>(null);
  const [detectedSource, setDetectedSource] = useState<string | undefined>();
  const [showOriginal, setShowOriginal] = useState(false);
  const [translateError, setTranslateError] = useState<string | null>(null);

  // Plain text used as the translation source (HTML is stripped to text).
  const sourceText = useMemo(() => {
    if (message.text) return message.text;
    if (message.html) return htmlToText(message.html);
    return '';
  }, [message.text, message.html]);

  const runTranslate = () => {
    if (translated) {
      // Toggle between translated / original views without re-requesting.
      setShowOriginal((v) => !v);
      return;
    }
    setTranslateError(null);
    translate.mutate(
      { q: sourceText, target },
      {
        onSuccess: (r) => {
          setTranslated(r.translated_text);
          setDetectedSource(r.detected_source);
          setShowOriginal(false);
        },
        onError: (e) =>
          setTranslateError(
            e instanceof ApiRequestError ? e.message : '翻訳に失敗しました。',
          ),
      },
    );
  };

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

  const canTranslate = translationEnabled && sourceText.trim().length > 0;
  const showingTranslation = !!translated && !showOriginal;

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
          {canTranslate && (
            <button
              className="icon-btn-sm"
              onClick={runTranslate}
              title={translated ? (showOriginal ? '翻訳を表示' : '原文を表示') : `翻訳（→ ${target}）`}
              aria-label="翻訳"
            >
              {translate.isPending ? <Spinner size={17} /> : <Icon name="translate" size={19} />}
            </button>
          )}
          {onReply && (
            <button className="icon-btn-sm" onClick={onReply} aria-label="この投稿に返信">
              <Icon name="reply" size={19} />
            </button>
          )}
        </div>
      </header>

      {translateError && (
        <div className="mb-3 flex items-center gap-2 rounded-lg px-3 py-2 text-[12.5px]" role="alert"
          style={{ background: 'color-mix(in srgb, var(--prio-high) 12%, var(--surface))', color: 'var(--prio-high)' }}>
          <Icon name="error" size={16} />
          {translateError}
        </div>
      )}

      {showingTranslation ? (
        <div>
          <div
            className="mb-3 flex items-center gap-2 rounded-lg px-3 py-2 text-[12px]"
            style={{ background: 'var(--brand-weak)', color: 'var(--text)' }}
          >
            <Icon name="translate" size={16} className="text-brand" />
            <span className="flex-1">
              翻訳済み{detectedSource ? `（${detectedSource} → ${target}）` : `（→ ${target}）`} ・ LibreTranslate
            </span>
            <button className="font-semibold hover:underline" onClick={() => setShowOriginal(true)}>
              原文を表示
            </button>
          </div>
          <div className="whitespace-pre-wrap text-[14.5px] leading-relaxed text-content">{translated}</div>
        </div>
      ) : (
        <MessageBody html={message.html} text={message.text} />
      )}
    </article>
  );
}

function MessageBody({ html, text }: { html?: string | null; text?: string | null }) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [imagesLoaded, setImagesLoaded] = useState(false);

  const safeHtml = useMemo(() => (html ? sanitizeEmailHtml(html) : null), [html]);
  const showImageBanner = !!safeHtml && /data-blocked-src/.test(safeHtml) && !imagesLoaded;

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
