// SPDX-License-Identifier: Apache-2.0
//
// Avatar and label color helpers, ported from the design mock's `renderVals`.

/** Deterministic avatar palette (background / foreground pairs). */
const AVATAR_PALETTE: ReadonlyArray<{ bg: string; fg: string }> = [
  { bg: 'oklch(0.93 0.05 255)', fg: 'oklch(0.42 0.14 255)' },
  { bg: 'oklch(0.93 0.05 165)', fg: 'oklch(0.40 0.12 165)' },
  { bg: 'oklch(0.94 0.06 345)', fg: 'oklch(0.45 0.13 345)' },
  { bg: 'oklch(0.95 0.05 80)', fg: 'oklch(0.46 0.10 80)' },
  { bg: 'oklch(0.93 0.05 30)', fg: 'oklch(0.47 0.15 30)' },
  { bg: 'oklch(0.93 0.05 305)', fg: 'oklch(0.44 0.14 305)' },
  { bg: 'oklch(0.93 0.045 200)', fg: 'oklch(0.41 0.11 200)' },
  { bg: 'oklch(0.93 0.05 145)', fg: 'oklch(0.40 0.12 145)' },
];

/** A small stable hash so the same sender keeps the same avatar color. */
function hashString(input: string): number {
  let h = 0;
  for (let i = 0; i < input.length; i++) {
    h = (h << 5) - h + input.charCodeAt(i);
    h |= 0;
  }
  return Math.abs(h);
}

export function avatarColors(seed: string): { bg: string; fg: string } {
  return AVATAR_PALETTE[hashString(seed) % AVATAR_PALETTE.length];
}

/** First grapheme of a display name, used as the avatar initial. */
export function initialOf(name: string | undefined, fallback = '?'): string {
  const trimmed = (name ?? '').trim();
  if (!trimmed) return fallback;
  return Array.from(trimmed)[0]!.toUpperCase();
}

/**
 * Background/foreground for a label chip. Labels carry a `color` from the API
 * (an oklch foreground value); we derive a soft tinted background from it.
 * Falls back gracefully for non-oklch values.
 */
export function labelChipColors(color: string): { bg: string; fg: string } {
  const match = color.match(
    /oklch\(\s*([\d.]+)\s+([\d.]+)\s+([\d.]+)/i,
  );
  if (match) {
    const [, , c, h] = match;
    return {
      fg: color,
      bg: `oklch(0.965 ${Math.min(parseFloat(c), 0.035).toFixed(3)} ${h})`,
    };
  }
  // Non-oklch (e.g. a HEX): tint via color-mix with the surface.
  return {
    fg: color,
    bg: `color-mix(in srgb, ${color} 14%, var(--surface))`,
  };
}
