// SPDX-License-Identifier: Apache-2.0

import clsx from 'clsx';
import { useAppearance, BRAND_PRESETS } from '@/stores/appearance';
import { useTranslateConfig, useTranslateLanguages, COMMON_TARGETS } from '@/hooks/translate';
import { Field, Select, SegmentedControl, Toggle } from '@/components/common/Form';
import { Icon } from '@/components/common/Icon';
import type { Density, ThemePref } from '@/lib/types';

export function AppearanceSettings() {
  const {
    theme,
    brand,
    density,
    aiSummaries,
    translateTarget,
    setTheme,
    setBrand,
    setDensity,
    setAiSummaries,
    setTranslateTarget,
  } = useAppearance();

  const translateConfig = useTranslateConfig();
  const translationEnabled = translateConfig.data?.enabled ?? false;
  const languages = useTranslateLanguages(translationEnabled);
  const targetOptions =
    languages.data && languages.data.length > 0 ? languages.data : COMMON_TARGETS;

  return (
    <div className="flex flex-col gap-7">
      <Field label="テーマ">
        <SegmentedControl<ThemePref>
          value={theme}
          onChange={setTheme}
          options={[
            { value: 'system', label: 'システム' },
            { value: 'light', label: 'ライト' },
            { value: 'dark', label: 'ダーク' },
          ]}
        />
      </Field>

      <Field label="キーカラー" hint="ボタン・選択中の行・バッジなどに使用されます。">
        <div className="flex flex-wrap items-center gap-2.5">
          {BRAND_PRESETS.map((p) => (
            <button
              key={p.hex}
              type="button"
              title={p.name}
              onClick={() => setBrand(p.hex)}
              className={clsx(
                'flex h-9 w-9 items-center justify-center rounded-full ring-offset-2 ring-offset-surface transition',
                brand.toLowerCase() === p.hex.toLowerCase() && 'ring-2 ring-content/40',
              )}
              style={{ background: p.hex }}
            >
              {brand.toLowerCase() === p.hex.toLowerCase() && (
                <Icon name="check" size={18} style={{ color: '#fff', mixBlendMode: 'difference' }} />
              )}
            </button>
          ))}
          <label
            className="flex h-9 w-9 cursor-pointer items-center justify-center rounded-full border border-line"
            title="カスタム"
          >
            <Icon name="palette" size={18} className="text-content-sub" />
            <input
              type="color"
              value={brand}
              onChange={(e) => setBrand(e.target.value)}
              className="absolute h-0 w-0 opacity-0"
              aria-label="カスタムキーカラー"
            />
          </label>
        </div>
      </Field>

      <Field label="表示密度">
        <SegmentedControl<Density>
          value={density}
          onChange={setDensity}
          options={[
            { value: 'comfortable', label: 'ゆったり' },
            { value: 'compact', label: 'コンパクト' },
          ]}
        />
      </Field>

      <div className="flex items-center justify-between">
        <div>
          <div className="text-[13.5px] font-medium text-content">AI要約を表示</div>
          <div className="text-[12px] text-content-sub">一覧・本文にAIの要約を表示します。</div>
        </div>
        <Toggle checked={aiSummaries} onChange={setAiSummaries} label="AI要約を表示" />
      </div>

      {translationEnabled && (
        <Field label="翻訳先の言語" hint="メール本文の「翻訳」で使用する言語です。">
          <Select
            value={translateTarget}
            onChange={(e) => setTranslateTarget(e.target.value)}
            className="max-w-xs"
          >
            {targetOptions.map((l) => (
              <option key={l.code} value={l.code}>
                {l.name}（{l.code}）
              </option>
            ))}
          </Select>
        </Field>
      )}
    </div>
  );
}
