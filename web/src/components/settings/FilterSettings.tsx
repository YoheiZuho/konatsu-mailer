// SPDX-License-Identifier: Apache-2.0
//
// Message-filter (auto-classification) management, modeled on Thunderbird's
// "message filters": each rule has conditions (all/any) and actions.

import { useState } from 'react';
import { useFilters, useCreateFilter, useUpdateFilter, useDeleteFilter } from '@/hooks/filters';
import { useFolders } from '@/hooks/queries';
import { Button, Field, Select, TextInput, Toggle, SegmentedControl } from '@/components/common/Form';
import { Icon } from '@/components/common/Icon';
import { CenteredSpinner, EmptyState } from '@/components/common/Feedback';
import type {
  FilterAction,
  FilterActionType,
  FilterCondition,
  FilterField,
  FilterOp,
  FilterRule,
  FilterRuleInput,
} from '@/lib/types';

const FIELDS: { value: FilterField; label: string }[] = [
  { value: 'subject', label: '件名' },
  { value: 'from', label: '差出人' },
  { value: 'to', label: '宛先' },
  { value: 'cc', label: 'Cc' },
  { value: 'recipient', label: '宛先・Cc・Bcc' },
  { value: 'body', label: '本文（プレビュー）' },
];
const OPS: { value: FilterOp; label: string }[] = [
  { value: 'contains', label: 'に次を含む' },
  { value: 'not_contains', label: 'に次を含まない' },
  { value: 'is', label: 'が次と一致する' },
  { value: 'is_not', label: 'が次と異なる' },
  { value: 'starts_with', label: 'が次で始まる' },
  { value: 'ends_with', label: 'が次で終わる' },
];
const ACTIONS: { value: FilterActionType; label: string }[] = [
  { value: 'move_folder', label: 'フォルダへ移動' },
  { value: 'add_label', label: 'ラベルを付与' },
  { value: 'set_category', label: 'カテゴリを設定' },
  { value: 'mark_read', label: '既読にする' },
  { value: 'star', label: 'スターを付ける' },
];
const CATEGORY_OPTIONS = [
  { value: 'primary', label: 'メイン' },
  { value: 'promotions', label: 'プロモーション' },
  { value: 'social', label: 'ソーシャル' },
  { value: 'newsletters', label: 'ニュースレター' },
];

const emptyDraft: FilterRuleInput = {
  name: '',
  enabled: true,
  match_type: 'all',
  conditions: [{ field: 'subject', op: 'contains', value: '' }],
  actions: [{ type: 'move_folder', value: '' }],
};

export function FilterSettings() {
  const { data: filters, isLoading } = useFilters();
  const createFilter = useCreateFilter();
  const updateFilter = useUpdateFilter();
  const deleteFilter = useDeleteFilter();

  const [editing, setEditing] = useState<{ id?: string; draft: FilterRuleInput } | null>(null);

  if (isLoading) return <CenteredSpinner />;

  if (editing) {
    return (
      <FilterEditor
        initial={editing.draft}
        saving={createFilter.isPending || updateFilter.isPending}
        onCancel={() => setEditing(null)}
        onSave={(draft) => {
          const done = () => setEditing(null);
          if (editing.id) updateFilter.mutate({ id: editing.id, ...draft }, { onSuccess: done });
          else createFilter.mutate(draft, { onSuccess: done });
        }}
      />
    );
  }

  return (
    <div className="flex flex-col gap-3">
      <p className="text-[12.5px] leading-relaxed text-content-sub">
        受信したメールを条件に応じて自動的に分類・整理します（フォルダ移動・ラベル付与・既読化など）。
        フィルターは新着メール受信時に上から順に評価されます。
      </p>

      {(filters ?? []).length === 0 ? (
        <EmptyState icon="filter_alt" title="フィルターがありません" description="条件と動作を設定して自動分類しましょう。" />
      ) : (
        <div className="flex flex-col gap-2">
          {(filters ?? []).map((f) => (
            <div key={f.id} className="flex items-center gap-3 rounded-xl border border-line p-3">
              <Icon name="filter_alt" size={20} className="text-content-sub" />
              <div className="min-w-0 flex-1">
                <div className="truncate text-[14px] font-semibold text-content">{f.name || '(無題)'}</div>
                <div className="truncate text-[12px] text-content-sub">
                  {f.conditions.length} 条件 · {f.actions.length} 動作
                </div>
              </div>
              {!f.enabled && (
                <span className="rounded-full bg-surface-sub px-2 py-0.5 text-[11px] text-content-sub">無効</span>
              )}
              <button className="icon-btn-sm" onClick={() => setEditing({ id: f.id, draft: toInput(f) })} aria-label="編集">
                <Icon name="edit" size={18} />
              </button>
              <button className="icon-btn-sm" onClick={() => deleteFilter.mutate(f.id)} aria-label="削除">
                <Icon name="delete" size={18} />
              </button>
            </div>
          ))}
        </div>
      )}

      <Button variant="secondary" className="self-start" onClick={() => setEditing({ draft: { ...emptyDraft } })}>
        <Icon name="add" size={18} />
        フィルターを追加
      </Button>
    </div>
  );
}

function toInput(f: FilterRule): FilterRuleInput {
  return {
    name: f.name,
    enabled: f.enabled,
    match_type: f.match_type,
    conditions: f.conditions.length ? f.conditions : [{ field: 'subject', op: 'contains', value: '' }],
    actions: f.actions.length ? f.actions : [{ type: 'move_folder', value: '' }],
  };
}

function FilterEditor({
  initial,
  saving,
  onSave,
  onCancel,
}: {
  initial: FilterRuleInput;
  saving: boolean;
  onSave: (draft: FilterRuleInput) => void;
  onCancel: () => void;
}) {
  const [draft, setDraft] = useState<FilterRuleInput>(initial);
  const { data: folders } = useFolders();

  const setConditions = (conditions: FilterCondition[]) => setDraft((d) => ({ ...d, conditions }));
  const setActions = (actions: FilterAction[]) => setDraft((d) => ({ ...d, actions }));

  return (
    <div className="flex flex-col gap-5">
      <Field label="フィルター名">
        <TextInput value={draft.name} onChange={(e) => setDraft({ ...draft, name: e.target.value })} placeholder="例：請求書を整理" />
      </Field>

      <div className="flex items-center justify-between">
        <span className="text-[12.5px] font-medium text-content-sub">条件の一致</span>
        <SegmentedControl<'all' | 'any'>
          value={draft.match_type}
          onChange={(v) => setDraft({ ...draft, match_type: v })}
          options={[
            { value: 'all', label: 'すべてに一致' },
            { value: 'any', label: 'いずれかに一致' },
          ]}
        />
      </div>

      <Field label="条件">
        <div className="flex flex-col gap-2">
          {draft.conditions.map((cond, i) => (
            <div key={i} className="flex items-center gap-2">
              <Select
                value={cond.field}
                onChange={(e) => setConditions(draft.conditions.map((c, j) => (i === j ? { ...c, field: e.target.value as FilterField } : c)))}
                className="w-[34%]"
              >
                {FIELDS.map((f) => (
                  <option key={f.value} value={f.value}>{f.label}</option>
                ))}
              </Select>
              <Select
                value={cond.op}
                onChange={(e) => setConditions(draft.conditions.map((c, j) => (i === j ? { ...c, op: e.target.value as FilterOp } : c)))}
                className="w-[30%]"
              >
                {OPS.map((o) => (
                  <option key={o.value} value={o.value}>{o.label}</option>
                ))}
              </Select>
              <TextInput
                value={cond.value}
                onChange={(e) => setConditions(draft.conditions.map((c, j) => (i === j ? { ...c, value: e.target.value } : c)))}
                className="min-w-0 flex-1"
                placeholder="値"
              />
              <RowButtons
                onAdd={() => setConditions([...draft.conditions, { field: 'subject', op: 'contains', value: '' }])}
                onRemove={draft.conditions.length > 1 ? () => setConditions(draft.conditions.filter((_, j) => j !== i)) : undefined}
              />
            </div>
          ))}
        </div>
      </Field>

      <Field label="動作">
        <div className="flex flex-col gap-2">
          {draft.actions.map((action, i) => (
            <div key={i} className="flex items-center gap-2">
              <Select
                value={action.type}
                onChange={(e) => setActions(draft.actions.map((a, j) => (i === j ? { type: e.target.value as FilterActionType, value: '' } : a)))}
                className="w-[40%]"
              >
                {ACTIONS.map((a) => (
                  <option key={a.value} value={a.value}>{a.label}</option>
                ))}
              </Select>
              <div className="min-w-0 flex-1">
                <ActionValue
                  action={action}
                  folders={(folders?.items ?? []).map((f) => f.name)}
                  onChange={(value) => setActions(draft.actions.map((a, j) => (i === j ? { ...a, value } : a)))}
                />
              </div>
              <RowButtons
                onAdd={() => setActions([...draft.actions, { type: 'mark_read', value: '' }])}
                onRemove={draft.actions.length > 1 ? () => setActions(draft.actions.filter((_, j) => j !== i)) : undefined}
              />
            </div>
          ))}
        </div>
      </Field>

      <div className="flex items-center gap-3">
        <label className="flex items-center gap-2.5 text-[13px] text-content">
          <Toggle checked={draft.enabled} onChange={(v) => setDraft({ ...draft, enabled: v })} />
          有効
        </label>
        <div className="flex-1" />
        <Button variant="ghost" onClick={onCancel}>キャンセル</Button>
        <Button variant="primary" onClick={() => onSave(draft)} loading={saving} disabled={!draft.name.trim()}>
          保存
        </Button>
      </div>
    </div>
  );
}

function ActionValue({
  action,
  folders,
  onChange,
}: {
  action: FilterAction;
  folders: string[];
  onChange: (value: string) => void;
}) {
  if (action.type === 'mark_read' || action.type === 'star') {
    return <span className="block px-1 text-[12.5px] text-content-sub">（値は不要）</span>;
  }
  if (action.type === 'move_folder') {
    return (
      <Select value={action.value} onChange={(e) => onChange(e.target.value)} className="w-full">
        <option value="">フォルダーを選択…</option>
        {folders.map((f) => (
          <option key={f} value={f}>{f}</option>
        ))}
      </Select>
    );
  }
  if (action.type === 'set_category') {
    return (
      <Select value={action.value} onChange={(e) => onChange(e.target.value)} className="w-full">
        <option value="">カテゴリを選択…</option>
        {CATEGORY_OPTIONS.map((c) => (
          <option key={c.value} value={c.value}>{c.label}</option>
        ))}
      </Select>
    );
  }
  // add_label
  return <TextInput value={action.value} onChange={(e) => onChange(e.target.value)} className="w-full" placeholder="ラベル名" />;
}

function RowButtons({ onAdd, onRemove }: { onAdd: () => void; onRemove?: () => void }) {
  return (
    <div className="flex flex-none gap-1">
      <button onClick={onAdd} className="icon-btn-sm" aria-label="行を追加">
        <Icon name="add" size={18} />
      </button>
      <button onClick={onRemove} disabled={!onRemove} className="icon-btn-sm disabled:opacity-30" aria-label="行を削除">
        <Icon name="remove" size={18} />
      </button>
    </div>
  );
}
