// SPDX-License-Identifier: Apache-2.0

import { useState } from 'react';
import { subscribeToPush, notificationPermission } from '@/lib/push';
import { useAppearance } from '@/stores/appearance';
import { useLabels } from '@/hooks/queries';
import { Button, Field } from '@/components/common/Form';
import { Icon } from '@/components/common/Icon';
import { InlineError } from '@/components/common/Feedback';
import { labelChipColors } from '@/lib/colors';

export function NotificationSettings() {
  const [permission, setPermission] = useState(notificationPermission());
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [done, setDone] = useState(false);

  const pushLabels = useAppearance((s) => s.pushLabels);
  const setPushLabels = useAppearance((s) => s.setPushLabels);
  const { data: labels } = useLabels();

  const toggleLabel = (name: string) => {
    setPushLabels(
      pushLabels.includes(name) ? pushLabels.filter((l) => l !== name) : [...pushLabels, name],
    );
  };

  const enable = async () => {
    setBusy(true);
    setError(null);
    try {
      const ok = await subscribeToPush();
      setPermission(notificationPermission());
      if (ok) setDone(true);
      else setError('通知が許可されませんでした。ブラウザの設定をご確認ください。');
    } catch (e) {
      setError((e as Error).message || '購読に失敗しました。');
    } finally {
      setBusy(false);
    }
  };

  if (permission === 'unsupported') {
    return (
      <InlineError message="このブラウザはプッシュ通知に対応していません。" />
    );
  }

  return (
    <div className="flex flex-col gap-4">
      <p className="text-[12.5px] leading-relaxed text-content-sub">
        重要なメールが届いたとき、AIの要約付きでプッシュ通知を受け取れます（重要度が高いメールのみ）。
        ブラウザを閉じていても通知されます。
      </p>

      <div className="flex items-center gap-3 rounded-xl border border-line p-4">
        <Icon
          name={permission === 'granted' ? 'notifications_active' : 'notifications'}
          size={24}
          className="text-content-sub"
        />
        <div className="flex-1">
          <div className="text-[14px] font-semibold text-content">プッシュ通知</div>
          <div className="text-[12px] text-content-sub">
            状態: {permission === 'granted' ? '許可済み' : permission === 'denied' ? 'ブロック中' : '未設定'}
          </div>
        </div>
        <Button variant="primary" onClick={enable} loading={busy} disabled={done && permission === 'granted'}>
          {done ? '登録しました' : '通知を有効にする'}
        </Button>
      </div>

      {error && <InlineError message={error} />}

      <Field
        label="通知するラベル"
        hint="選択したラベルが付いたメールのみ通知します。未選択の場合は重要度が高いメール全般を通知します。"
      >
        {(labels ?? []).length === 0 ? (
          <p className="text-[12.5px] text-content-sub">ラベルがまだありません。</p>
        ) : (
          <div className="flex flex-wrap gap-2">
            {(labels ?? []).map((l) => {
              const active = pushLabels.includes(l.name);
              const { bg, fg } = labelChipColors(l.color);
              return (
                <button
                  key={l.id ?? l.name}
                  type="button"
                  onClick={() => toggleLabel(l.name)}
                  className="flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-[12.5px] font-semibold transition-colors"
                  style={
                    active
                      ? { background: bg, color: fg, borderColor: 'transparent' }
                      : { borderColor: 'var(--line)', color: 'var(--text-sub)' }
                  }
                  aria-pressed={active}
                >
                  {active && <Icon name="check" size={14} />}
                  {l.name}
                </button>
              );
            })}
          </div>
        )}
      </Field>
    </div>
  );
}
