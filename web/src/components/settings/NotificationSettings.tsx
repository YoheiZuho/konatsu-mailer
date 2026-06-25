// SPDX-License-Identifier: Apache-2.0

import { useState } from 'react';
import { subscribeToPush, notificationPermission } from '@/lib/push';
import { Button } from '@/components/common/Form';
import { Icon } from '@/components/common/Icon';
import { InlineError } from '@/components/common/Feedback';

export function NotificationSettings() {
  const [permission, setPermission] = useState(notificationPermission());
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [done, setDone] = useState(false);

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
    </div>
  );
}
