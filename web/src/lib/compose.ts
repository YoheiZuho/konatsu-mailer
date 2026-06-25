// SPDX-License-Identifier: Apache-2.0
//
// Helpers for building compose/reply HTML bodies (signature, quoting).

import { htmlToText } from '@/lib/sanitize';
import { formatFullDate } from '@/lib/format';
import type { ThreadMessage } from '@/lib/types';

export function escapeHtml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

/** Convert plain text (e.g. an AI draft or signature) to simple HTML. */
export function textToHtml(text: string): string {
  return escapeHtml(text).replace(/\r?\n/g, '<br>');
}

/** HTML for the user's signature block, or "" when no signature is set. */
export function signatureHtml(signature: string): string {
  const trimmed = signature.trim();
  if (!trimmed) return '';
  return `<br><br><div data-signature="1">${textToHtml(trimmed)}</div>`;
}

/** A blockquote of the original message, like conventional reply quoting. */
export function quoteHtml(message: ThreadMessage): string {
  const sender = message.from.name || message.from.addr;
  const original = (message.text || (message.html ? htmlToText(message.html) : '')).trim();
  const quoted = textToHtml(original);
  return (
    `<br><br>` +
    `<div class="gmail_attr" style="color:#5f6368;font-size:13px">` +
    `${escapeHtml(formatFullDate(message.date))} ${escapeHtml(sender)} のメッセージ:</div>` +
    `<blockquote style="margin:0 0 0 8px;padding-left:12px;border-left:2px solid #c8ccd1;color:#5f6368">` +
    `${quoted}</blockquote>`
  );
}
