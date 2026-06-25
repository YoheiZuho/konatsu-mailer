# LLM（AI 接続）の設定

konatsu のメール自動分類・要約・返信案は **OpenAI API 互換の Chat Completions エンドポイント**を利用します。`base_url` を変えるだけで、クラウド／ローカルの各種 LLM に切り替えられます。

> 設計の詳細: [詳細設計.md §5](../詳細設計.md)

## 設定手順

1. 右上 ⚙️（設定）→ **「AI 接続」** タブを開きます。
2. 「AI 接続を追加」を押します。
3. **プロバイダプリセット** を選ぶと `base_url` とモデルが補完されます（カスタムも可）。
4. 必要項目を入力します。
   - **表示名** — 任意の識別名
   - **モデル** — 例 `gpt-4o-mini`, `llama3.1`
   - **base_url** — エンドポイント（末尾 `/v1`）
   - **API キー** — クラウドは必須、ローカルは空欄で可
   - **Temperature / 最大トークン**
   - **JSON Schema 対応** — 構造化出力に対応するモデルはオン
   - **既定にする** — このユーザーの解析に使う既定設定にする
5. 保存後、各設定の **「接続テスト」** で疎通を確認します。

## 対応プロバイダ例

| プロバイダ | base_url 例 | API キー | 備考 |
| :-- | :-- | :-- | :-- |
| OpenAI | `https://api.openai.com/v1` | 必須 | JSON Schema 対応 |
| Azure OpenAI | `https://{resource}.openai.azure.com/openai/deployments/{dep}` | 必須 | `api-version` クエリが必要 |
| Ollama | `http://localhost:11434/v1` | 不要 | ローカル。OpenAI 互換 |
| LM Studio | `http://localhost:1234/v1` | 不要 | ローカル |
| vLLM / llama.cpp | `http://localhost:8000/v1` | 任意 | ローカル |
| Groq | `https://api.groq.com/openai/v1` | 必須 | JSON Schema 可否はモデル依存 |
| OpenRouter | `https://openrouter.ai/api/v1` | 必須 | 同上 |

> Docker コンテナ内のバックエンドから**ホスト上**のローカル LLM（Ollama 等）へ接続する場合、`base_url` のホストは `localhost` ではなく `host.docker.internal`（macOS/Windows）を使うか、LLM を同じ Docker ネットワークに配置してください。

## 接続テストの挙動

`POST /api/llm-configs/:id/test` は次の順で疎通確認します（[詳細設計 付録A](../詳細設計.md)）。

1. `GET {base_url}/models` を試行（成功すればモデル一覧を返却）
2. 失敗時は `POST {base_url}/chat/completions` に最小リクエストを送信
3. 結果を `{ ok: true, models: [...] }` または `{ ok: false, error: "..." }` として返し、UI にバッジ表示

## 構造化出力（3 段フォールバック）

互換エンドポイントごとに JSON 制約のサポート度が異なるため、段階的に縮退します。

1. **json_schema**（推奨, `supports_json_schema=true`）
2. **json_object**（多くのローカルサーバが対応）
3. **プロンプトのみ**（応答から JSON を抽出、失敗時 1 回リトライ）

いずれも失敗した場合は当該メールの解析状態を `error` とし、**AI 要約なしで通常表示**します（LLM 障害がメール閲覧をブロックしません）。

## コスト・負荷の最適化

- LLM を回す前にフィルタ層で対象を選別します（未知の送信者・特定キーワード・添付・自分宛など）。既知のニュースレターや自動送信は `skipped` になります。
- プロバイダ単位のトークンバケットで同時実行/レートを制御します（ローカル GPU の過負荷回避にも有効）。
- 本文はトークン上限まで切り詰めて送信します。

## プライバシー

`base_url` をローカル LLM に向ければ、メール本文を社外（外部 API）に送信せずに解析できます。機微情報を扱う組織での運用に適しています。
