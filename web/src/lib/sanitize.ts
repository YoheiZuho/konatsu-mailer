// SPDX-License-Identifier: Apache-2.0
//
// HTML email sanitization (design doc §11): DOMPurify + neutralize external
// images by default (privacy / tracking-pixel protection).

import DOMPurify from 'dompurify';

let hookInstalled = false;

function installHook() {
  if (hookInstalled) return;
  hookInstalled = true;
  // Defer remote images: stash the URL and blank the src so nothing loads until
  // the user opts in. Links always open in a new tab with safe rel.
  DOMPurify.addHook('afterSanitizeAttributes', (node) => {
    if (node.tagName === 'IMG' && node.getAttribute('src')) {
      const src = node.getAttribute('src')!;
      if (/^https?:/i.test(src)) {
        node.setAttribute('data-blocked-src', src);
        node.removeAttribute('src');
      }
    }
    if (node.tagName === 'A') {
      node.setAttribute('target', '_blank');
      node.setAttribute('rel', 'noopener noreferrer nofollow');
    }
  });
}

export function sanitizeEmailHtml(html: string): string {
  installHook();
  return DOMPurify.sanitize(html, {
    USE_PROFILES: { html: true },
    FORBID_TAGS: ['style', 'script', 'iframe', 'object', 'embed', 'form'],
    FORBID_ATTR: ['srcset'],
  });
}

/** Restore deferred remote images within a rendered email container. */
export function loadRemoteImages(container: HTMLElement): void {
  container.querySelectorAll<HTMLImageElement>('img[data-blocked-src]').forEach((img) => {
    img.src = img.getAttribute('data-blocked-src')!;
    img.removeAttribute('data-blocked-src');
  });
}

export function hasBlockedImages(container: HTMLElement | null): boolean {
  return !!container?.querySelector('img[data-blocked-src]');
}

/** Extract readable plain text from an HTML string (for translation, etc.). */
export function htmlToText(html: string): string {
  const doc = new DOMParser().parseFromString(html, 'text/html');
  return (doc.body.textContent ?? '').replace(/\n{3,}/g, '\n\n').trim();
}
