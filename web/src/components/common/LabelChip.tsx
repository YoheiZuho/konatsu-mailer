// SPDX-License-Identifier: Apache-2.0

import { labelChipColors } from '@/lib/colors';
import type { Label } from '@/lib/types';

interface LabelChipProps {
  label: Label;
  onRemove?: () => void;
}

/** Small colored pill for a mail label. */
export function LabelChip({ label, onRemove }: LabelChipProps) {
  const { bg, fg } = labelChipColors(label.color);
  return (
    <span
      className="inline-flex h-[18px] flex-none items-center gap-1 whitespace-nowrap rounded-[5px] px-2 text-[11px] font-semibold"
      style={{ background: bg, color: fg }}
    >
      {label.name}
      {onRemove && (
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation();
            onRemove();
          }}
          className="-mr-1 flex items-center opacity-70 hover:opacity-100"
          aria-label={`ラベル「${label.name}」を外す`}
        >
          <span className="material-symbols-outlined" style={{ fontSize: 13 }}>
            close
          </span>
        </button>
      )}
    </span>
  );
}
