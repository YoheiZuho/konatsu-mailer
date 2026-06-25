// SPDX-License-Identifier: Apache-2.0

import { useState } from 'react';
import clsx from 'clsx';
import { useUI } from '@/stores/ui';
import { Modal } from '@/components/common/Modal';
import { Icon } from '@/components/common/Icon';
import { AppearanceSettings } from '@/components/settings/AppearanceSettings';
import { AccountSettings } from '@/components/settings/AccountSettings';
import { LLMSettings } from '@/components/settings/LLMSettings';
import { LabelSettings } from '@/components/settings/LabelSettings';
import { NotificationSettings } from '@/components/settings/NotificationSettings';

type TabId = 'appearance' | 'accounts' | 'llm' | 'labels' | 'notifications';

const TABS: ReadonlyArray<{ id: TabId; icon: string; label: string }> = [
  { id: 'appearance', icon: 'palette', label: '外観' },
  { id: 'accounts', icon: 'alternate_email', label: 'アカウント' },
  { id: 'llm', icon: 'smart_toy', label: 'AI接続' },
  { id: 'labels', icon: 'label', label: 'ラベル' },
  { id: 'notifications', icon: 'notifications', label: '通知' },
];

export function SettingsDialog() {
  const close = useUI((s) => s.setSettingsOpen);
  const [tab, setTab] = useState<TabId>('appearance');

  return (
    <Modal open onClose={() => close(false)} widthClass="max-w-3xl" labelledBy="settings-title">
      <div className="flex h-[80vh] max-h-[640px] flex-col">
        <div className="flex h-14 flex-none items-center border-b border-line px-5">
          <h2 id="settings-title" className="flex-1 text-[16px] font-semibold text-content">
            設定
          </h2>
          <button className="icon-btn-sm" onClick={() => close(false)} aria-label="閉じる">
            <Icon name="close" size={20} />
          </button>
        </div>

        <div className="flex min-h-0 flex-1">
          {/* Tab rail */}
          <nav className="flex w-14 flex-none flex-col gap-1 border-r border-line p-2 sm:w-44">
            {TABS.map((t) => {
              const active = tab === t.id;
              return (
                <button
                  key={t.id}
                  onClick={() => setTab(t.id)}
                  className={clsx(
                    'flex items-center gap-3 rounded-lg px-2.5 py-2.5 text-left text-[13.5px] font-medium transition-colors sm:px-3',
                    active ? '' : 'text-content-sub hover:bg-hover',
                  )}
                  style={active ? { background: 'var(--brand-weak)', color: 'var(--text)' } : undefined}
                  aria-current={active ? 'page' : undefined}
                  title={t.label}
                >
                  <Icon name={t.icon} size={20} className={active ? 'text-brand' : undefined} />
                  <span className="hidden sm:inline">{t.label}</span>
                </button>
              );
            })}
          </nav>

          {/* Panel */}
          <div className="min-w-0 flex-1 overflow-y-auto p-5 sm:p-6">
            {tab === 'appearance' && <AppearanceSettings />}
            {tab === 'accounts' && <AccountSettings />}
            {tab === 'llm' && <LLMSettings />}
            {tab === 'labels' && <LabelSettings />}
            {tab === 'notifications' && <NotificationSettings />}
          </div>
        </div>
      </div>
    </Modal>
  );
}
