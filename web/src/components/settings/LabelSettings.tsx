// SPDX-License-Identifier: Apache-2.0

import { useState } from 'react';
import { useLabels } from '@/hooks/queries';
import { useCreateLabel, useDeleteLabel } from '@/hooks/mutations';
import { Button, TextInput } from '@/components/common/Form';
import { CenteredSpinner, EmptyState } from '@/components/common/Feedback';
import { Icon } from '@/components/common/Icon';
import { labelChipColors } from '@/lib/colors';

const LABEL_COLORS: ReadonlyArray<string> = [
  'oklch(0.44 0.13 255)',
  'oklch(0.52 0.16 27)',
  'oklch(0.46 0.13 305)',
  'oklch(0.47 0.09 80)',
  'oklch(0.48 0.13 345)',
  'oklch(0.42 0.11 165)',
];

export function LabelSettings() {
  const { data: labels, isLoading } = useLabels();
  const createLabel = useCreateLabel();
  const deleteLabel = useDeleteLabel();
  const [name, setName] = useState('');
  const [color, setColor] = useState(LABEL_COLORS[0]);

  const add = () => {
    if (!name.trim()) return;
    createLabel.mutate({ name: name.trim(), color }, { onSuccess: () => setName('') });
  };

  if (isLoading) return <CenteredSpinner />;

  return (
    <div className="flex flex-col gap-4">
      {(labels ?? []).length === 0 && (
        <EmptyState
          icon="label"
          title="ラベルがありません"
          description="ラベルを作成すると手動付与でき、AIの分類候補にもなります。"
        />
      )}

      <div className="flex flex-col gap-2">
        {(labels ?? []).map((l) => {
          const { bg, fg } = labelChipColors(l.color);
          return (
            <div
              key={l.id ?? l.name}
              className="flex items-center gap-3 rounded-lg border border-line px-3 py-2.5"
            >
              <span className="h-3.5 w-3.5 flex-none rounded" style={{ background: l.color }} />
              <span className="flex-1 text-[14px] text-content">{l.name}</span>
              {l.is_system && (
                <span
                  className="flex items-center gap-1 rounded px-1.5 py-0.5 text-[10.5px] font-semibold"
                  style={{ background: bg, color: fg }}
                >
                  <Icon name="auto_awesome" size={12} />
                  AI
                </span>
              )}
              {l.id && (
                <button
                  className="icon-btn-sm"
                  onClick={() => deleteLabel.mutate(l.id!)}
                  aria-label="ラベルを削除"
                >
                  <Icon name="delete" size={18} />
                </button>
              )}
            </div>
          );
        })}
      </div>

      <div className="flex items-center gap-3 rounded-xl border border-line p-3">
        <div className="flex flex-none gap-1.5">
          {LABEL_COLORS.map((c) => (
            <button
              key={c}
              type="button"
              onClick={() => setColor(c)}
              aria-label="色を選択"
              className="h-6 w-6 rounded-full ring-offset-2 ring-offset-surface"
              style={{ background: c, boxShadow: color === c ? '0 0 0 2px var(--text)' : undefined }}
            />
          ))}
        </div>
        <TextInput
          value={name}
          onChange={(e) => setName(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && add()}
          placeholder="新しいラベル名"
          className="flex-1"
        />
        <Button variant="primary" onClick={add} loading={createLabel.isPending} className="h-10">
          追加
        </Button>
      </div>
    </div>
  );
}
