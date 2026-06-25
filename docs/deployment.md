# デプロイとセキュリティ

konatsu を本番環境で運用する際の要点です。

## デプロイ構成

最小構成は同梱の `docker-compose.yml`（`db` + `backend` + `frontend`）です。フロントエンドの nginx が同一オリジンで `/api`（REST・WebSocket・SSE）をバックエンドへプロキシします。

```
[Internet] → [HTTPS リバースプロキシ] → frontend(nginx:80) ─/api→ backend:8080 → db
```

本番では konatsu の前段に **TLS を終端するリバースプロキシ**（Caddy / Traefik / nginx / クラウド LB）を置き、HTTPS で公開してください。PWA・Service Worker・Web Push は **HTTPS（または localhost）必須**です。

## フロントエンドの API ベース URL（重要）

`VITE_API_BASE_URL` は**ビルド時にバンドルへ埋め込まれ、ブラウザから直接使用**されます。したがって **ブラウザから到達可能な値**でなければなりません。

- ✅ 推奨: `"/api"`（既定）。frontend の nginx が同一オリジンで backend へプロキシするため、CORS 不要・最も堅牢。
- ⚠️ `http://backend:8080` のような **Docker 内部サービス名は使用しない**でください。`backend` はコンテナネットワーク内でのみ解決され、利用者のブラウザからは名前解決できず API 呼び出しが失敗します。
- バックエンドを別ホスト/別ドメインで公開する場合のみ、その**公開 URL**（例 `https://api.example.com`）を指定し、バックエンド側 CORS を適切に設定してください。

```yaml
# docker-compose.yml（frontend）
build:
  args:
    VITE_API_BASE_URL: /api      # 同一オリジン運用（推奨）
```

## シークレット管理

| 項目 | 対応 |
| :-- | :-- |
| `MASTER_ENC_KEY` | 32 バイトのランダム値を安全に保管。**変更すると保存済み暗号化データを復号不能**になるため、ローテーションには再暗号化が必要。 |
| `JWT_SECRET` | 十分に長いランダム値。漏洩時は全セッションが偽造可能になるため要保護。 |
| `VAPID_PRIVATE_KEY` | 秘匿。漏洩時は鍵ペアを再生成（既存購読は無効化）。 |
| `LLM_DEFAULT_API_KEY` / 各 `llm_configs` | DB 内は AES-256-GCM で暗号化保存。 |

`.env` はリポジトリにコミットしないでください（`.gitignore` 済み）。本番ではシークレットマネージャや Docker secrets の利用を推奨します。

## 新規登録の制御

公開インスタンスで不特定の登録を防ぐには `ALLOW_REGISTRATION=false` を設定します。最初の管理ユーザーを作成後に閉じる運用が安全です。

## Web Push (VAPID)

```bash
npx web-push generate-vapid-keys
```

生成した公開鍵/秘密鍵を `VAPID_PUBLIC_KEY` / `VAPID_PRIVATE_KEY` に、連絡先を `VAPID_SUBJECT`（`mailto:` または URL）に設定します。

## ストリーミング/WebSocket とタイムアウト（注意）

AI 下書きの SSE（`/api/ai/draft`）と WebSocket（`/api/ws`）は長時間接続です。プロキシ側はこれらのパスでバッファリング無効・長いタイムアウトが必要です（同梱の `web/nginx.conf` は `/api` で `proxy_buffering off` と長い `proxy_read_timeout` を設定済み）。

> バックエンドの `http.Server` は SSE/WebSocket のために Read/Write タイムアウトを無効化し（`ReadHeaderTimeout` のみハンドシェイク保護に使用）、長時間接続を維持します。

## セキュリティチェックリスト

- [ ] HTTPS で公開（TLS 終端プロキシ）
- [ ] `MASTER_ENC_KEY` / `JWT_SECRET` を強力なランダム値に変更
- [ ] `ALLOW_REGISTRATION` を運用方針に合わせて設定
- [ ] IMAP/SMTP は TLS/STARTTLS を使用（既定 ON）
- [ ] PostgreSQL を外部公開しない（既定で内部のみ）
- [ ] `VITE_API_BASE_URL` がブラウザから到達可能（推奨 `/api`）
- [ ] `.env` を秘匿・バックアップ
- [ ] LLM にローカルエンドポイントを使う場合、本文が外部に出ないことを確認

## バックアップ

- **DB**: `./db_data`（バインドマウント）。`pg_dump` での論理バックアップを推奨。
  ```bash
  docker compose exec db pg_dump -U konatsu konatsu > backup.sql
  ```
- **シークレット**: `MASTER_ENC_KEY` を失うと暗号化データを復号できません。DB バックアップと合わせて安全に保管してください。

## アップデート

```bash
git pull
docker compose up --build -d
```

マイグレーションはバックエンド起動時に自動適用されます。
