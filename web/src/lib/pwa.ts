// SPDX-License-Identifier: Apache-2.0
//
// Service worker registration (vite-plugin-pwa, injectRegister: null).

import { registerSW } from 'virtual:pwa-register';

export function registerServiceWorker() {
  // Auto-update; the SW takes control on next navigation. Errors are non-fatal.
  try {
    registerSW({ immediate: true });
  } catch {
    /* SW unsupported or registration blocked — app still works online */
  }
}
