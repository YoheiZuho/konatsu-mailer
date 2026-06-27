// SPDX-License-Identifier: Apache-2.0

import { useEffect } from 'react';
import clsx from 'clsx';
import { useThread } from '@/hooks/queries';
import { useToggleStar, useMarkRead } from '@/hooks/mutations';
import { useTranslateConfig } from '@/hooks/translate';
import { useAppearance } from '@/stores/appearance';
import { useUI } from '@/stores/ui';
import { Icon } from '@/components/common/Icon';
import { LabelChip } from '@/components/common/LabelChip';
import { CenteredSpinner, EmptyState, InlineError } from '@/components/common/Feedback';
import { AiSummaryCard } from '@/components/mail/AiSummaryCard';
import { ThreadMessage } from '@/components/mail/ThreadMessage';
import { signatureHtml, quoteHtml } from '@/lib/compose';

interface ReadingPaneProps {
  emailId: string | null;
  onBack: () => void;
  showBack: boolean;
  className?: string;
}

export function ReadingPane({ emailId, onBack, showBack, className }: ReadingPaneProps) {
  const aiOn = useAppearance((s) => s.aiSummaries);
  const signature = useAppearance((s) => s.signature);
  const thread = useThread(emailId);
  const toggleStar = useToggleStar();
  const markRead = useMarkRead();
  const openCompose = useUI((s) => s.openCompose);
  const translationEnabled = useTranslateConfig().data?.enabled ?? false;

  // Escape closes the preview (returns to the full-width list).
  useEffect(() => {
    if (!emailId) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onBack();
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [emailId, onBack]);

  if (!emailId) {
    return (
      <section className={clsx('flex bg-bg', className)}>
        <EmptyState
          icon="drafts"
          title="メールを選択してください"
          description="一覧からメールを選ぶと、ここに本文とAI要約が表示されます。"
        />
      </section>
    );
  }

  const data = thread.data;
  const messages = data?.messages ?? [];
  const last = messages.at(-1);
  const aiSummary = [...messages].reverse().find((m) => m.ai_summary)?.ai_summary ?? null;
  const aiPriority = [...messages].reverse().find((m) => m.ai_priority != null)?.ai_priority ?? null;
  const subject = data?.subject || messages[0]?.subject || '(件名なし)';

  const startReply = () => {
    if (!last) return;
    openCompose({
      mode: 'reply',
      to: last.from.addr,
      subject: subject.startsWith('Re:') ? subject : `Re: ${subject}`,
      inReplyTo: last.id,
      threadId: data?.thread_id,
      body: signatureHtml(signature) + quoteHtml(last),
    });
  };

  const startForward = () => {
    openCompose({
      mode: 'forward',
      subject: subject.startsWith('Fwd:') ? subject : `Fwd: ${subject}`,
      body: signatureHtml(signature) + (last ? quoteHtml(last) : ''),
    });
  };

  return (
    <section className={clsx('flex flex-col bg-bg', className)}>
      {/* Toolbar (design 案B) */}
      <div className="flex h-12 flex-none items-center gap-1 border-b border-line px-3">
        <button className="icon-btn-sm" onClick={onBack} aria-label={showBack ? '一覧へ戻る' : 'プレビューを閉じる'} title="閉じる（一覧へ）">
          <Icon name={showBack ? 'arrow_back' : 'close'} size={21} />
        </button>
        <div className="mx-1 h-[22px] w-px bg-line" />
        <ToolbarButton icon="archive" label="アーカイブ" disabled />
        <ToolbarButton icon="report" label="迷惑メール報告" disabled />
        <ToolbarButton icon="delete" label="削除" disabled />
        <button
          className="icon-btn-sm"
          onClick={() => emailId && markRead.mutate({ id: emailId, is_read: false })}
          aria-label="未読にする"
          title="未読にする"
        >
          <Icon name="mark_email_unread" size={20} />
        </button>
        <ToolbarButton icon="schedule" label="スヌーズ" disabled />
        <div className="flex-1" />
      </div>

      {/* Body */}
      <div key={emailId} className="animate-fade-in min-h-0 flex-1 overflow-y-auto px-3 py-5 sm:px-7 sm:py-6">
        {thread.isLoading ? (
          <CenteredSpinner />
        ) : thread.isError ? (
          <InlineError message={thread.error?.message ?? 'メールの取得に失敗しました'} />
        ) : (
          <div className="mx-auto max-w-3xl">
            <div className="mb-[18px] flex items-start gap-3">
              <h2 className="flex-1 text-[21px] font-semibold leading-snug text-content">{subject}</h2>
              <button
                onClick={() => emailId && toggleStar.mutate({ id: emailId, is_starred: !data?.is_starred })}
                className="flex-none text-content-sub/50 hover:text-brand"
                aria-label="スター"
              >
                <Icon
                  name="star"
                  size={22}
                  filled={data?.is_starred}
                  style={data?.is_starred ? { color: 'var(--brand)' } : undefined}
                />
              </button>
            </div>

            {(data?.labels?.length ?? 0) > 0 && (
              <div className="mb-5 flex flex-wrap gap-2">
                {data!.labels!.map((l) => (
                  <LabelChip key={l.name} label={l} />
                ))}
              </div>
            )}

            {aiOn && aiSummary && <AiSummaryCard summary={aiSummary} priority={aiPriority} />}

            {messages.length === 0 ? (
              <EmptyState icon="mail" title="本文を表示できません" />
            ) : (
              messages.map((m, i) => (
                <ThreadMessage
                  key={m.id}
                  message={m}
                  defaultCollapsed={i < messages.length - 1}
                  translationEnabled={translationEnabled}
                />
              ))
            )}
          </div>
        )}
      </div>

      {/* Footer: 返信 / 転送 (design 案B) */}
      {data && last && (
        <div className="flex flex-none gap-2.5 border-t border-line px-6 py-3.5">
          <FooterButton icon="reply" label="返信" onClick={startReply} />
          <FooterButton icon="forward" label="転送" onClick={startForward} />
        </div>
      )}
    </section>
  );
}

function ToolbarButton({ icon, label, disabled }: { icon: string; label: string; disabled?: boolean }) {
  return (
    <button
      className={clsx('icon-btn-sm', disabled && 'cursor-not-allowed opacity-40')}
      aria-label={label}
      title={disabled ? `${label}（未対応）` : label}
      disabled={disabled}
    >
      <Icon name={icon} size={20} />
    </button>
  );
}

function FooterButton({ icon, label, onClick }: { icon: string; label: string; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className="flex h-10 items-center gap-2 rounded-[20px] border border-line bg-surface px-5 text-[14px] font-semibold text-content transition-colors hover:bg-hover"
    >
      <Icon name={icon} size={19} />
      {label}
    </button>
  );
}
