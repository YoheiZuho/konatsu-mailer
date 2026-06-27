// SPDX-License-Identifier: Apache-2.0
//
// Web Push subscription flow (design doc §10.3): fetch the VAPID public key,
// subscribe via the service worker, and register the subscription server-side.

import { api } from '@/lib/api';

/** Whether the browser supports the APIs required for Web Push. */
export function pushSupported(): boolean {
  return 'serviceWorker' in navigator && 'PushManager' in window && 'Notification' in window;
}

function urlBase64ToUint8Array(base64: string): Uint8Array<ArrayBuffer> {
  const padding = '='.repeat((4 - (base64.length % 4)) % 4);
  const normalized = (base64 + padding).replace(/-/g, '+').replace(/_/g, '/');
  const raw = atob(normalized);
  const out = new Uint8Array(new ArrayBuffer(raw.length));
  for (let i = 0; i < raw.length; i++) out[i] = raw.charCodeAt(i);
  return out;
}

/**
 * Request notification permission and subscribe to push. Throws an Error with a
 * specific, user-facing message on each distinct failure (so the UI can explain
 * what went wrong — e.g. VAPID not configured vs. permission denied).
 */
export async function subscribeToPush(): Promise<void> {
  if (!pushSupported()) {
    throw new Error('このブラウザはプッシュ通知に対応していません。');
  }

  const permission = await Notification.requestPermission();
  if (permission !== 'granted') {
    throw new Error('通知が許可されませんでした。ブラウザの設定で通知を許可してください。');
  }

  const { public_key } = await api.get<{ public_key: string }>('/push/vapid-public-key');
  if (!public_key) {
    throw new Error(
      'サーバーでプッシュ通知（VAPID鍵）が未設定です。VAPID_PUBLIC_KEY / VAPID_PRIVATE_KEY を設定してください。',
    );
  }

  const registration = await navigator.serviceWorker.ready;
  const existing = await registration.pushManager.getSubscription();
  const subscription =
    existing ??
    (await registration.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: urlBase64ToUint8Array(public_key),
    }));

  const json = subscription.toJSON();
  await api.post('/push/subscribe', {
    endpoint: subscription.endpoint,
    p256dh: json.keys?.p256dh,
    auth: json.keys?.auth,
    user_agent: navigator.userAgent,
  });
}

export function notificationPermission(): NotificationPermission | 'unsupported' {
  if (!pushSupported()) return 'unsupported';
  return Notification.permission;
}
