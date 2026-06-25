// SPDX-License-Identifier: Apache-2.0

import clsx from 'clsx';
import { useUI, type Folder } from '@/stores/ui';
import { useLabels } from '@/hooks/queries';
import { Icon } from '@/components/common/Icon';

const FOLDERS: ReadonlyArray<{ key: Folder; icon: string; label: string }> = [
  { key: 'INBOX', icon: 'inbox', label: '受信トレイ' },
  { key: 'STARRED', icon: 'star', label: 'スター付き' },
  { key: 'IMPORTANT', icon: 'label_important', label: '重要' },
  { key: 'SENT', icon: 'send', label: '送信済み' },
  { key: 'DRAFTS', icon: 'draft', label: '下書き' },
  { key: 'SPAM', icon: 'report', label: '迷惑メール' },
  { key: 'TRASH', icon: 'delete', label: 'ゴミ箱' },
];

export function Sidebar() {
  const folder = useUI((s) => s.folder);
  const labelFilter = useUI((s) => s.labelFilter);
  const setFolder = useUI((s) => s.setFolder);
  const setLabelFilter = useUI((s) => s.setLabelFilter);
  const openCompose = useUI((s) => s.openCompose);
  const { data: labels } = useLabels();

  return (
    <nav className="flex w-[236px] flex-none flex-col overflow-y-auto border-r border-line bg-surface py-2 pb-4">
      <button
        onClick={() => openCompose()}
        className="mb-3.5 ml-3.5 mr-3 mt-1.5 flex h-12 items-center gap-3.5 self-start rounded-3xl border border-line bg-surface pl-4 pr-6 text-[14px] font-semibold text-content shadow-fab transition hover:bg-hover hover:shadow-md"
      >
        <Icon name="edit" size={22} className="text-brand" />
        作成
      </button>

      {FOLDERS.map((f) => (
        <SidebarItem
          key={f.key}
          icon={f.icon}
          label={f.label}
          active={folder === f.key && !labelFilter}
          onClick={() => setFolder(f.key)}
        />
      ))}

      <div className="mx-4 my-2.5 h-px bg-line" />
      <div className="pb-1 pl-[22px] pr-3 pt-3 text-[13px] font-semibold text-content-sub">ラベル</div>

      {(labels ?? []).map((label) => (
        <SidebarItem
          key={label.id ?? label.name}
          dotColor={label.color}
          label={label.name}
          active={labelFilter === label.name}
          onClick={() => setLabelFilter(label.name)}
        />
      ))}

      <div className="flex-1" />
      <div
        className="mx-4 mt-3.5 flex items-center gap-2 rounded-lg px-3 py-2 text-[11.5px] leading-snug text-content-sub"
        style={{ background: 'var(--brand-weak)' }}
      >
        <Icon name="auto_awesome" size={16} className="flex-none text-brand" />
        <span>ラベルはAIが自動付与します</span>
      </div>
    </nav>
  );
}

interface ItemProps {
  icon?: string;
  dotColor?: string;
  label: string;
  active: boolean;
  onClick: () => void;
}

function SidebarItem({ icon, dotColor, label, active, onClick }: ItemProps) {
  return (
    <button
      onClick={onClick}
      className={clsx(
        'mr-2.5 flex h-9 items-center gap-[18px] rounded-r-full pl-[22px] pr-3 text-left text-[14px] transition-colors',
        active ? 'font-bold' : 'text-content hover:bg-hover',
      )}
      style={active ? { background: 'var(--brand-weak)', color: 'var(--text)' } : undefined}
      aria-current={active ? 'page' : undefined}
    >
      <span className="flex w-5 flex-none justify-center">
        {icon ? (
          <Icon name={icon} size={20} className={active ? 'text-brand' : 'text-content-sub'} />
        ) : (
          <span
            className="h-[11px] w-[11px] rounded-[3px]"
            style={{ background: dotColor ?? 'var(--text-sub)' }}
          />
        )}
      </span>
      <span className="flex-1 truncate">{label}</span>
    </button>
  );
}
