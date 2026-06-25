// SPDX-License-Identifier: Apache-2.0
//
// Locale-aware date/time formatting for mail lists and headers.

const time = new Intl.DateTimeFormat('ja-JP', { hour: '2-digit', minute: '2-digit' });
const monthDay = new Intl.DateTimeFormat('ja-JP', { month: 'long', day: 'numeric' });
const full = new Intl.DateTimeFormat('ja-JP', {
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
  hour: '2-digit',
  minute: '2-digit',
});

function startOfDay(d: Date): number {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime();
}

/** Compact list timestamp: time for today, "M月D日" otherwise. */
export function formatListDate(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '';
  const now = new Date();
  if (startOfDay(d) === startOfDay(now)) return time.format(d);
  if (d.getFullYear() === now.getFullYear()) return monthDay.format(d);
  return new Intl.DateTimeFormat('ja-JP', { year: 'numeric', month: 'numeric', day: 'numeric' }).format(d);
}

/** Full timestamp shown in a message header. */
export function formatFullDate(iso: string): string {
  const d = new Date(iso);
  return Number.isNaN(d.getTime()) ? '' : full.format(d);
}

/** Relative date-group label used by the list ("今日" / "今週" / "M月D日"). */
export function dateGroup(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '';
  const now = new Date();
  const day = 24 * 60 * 60 * 1000;
  const diff = startOfDay(now) - startOfDay(d);
  if (diff <= 0) return '今日';
  if (diff <= day) return '昨日';
  if (diff <= 7 * day) return '今週';
  if (diff <= 30 * day) return '今月';
  return monthDay.format(d);
}
