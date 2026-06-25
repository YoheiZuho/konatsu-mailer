// SPDX-License-Identifier: Apache-2.0
//
// Custom service worker (vite-plugin-pwa injectManifest). Handles:
//   - Workbox precaching of the app shell (CacheFirst for static assets)
//   - NetworkFirst for API GETs so the UI works briefly offline (design §10.2)
//   - Web Push `push` + `notificationclick` (design §10.2 / §10.3)

/// <reference lib="webworker" />

import { clientsClaim } from 'workbox-core';
import { precacheAndRoute, cleanupOutdatedCaches } from 'workbox-precaching';
import { registerRoute } from 'workbox-routing';
import { NetworkFirst } from 'workbox-strategies';

declare const self: ServiceWorkerGlobalScope & {
  __WB_MANIFEST: Array<{ url: string; revision: string | null }>;
};

self.skipWaiting();
clientsClaim();

// Precache the build output (app shell, JS/CSS/fonts/icons).
cleanupOutdatedCaches();
precacheAndRoute(self.__WB_MANIFEST);

// API responses: try network, fall back to the last cached copy when offline.
registerRoute(
  ({ url, request }) => request.method === 'GET' && url.pathname.startsWith('/api/'),
  new NetworkFirst({
    cacheName: 'api-cache',
    networkTimeoutSeconds: 5,
  }),
);

interface PushPayload {
  title?: string;
  body?: string;
  data?: { email_id?: string };
  actions?: Array<{ action: string; title: string }>;
}

self.addEventListener('push', (event: PushEvent) => {
  let payload: PushPayload = {};
  try {
    payload = event.data?.json() ?? {};
  } catch {
    payload = { body: event.data?.text() };
  }

  const title = payload.title || 'konatsu';
  // `actions` is valid for persistent notifications but missing from the base
  // lib.dom NotificationOptions type, hence the widened cast.
  const options: NotificationOptions = {
    body: payload.body ?? '',
    icon: '/icons/icon-192.png',
    badge: '/icons/icon-192.png',
    data: payload.data ?? {},
    actions: payload.actions ?? [
      { action: 'reply', title: '今すぐ返信' },
      { action: 'later', title: '後で読む' },
    ],
    tag: payload.data?.email_id,
  } as NotificationOptions & {
    actions?: Array<{ action: string; title: string }>;
  };
  event.waitUntil(self.registration.showNotification(title, options));
});

self.addEventListener('notificationclick', (event: NotificationEvent) => {
  event.notification.close();
  if (event.action === 'later') return;

  const emailId = (event.notification.data as { email_id?: string } | undefined)?.email_id;
  const target = emailId ? `/mail/${emailId}` : '/';

  event.waitUntil(
    self.clients.matchAll({ type: 'window', includeUncontrolled: true }).then((clientList) => {
      // Focus an existing tab and navigate it; otherwise open a new window.
      for (const client of clientList) {
        if ('focus' in client) {
          (client as WindowClient).navigate(target);
          return (client as WindowClient).focus();
        }
      }
      return self.clients.openWindow(target);
    }),
  );
});
