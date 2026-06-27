// SPDX-License-Identifier: Apache-2.0

import { useState } from 'react';
import clsx from 'clsx';
import { useUI } from '@/stores/ui';
import { useFolders, useLabels } from '@/hooks/queries';
import { useMoveEmail, useAssignLabels } from '@/hooks/mutations';
import { Icon } from '@/components/common/Icon';
import { EMAIL_DND_TYPE } from '@/components/mail/MailRow';
import type { FolderRole, MailFolder } from '@/lib/types';

const ROLE_META: Record<Exclude<FolderRole, ''>, { label: string; icon: string; order: number }> = {
  inbox: { label: '受信トレイ', icon: 'inbox', order: 0 },
  sent: { label: '送信済み', icon: 'send', order: 3 },
  drafts: { label: '下書き', icon: 'draft', order: 4 },
  junk: { label: '迷惑メール', icon: 'report', order: 5 },
  archive: { label: 'アーカイブ', icon: 'archive', order: 6 },
  trash: { label: 'ゴミ箱', icon: 'delete', order: 7 },
};

const folderLabel = (f: MailFolder) => (f.role && ROLE_META[f.role] ? ROLE_META[f.role].label : f.name);
const folderIcon = (f: MailFolder) => (f.role && ROLE_META[f.role] ? ROLE_META[f.role].icon : 'folder');
const folderOrder = (f: MailFolder) => (f.role && ROLE_META[f.role] ? ROLE_META[f.role].order : 8);

export function Sidebar({ width }: { width?: number }) {
  const folder = useUI((s) => s.folder);
  const view = useUI((s) => s.view);
  const labelFilter = useUI((s) => s.labelFilter);
  const setFolder = useUI((s) => s.setFolder);
  const setView = useUI((s) => s.setView);
  const setLabelFilter = useUI((s) => s.setLabelFilter);
  const openCompose = useUI((s) => s.openCompose);

  const { data } = useFolders();
  const { data: labels } = useLabels();
  const moveEmail = useMoveEmail();
  const assignLabels = useAssignLabels();

  const folders = data?.items ?? [];
  const inbox = folders.find((f) => f.role === 'inbox' || f.name === 'INBOX');
  const others = folders
    .filter((f) => f !== inbox)
    .sort((a, b) => folderOrder(a) - folderOrder(b) || a.name.localeCompare(b.name));

  const folderActive = (name: string) => view === 'folder' && folder === name && !labelFilter;

  return (
    <nav
      className="flex flex-none flex-col overflow-y-auto bg-surface py-2 pb-4"
      style={width ? { width } : undefined}
    >
      <button
        onClick={() => openCompose()}
        className="mb-3.5 ml-3.5 mr-3 mt-1.5 flex h-12 items-center gap-3.5 self-start rounded-3xl border border-line bg-surface pl-4 pr-6 text-[14px] font-semibold text-content shadow-fab transition-all duration-150 hover:bg-hover hover:shadow-md active:scale-[0.98]"
      >
        <Icon name="edit" size={22} className="text-brand" />
        作成
      </button>

      <Item
        icon="inbox"
        label="受信トレイ"
        count={inbox?.unread}
        active={folderActive(inbox?.name ?? 'INBOX')}
        onClick={() => setFolder(inbox?.name ?? 'INBOX')}
        onDropEmail={(id) => moveEmail.mutate({ id, folder: inbox?.name ?? 'INBOX' })}
      />
      <Item
        icon="star"
        label="スター付き"
        count={data?.starred_unread}
        active={view === 'starred'}
        onClick={() => setView('starred')}
      />
      <Item
        icon="label_important"
        label="重要"
        count={data?.important_unread}
        active={view === 'important'}
        onClick={() => setView('important')}
      />

      {others.map((f) => (
        <Item
          key={f.name}
          icon={folderIcon(f)}
          label={folderLabel(f)}
          count={f.unread}
          active={folderActive(f.name)}
          onClick={() => setFolder(f.name)}
          onDropEmail={(id) => moveEmail.mutate({ id, folder: f.name })}
        />
      ))}

      <div className="mx-4 my-2.5 h-px bg-line" />
      <div className="pb-1 pl-[22px] pr-3 pt-3 text-[13px] font-semibold text-content-sub">ラベル</div>

      {(labels ?? []).map((label) => (
        <Item
          key={label.id ?? label.name}
          dotColor={label.color}
          label={label.name}
          active={labelFilter === label.name}
          onClick={() => setLabelFilter(label.name)}
          onDropEmail={(id) => assignLabels.mutate({ id, add: [label.name] })}
        />
      ))}

      {/* The AI-label hint is only shown until the first label exists. */}
      {(labels ?? []).length === 0 && (
        <div
          className="mx-4 mt-3.5 flex items-center gap-2 rounded-lg px-3 py-2 text-[11.5px] leading-snug text-content-sub"
          style={{ background: 'var(--brand-weak)' }}
        >
          <Icon name="auto_awesome" size={16} className="flex-none text-brand" />
          <span>ラベルはAIが自動付与します</span>
        </div>
      )}
    </nav>
  );
}

interface ItemProps {
  icon?: string;
  dotColor?: string;
  label: string;
  count?: number;
  active: boolean;
  onClick: () => void;
  onDropEmail?: (emailID: string) => void;
}

function Item({ icon, dotColor, label, count, active, onClick, onDropEmail }: ItemProps) {
  const [dragOver, setDragOver] = useState(false);
  const dropHandlers = onDropEmail
    ? {
        onDragOver: (e: React.DragEvent) => {
          if (e.dataTransfer.types.includes(EMAIL_DND_TYPE)) {
            e.preventDefault();
            setDragOver(true);
          }
        },
        onDragLeave: () => setDragOver(false),
        onDrop: (e: React.DragEvent) => {
          e.preventDefault();
          setDragOver(false);
          const id = e.dataTransfer.getData(EMAIL_DND_TYPE);
          if (id) onDropEmail(id);
        },
      }
    : {};
  return (
    <button
      onClick={onClick}
      {...dropHandlers}
      className={clsx(
        'mr-2.5 flex h-9 items-center gap-[18px] rounded-r-full pl-[22px] pr-3 text-left text-[14px] transition-colors',
        active ? 'font-bold' : 'text-content hover:bg-hover',
        dragOver && 'ring-2 ring-inset ring-brand',
      )}
      style={active ? { background: 'var(--brand-weak)', color: 'var(--text)' } : undefined}
      aria-current={active ? 'page' : undefined}
    >
      <span className="flex w-5 flex-none justify-center">
        {icon ? (
          <Icon name={icon} size={20} className={active ? 'text-brand' : 'text-content-sub'} />
        ) : (
          <span className="h-[11px] w-[11px] rounded-[3px]" style={{ background: dotColor ?? 'var(--text-sub)' }} />
        )}
      </span>
      <span className="flex-1 truncate">{label}</span>
      {count != null && count > 0 && (
        <span className={clsx('flex-none text-[12px]', active ? 'font-bold' : 'font-semibold text-content-sub')}>
          {count > 999 ? '999+' : count}
        </span>
      )}
    </button>
  );
}
