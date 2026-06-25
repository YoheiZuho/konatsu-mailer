// SPDX-License-Identifier: Apache-2.0

import { avatarColors, initialOf } from '@/lib/colors';

interface AvatarProps {
  name?: string | null;
  /** Used to seed the color; defaults to `name`. */
  seed?: string;
  size?: number;
}

/** Circular initial avatar with a deterministic color. */
export function Avatar({ name, seed, size = 38 }: AvatarProps) {
  const display = name ?? '';
  const { bg, fg } = avatarColors(seed ?? display);
  return (
    <div
      className="flex flex-none items-center justify-center rounded-full font-semibold"
      style={{
        width: size,
        height: size,
        background: bg,
        color: fg,
        fontSize: Math.round(size * 0.37),
      }}
      aria-hidden="true"
    >
      {initialOf(display)}
    </div>
  );
}
