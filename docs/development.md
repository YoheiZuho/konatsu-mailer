# 開発ガイド

## 前提

- Go 1.23+
- Node.js 20+（22 推奨）
- Docker / Docker Compose v2（DB 用）
- （任意）`golang-migrate` CLI、`golangci-lint`

## ディレクトリ構成

```
konatsu-mailer/
├─ cmd/server/main.go      # エントリポイント（API + 同期 + 解析を起動）
├─ internal/
│  ├─ config/              # 環境変数のロード
│  ├─ domain/              # エンティティ・値オブジェクト
│  ├─ store/               # PostgreSQL（pgx）アクセス・マイグレーション実行
│  ├─ imapsync/            # IMAP IDLE 同期ワーカー・MIME パース
│  ├─ smtpsend/            # SMTP 送信
│  ├─ llm/                 # OpenAI 互換クライアント・分類/要約・プロンプト
│  ├─ api/                 # Gin ハンドラ・ルーティング・JWT・ミドルウェア
│  ├─ ws/                  # WebSocket Hub
│  ├─ push/                # Web Push (VAPID)
│  └─ crypto/              # AES-256-GCM ラッパ
├─ migrations/             # golang-migrate 用 SQL
├─ web/                    # フロントエンド（React + Vite）→ web/README.md
├─ docker-compose.yml
├─ Dockerfile              # バックエンドのマルチステージビルド
└─ docs/
```

フロントエンドの構成は [web/README.md](../web/README.md) を参照。

## 起動方法

### A. 全部 Docker（最も簡単）

```bash
docker compose up --build      # または make run
```

### B. バックエンドはローカル、DB は Docker

1. DB だけ起動し、ポートを公開します（`docker-compose.yml` の `db` に一時的に `ports: ["5432:5432"]` を追加）。
   ```bash
   docker compose up -d db
   ```
2. 環境変数を設定して起動します。
   ```bash
   export DATABASE_URL="postgres://konatsu:konatsu@localhost:5432/konatsu?sslmode=disable"
   export MASTER_ENC_KEY="$(openssl rand -hex 16)"
   export JWT_SECRET="$(openssl rand -base64 32)"
   go run ./cmd/server
   ```
   起動時にマイグレーションが自動適用されます。

### C. フロントエンド開発サーバー

バックエンド（A か B）が起動している状態で:

```bash
cd web
npm install
npm run dev        # http://localhost:5173
```

Vite が `/api` と WebSocket を `http://localhost:8080`（`VITE_DEV_API_TARGET` で変更可）へプロキシします。

## マイグレーション

マイグレーションは **バックエンド起動時に自動適用**されます（`store.Migrate`）。手動で操作したい場合は [golang-migrate](https://github.com/golang-migrate/migrate) CLI を使います。

```bash
migrate -path ./migrations -database "$DATABASE_URL" up
migrate -path ./migrations -database "$DATABASE_URL" down 1
```

> 本番用の distroless バックエンドイメージには `migrate` バイナリは含まれません（自動適用のため不要）。

## ビルド・検査

```bash
# バックエンド
go build ./...
go vet ./...
go test -race -count=1 ./...     # make test
go fmt ./...                     # make fmt
golangci-lint run ./...          # make lint（要 golangci-lint）

# フロントエンド
cd web
npm run typecheck
npm run build
```

## Makefile

| ターゲット | 内容 |
| :-- | :-- |
| `make build` | `docker compose build` |
| `make run` | `docker compose up --build` |
| `make test` | `go test -race ./...` |
| `make lint` | `golangci-lint run ./...` |
| `make fmt` | `go fmt ./...` |

## PWA アイコンの再生成

ブランドマークを変えた場合、依存ライブラリなしの生成スクリプトで PNG を作り直します。

```bash
cd web
node scripts/generate-icons.mjs   # public/icons/*.png を更新
```

## 実装状況メモ

- **メール同期**: `internal/imapsync` が `main.go` から起動され、`is_active` な各アカウントに 1 goroutine を割り当てます。現状は **ポーリング方式（既定 30 秒間隔で最新メールを取得）** の MVP で、UID 競合は upsert で冪等化しています。接続は実装 TLS（`imap_use_tls=true`、993）／STARTTLS（false、143）の両対応。IMAP IDLE はフォローアップ予定。
- **フォルダ**: 同期時に IMAP の `LIST`（SPECIAL-USE）で実フォルダを取得し、`accounts.folders` に保存。INBOX に加え特殊用途フォルダ（Sent / Junk(迷惑メール) / Trash / Drafts / Archive）を同期します。サイドバーは `GET /api/folders` の実フォルダを表示。任意のカスタムフォルダの本格同期はフォローアップ（現状は特殊用途＋INBOX）。
- **LLM 解析パイプライン**: `internal/analysis` がワーカープール（`LLM_WORKERS`）で新着メールを解析。`internal/llm`（go-openai, OpenAI 互換）で要約・重要度(1-5)・ラベル・スパム判定を生成し、`emails.ai_summary/ai_priority` と `email_labels`(source=ai) に保存、`MAIL_ANALYZED` を配信。`prefs.ai_filters` のカテゴリは `shouldAnalyze` でスキップ（§6.3）。
- **プッシュ送信**: 重要度 ≥ `NOTIFY_THRESHOLD` で `internal/push`（webpush-go/VAPID）が AI 要約付き通知を送信。`prefs.push_labels` を選択時はそのラベルが付いたメールのみ通知。404/410 の購読は自動削除。
- **AI 下書き/返信案**: `POST /api/ai/draft` が既定 LLM 接続でストリーミング生成（SSE）。
- **LLM 接続設定 / ラベル**: `/api/llm-configs`（CRUD＋`/test`）と `/api/labels`（CRUD）を実装。API キーは AES-256-GCM で暗号化保存。
- **カテゴリ分類**: 同期時に分類ヘッダ（List-Id / List-Unsubscribe / Precedence / Auto-Submitted）と送信元から `emails.category`（primary/promotions/social/newsletters）をヒューリスティック判定。受信トレイの**カテゴリタブ**（メイン/プロモーション/ソーシャル/ニュースレター）で絞り込み。
- **メッセージフィルタ（自動分類）**: `filters` テーブル＋`GET/POST/PATCH/DELETE /api/filters`。新着メール受信時に条件（件名/差出人/宛先/Cc/本文・含む/一致/前方一致等、すべて/いずれか）を評価し、アクション（フォルダ移動=IMAP MOVE / ラベル付与 / 既読化=IMAP `\Seen` / スター / カテゴリ設定）を適用。設定→「フィルター」で管理。
- **同期件数**: 1 フォルダあたり最新 300 件まで取得（`imapsync.initialFetch`）。一覧は keyset カーソルで無限スクロール。
- **本文の文字化け対策**: IMAP 由来文字列は保存前に妥当な UTF-8 へサニタイズ（部分取得や非 UTF-8 charset による不正バイト列を除去）。
- **送信**: `internal/smtpsend` は SMTPS（ポート 465 の実装 TLS）と STARTTLS（587）の両対応。
- **リアルタイム**: `internal/ws` の Hub が `NEW_MAIL` / `SYNC_STATUS` を配信。`/api/ws` は `coder/websocket` でアップグレード。
- **未実装（フォローアップ）**: 添付ファイルのダウンロード、本文の S3 保存、LLM プロバイダ単位のレート制御（トークンバケット）、カスタムフォルダのオンデマンド同期。
- フロントエンドの API 契約は [詳細設計.md §7/§8](../詳細設計.md) が一次情報です。
