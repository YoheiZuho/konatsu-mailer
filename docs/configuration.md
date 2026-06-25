# 設定リファレンス

konatsu の設定は環境変数で行います。Docker Compose では `.env`（[.env.example](../.env.example) をコピー）で上書きできます。各変数には `docker-compose.yml` 内に既定値があるため、お試し起動なら未設定でも動作しますが、**ネットワークへ公開する前にシークレットは必ず変更**してください。

## バックエンド環境変数

| 変数 | 必須 | 既定値 | 説明 |
| :-- | :-: | :-- | :-- |
| `DATABASE_URL` | ✓ | （Compose が自動設定） | PostgreSQL 接続文字列。例: `postgres://konatsu:konatsu@db:5432/konatsu?sslmode=disable` |
| `MASTER_ENC_KEY` | ✓ | （Compose に開発用既定値） | パスワード/API キー暗号化用の **32 バイト** 鍵（AES-256-GCM）。`openssl rand -hex 16` で生成。**変更すると既存の暗号化データを復号できなくなります。** |
| `JWT_SECRET` | ✓ | （Compose に開発用既定値） | JWT 署名鍵。`openssl rand -base64 32` で生成。 |
| `ALLOW_REGISTRATION` | | `true` | 新規登録の開放/停止。`false` で `POST /api/auth/register` が `403 registration_disabled` を返し、ログイン画面の「新規登録」リンクも非表示になります。 |
| `VAPID_PUBLIC_KEY` | | （空） | Web Push の VAPID 公開鍵。空ならプッシュ通知は無効。 |
| `VAPID_PRIVATE_KEY` | | （空） | VAPID 秘密鍵。 |
| `VAPID_SUBJECT` | | `mailto:admin@example.com` | VAPID の連絡先（`mailto:` または URL）。 |
| `LLM_WORKERS` | | `4` | LLM 解析ワーカープールの同時実行数。 |
| `NOTIFY_THRESHOLD` | | `4` | プッシュ通知を送る重要度の閾値（1〜5）。この値以上で通知。 |
| `LLM_DEFAULT_BASE_URL` | | `https://api.openai.com/v1` | LLM 接続のフォールバック既定値（DB の `llm_configs` が優先）。 |
| `LLM_DEFAULT_MODEL` | | `gpt-4o-mini` | 既定モデル名。 |
| `LLM_DEFAULT_API_KEY` | | （空） | 既定 API キー。 |

> **補足**: `MASTER_ENC_KEY` は実装上「生の文字列バイト列」を鍵として使用し、長さが 32 でなければパディング/切り詰めされます。`openssl rand -hex 16`（= 32 文字 = 32 バイト）が安全かつ確実です。

## フロントエンド環境変数（ビルド時）

`web/` のビルドに使用します（`web/.env.example` も参照）。

| 変数 | 既定値 | 説明 |
| :-- | :-- | :-- |
| `VITE_API_BASE_URL` | `/api` | ブラウザから見た API のベース URL。Compose では nginx が同一オリジンで `/api` をバックエンドへプロキシするため変更不要。 |
| `VITE_DEV_API_TARGET` | `http://localhost:8080` | 開発サーバー（`npm run dev`）が `/api` と WebSocket をプロキシする先。本番ビルドでは未使用。 |

## Compose レベルの変数

| 変数 | 既定値 | 説明 |
| :-- | :-- | :-- |
| `FRONTEND_PORT` | `3000` | Web UI を公開するホストポート（コンテナは常に `:80`）。 |

## ポート

| サービス | コンテナ | ホスト公開 |
| :-- | :-- | :-- |
| frontend (nginx) | `80` | `${FRONTEND_PORT:-3000}` |
| backend (Gin) | `8080` | `8080` |
| db (PostgreSQL) | `5432` | （非公開・内部ネットワークのみ） |

> ローカルでバックエンドを直接実行（`go run`）して Docker の DB に接続したい場合は、`docker-compose.yml` の `db` サービスに `ports: ["5432:5432"]` を一時的に追加してください。詳細は [development.md](development.md)。

## 永続化

PostgreSQL のデータはリポジトリ直下の `./db_data`（バインドマウント）に保存されます。`.gitignore` 済みです。完全初期化するにはコンテナ停止後に `db_data` を削除します。

```bash
docker compose down
rm -rf db_data
```
