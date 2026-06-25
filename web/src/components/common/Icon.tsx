// SPDX-License-Identifier: Apache-2.0

import clsx from 'clsx';

interface IconProps {
  /** Material Symbols Outlined ligature name, e.g. "inbox". */
  name: string;
  /** Font size in px. */
  size?: number;
  filled?: boolean;
  className?: string;
  style?: React.CSSProperties;
}

/** Renders a Material Symbols Outlined glyph. */
export function Icon({ name, size = 22, filled, className, style }: IconProps) {
  return (
    <span
      className={clsx('material-symbols-outlined', filled && 'filled', className)}
      style={{ fontSize: size, ...style }}
      aria-hidden="true"
    >
      {name}
    </span>
  );
}
