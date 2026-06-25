// SPDX-License-Identifier: Apache-2.0

import clsx from 'clsx';
import { useThread } from '@/hooks/queries';
import { useToggleStar, useMarkRead, useReanalyze } from '@/hooks/mutations';
import { useTranslateConfig } from '@/hooks/translate';
import { useAppearance } from '@/stores/appearance';
import { Icon } from '@/components/common/Icon';
import { LabelChip } from '@/components/common/LabelChip';
import { CenteredSpinner, EmptyState, InlineError } from '@/components/common/Feedback';
import { AiSummaryCard } from '@/components/mail/AiSummaryCard';
import { ThreadMessage } from '@/components/mail/ThreadMessage';
import { QuickReply } from '@/components/mail/QuickReply';

interface ReadingPaneProps {
  emailId: string | null;
  onBack: () => void;
  showBack: boolean;
  className?: string;
}

export function ReadingPane({ emailId, onBack, showBack, className }: ReadingPaneProps) {
  const aiOn = useAppearance((s) => s.aiSummaries);
  const thread = useThread(emailId);
  const translateConfig = useTranslateConfig();
  const translationEnabled = translateConfig.data?.enabled ?? false;
  const toggleStar = useToggleStar();
  const markRead = useMarkRead();
  const reanalyze = useReanalyze();

  if (!emailId) {
    return (
      <section className={clsx('flex bg-surface', className)}>
        <EmptyState
          icon="drafts"
          title="メールを選択してください"
          description="左の一覧からメールを選ぶと、ここに本文とAI要約が表示されます。"
        />
      </section>
    );
  }

  const data = thread.data;
  const messages = data?.messages ?? [];
  const last = messages.at(-1);
  // The AI summary/priority is most useful from the latest analyzed message.
  const aiSummary = [...messages].reverse().find((m) => m.ai_summary)?.ai_summary ?? null;
  const aiPriority = [...messages].reverse().find((m) => m.ai_priority != null)?.ai_priority ?? null;

  return (
    <section className={clsx('flex flex-col bg-surface', className)}>
      {/* Toolbar */}
      <div className="flex h-12 flex-none items-center gap-1 border-b border-line px-3">
        {showBack && (
          <button className="icon-btn-sm" onClick={onBack} aria-label="一覧へ戻る">
            <Icon name="arrow_back" size={21} />
          </button>
        )}
        <button
          className="icon-btn-sm"
          onClick={() => emailId && toggleStar.mutate({ id: emailId, is_starred: !data?.is_starred })}
          aria-label="スター"
        >
          <Icon
            name="star"
            size={20}
            filled={data?.is_starred}
            style={data?.is_starred ? { color: 'var(--brand)' } : undefined}
          />
        </button>
        <button
          className="icon-btn-sm"
          onClick={() => emailId && markRead.mutate({ id: emailId, is_read: false })}
          aria-label="未読にする"
        >
          <Icon name="mark_email_unread" size={20} />
        </button>
        <button
          className="icon-btn-sm"
          onClick={() => emailId && reanalyze.mutate(emailId)}
          aria-label="AIで再解析"
          title="AIで再解析"
        >
          <Icon name="auto_awesome" size={20} />
        </button>
        <div className="flex-1" />
      </div>

      {/* Body */}
      <div className="min-h-0 flex-1 overflow-y-auto px-5 py-6 sm:px-8">
        {thread.isLoading ? (
          <CenteredSpinner />
        ) : thread.isError ? (
          <InlineError message={thread.error?.message ?? 'メールの取得に失敗しました'} />
        ) : (
          <div className="mx-auto max-w-3xl">
            <div className="mb-3.5 flex items-start gap-3">
              <h1 className="flex-1 text-[22px] font-semibold leading-snug text-content">
                {data?.subject || messages[0]?.subject || '(件名なし)'}
              </h1>
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

      {/* Reply */}
      {data && last && <QuickReply threadId={data.thread_id} subject={data.subject ?? ''} replyTo={last} />}
    </section>
  );
}
