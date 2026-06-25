// SPDX-License-Identifier: Apache-2.0
//
// LLM connection settings (design doc §5 / §9.5). OpenAI-compatible endpoints:
// any provider that speaks the Chat Completions API works by changing base_url.

import { useState } from 'react';
import { useLLMConfigs } from '@/hooks/queries';
import {
  useCreateLLMConfig,
  useDeleteLLMConfig,
  useTestLLMConfig,
} from '@/hooks/mutations';
import { Button, Field, Select, TextInput, Toggle } from '@/components/common/Form';
import { CenteredSpinner, EmptyState, InlineError, Spinner } from '@/components/common/Feedback';
import { Icon } from '@/components/common/Icon';
import { ApiRequestError } from '@/lib/api';
import type { LLMConfig, LLMConfigInput, LLMTestResult } from '@/lib/types';

/** base_url presets for common OpenAI-compatible providers (design doc §5.1). */
const PROVIDER_PRESETS: ReadonlyArray<{ name: string; base_url: string; model: string }> = [
  { name: 'OpenAI', base_url: 'https://api.openai.com/v1', model: 'gpt-4o-mini' },
  { name: 'Ollama', base_url: 'http://localhost:11434/v1', model: 'llama3.1' },
  { name: 'LM Studio', base_url: 'http://localhost:1234/v1', model: 'local-model' },
  { name: 'vLLM / llama.cpp', base_url: 'http://localhost:8000/v1', model: '' },
  { name: 'Groq', base_url: 'https://api.groq.com/openai/v1', model: 'llama-3.1-8b-instant' },
  { name: 'OpenRouter', base_url: 'https://openrouter.ai/api/v1', model: '' },
];

const blank: LLMConfigInput = {
  name: '',
  base_url: 'https://api.openai.com/v1',
  model: 'gpt-4o-mini',
  api_key: '',
  temperature: 0.2,
  max_tokens: 512,
  supports_json_schema: true,
  request_timeout_ms: 30000,
  is_default: false,
  is_active: true,
};

export function LLMSettings() {
  const { data: configs, isLoading } = useLLMConfigs();
  const createConfig = useCreateLLMConfig();
  const deleteConfig = useDeleteLLMConfig();
  const [adding, setAdding] = useState(false);
  const [form, setForm] = useState<LLMConfigInput>(blank);
  const [formError, setFormError] = useState<string | null>(null);

  const set = <K extends keyof LLMConfigInput>(key: K, value: LLMConfigInput[K]) =>
    setForm((f) => ({ ...f, [key]: value }));

  const submit = () => {
    setFormError(null);
    createConfig.mutate(
      { ...form, name: form.name || form.model },
      {
        onSuccess: () => {
          setForm(blank);
          setAdding(false);
        },
        onError: (e) =>
          setFormError(e instanceof ApiRequestError ? e.message : '保存に失敗しました。'),
      },
    );
  };

  if (isLoading) return <CenteredSpinner />;

  return (
    <div className="flex flex-col gap-4">
      <p className="text-[12.5px] leading-relaxed text-content-sub">
        OpenAI API 互換のエンドポイントに接続できます。base_url を変更するだけで Ollama / LM
        Studio / vLLM / Groq などローカル・各社モデルへ切り替え可能です。
      </p>

      {(configs ?? []).length === 0 && !adding && (
        <EmptyState
          icon="smart_toy"
          title="AI接続が未設定です"
          description="メールの自動分類・要約・返信案にはLLM接続が必要です。"
        />
      )}

      {(configs ?? []).map((c) => (
        <LLMConfigRow key={c.id} config={c} onDelete={() => deleteConfig.mutate(c.id)} />
      ))}

      {adding ? (
        <div className="flex flex-col gap-4 rounded-xl border border-line p-4">
          <Field label="プロバイダプリセット" hint="選択すると base_url とモデルを補完します。">
            <Select
              defaultValue=""
              onChange={(e) => {
                const p = PROVIDER_PRESETS.find((x) => x.name === e.target.value);
                if (p) setForm((f) => ({ ...f, name: p.name, base_url: p.base_url, model: p.model || f.model }));
              }}
            >
              <option value="">カスタム</option>
              {PROVIDER_PRESETS.map((p) => (
                <option key={p.name} value={p.name}>
                  {p.name}
                </option>
              ))}
            </Select>
          </Field>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field label="表示名">
              <TextInput value={form.name} onChange={(e) => set('name', e.target.value)} placeholder="OpenAI gpt-4o-mini" />
            </Field>
            <Field label="モデル">
              <TextInput value={form.model} onChange={(e) => set('model', e.target.value)} placeholder="gpt-4o-mini" />
            </Field>
            <Field label="base_url">
              <TextInput value={form.base_url} onChange={(e) => set('base_url', e.target.value)} />
            </Field>
            <Field label="API キー" hint="ローカルサーバでは空欄で可。保存後は表示されません。">
              <TextInput
                type="password"
                value={form.api_key ?? ''}
                onChange={(e) => set('api_key', e.target.value)}
                autoComplete="off"
                placeholder="sk-…"
              />
            </Field>
            <Field label="Temperature">
              <TextInput
                type="number"
                step="0.1"
                min="0"
                max="2"
                value={form.temperature}
                onChange={(e) => set('temperature', Number(e.target.value))}
              />
            </Field>
            <Field label="最大トークン">
              <TextInput
                type="number"
                value={form.max_tokens}
                onChange={(e) => set('max_tokens', Number(e.target.value))}
              />
            </Field>
          </div>

          <div className="flex flex-wrap gap-5">
            <label className="flex items-center gap-2.5 text-[13px] text-content">
              <Toggle checked={form.supports_json_schema} onChange={(v) => set('supports_json_schema', v)} />
              JSON Schema 対応
            </label>
            <label className="flex items-center gap-2.5 text-[13px] text-content">
              <Toggle checked={form.is_default} onChange={(v) => set('is_default', v)} />
              既定にする
            </label>
          </div>

          {formError && <InlineError message={formError} />}

          <div className="flex gap-2">
            <Button variant="primary" onClick={submit} loading={createConfig.isPending}>
              保存
            </Button>
            <Button variant="ghost" onClick={() => setAdding(false)}>
              キャンセル
            </Button>
          </div>
        </div>
      ) : (
        <Button variant="secondary" className="self-start" onClick={() => setAdding(true)}>
          <Icon name="add" size={18} />
          AI接続を追加
        </Button>
      )}
    </div>
  );
}

function LLMConfigRow({ config, onDelete }: { config: LLMConfig; onDelete: () => void }) {
  const test = useTestLLMConfig();
  const [result, setResult] = useState<LLMTestResult | null>(null);

  const runTest = () => {
    setResult(null);
    test.mutate(config.id, {
      onSuccess: (r) => setResult(r),
      onError: (e) => setResult({ ok: false, error: e instanceof ApiRequestError ? e.message : '接続失敗' }),
    });
  };

  return (
    <div className="flex flex-col gap-3 rounded-xl border border-line p-4">
      <div className="flex items-center gap-3">
        <Icon name="smart_toy" size={22} className="text-content-sub" />
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="truncate text-[14px] font-semibold text-content">{config.name}</span>
            {config.is_default && (
              <span
                className="rounded-full px-2 py-0.5 text-[10.5px] font-semibold"
                style={{ background: 'var(--brand-weak)', color: 'var(--text)' }}
              >
                既定
              </span>
            )}
          </div>
          <div className="truncate text-[12px] text-content-sub">
            {config.model} · {config.base_url}
          </div>
        </div>
        <button className="icon-btn-sm" onClick={onDelete} aria-label="削除">
          <Icon name="delete" size={19} />
        </button>
      </div>

      <div className="flex items-center gap-3">
        <Button variant="secondary" className="h-8 px-4 text-[13px]" onClick={runTest} loading={test.isPending}>
          接続テスト
        </Button>
        {test.isPending && <Spinner size={16} />}
        {result && (
          <span
            className="flex items-center gap-1.5 text-[12.5px] font-medium"
            style={{ color: result.ok ? 'var(--brand-strong)' : 'var(--prio-high)' }}
          >
            <Icon name={result.ok ? 'check_circle' : 'error'} size={17} />
            {result.ok
              ? `接続成功${result.models?.length ? `（${result.models.length} モデル）` : ''}`
              : result.error || '接続失敗'}
          </span>
        )}
      </div>
    </div>
  );
}
