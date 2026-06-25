// SPDX-License-Identifier: Apache-2.0

import { useSyncExternalStore } from 'react';

/** Subscribe to a CSS media query, SSR-safe via useSyncExternalStore. */
export function useMediaQuery(query: string): boolean {
  return useSyncExternalStore(
    (callback) => {
      const mql = window.matchMedia(query);
      mql.addEventListener('change', callback);
      return () => mql.removeEventListener('change', callback);
    },
    () => window.matchMedia(query).matches,
    () => false,
  );
}
