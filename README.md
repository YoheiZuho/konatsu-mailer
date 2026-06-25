# konatsu

**AI 駆動型リッチ Web メールクライアント。**

従来の Web メール（リクエストベースの同期）とは異なり、バックグラウンドで常時メールサーバーを監視（IMAP IDLE）し、リアルタイムな新着検知と LLM による自動分類・要約・返信案を提供するモダンなメール体験を目指したアプリケーションです。

- **バックエンド**: Go (Gin) — REST API・WebSocket・IMAP 同期ワーカー・LLM 解析パイプラインを単一バイナリで提供
- **フロントエンド**: React + TypeScript + Vite — Gmail ライクな三ペイン UI、インストール可能な PWA
- **データベース**: PostgreSQL 16+
- **LLM**: OpenAI API 互換エンドポイント（OpenAI / Azure / Ollama / LM Studio / vLLM / Groq / OpenRouter など）に設定変更のみで接続

---

## 主な機能

- 🔄 **リアルタイム同期** — IMAP IDLE による新着メールの即時検知
- 🤖 **AI 連携** — メールの自動分類・ラベル付与・重要度判定・1 行要約、「AI で返信案／下書き」生成（ストリーミング）
- ⚡ **高速 UI** — 仮想スクロールのメール一覧、楽観的更新、WebSocket によるちらつきのない反映
- 🎨 **テーマ & キーカラー** — ライト/ダーク/システム連動、ブランドカラーをユーザーが自由に選択
- 📱 **PWA** — ホーム画面へインストール、オフライン起動、Web Push 通知（重要メールを AI 要約付きで通知）
- 🔌 **OpenAI 互換 LLM** — `base_url` を変えるだけでクラウド／ローカル LLM を切替（社外にメール本文を出さない運用も可能）
- 🔐 **暗号化** — メール／API のパスワードは AES-256-GCM で暗号化保存

---

## 必要環境

- [Docker](https://docs.docker.com/get-docker/) および Docker Compose v2
- 利用したいメールアカウントの **IMAP/SMTP 情報**
- （任意）LLM 機能を使う場合：OpenAI API キー、もしくはローカル LLM（Ollama など）

---

## クイックスタート

```bash
# 1. 取得
git clone <repository-url> konatsu-mailer
cd konatsu-mailer

# 2. 設定（任意だが本番では必須）。シークレットを生成して .env に記入
cp .env.example .env
#   MASTER_ENC_KEY:  openssl rand -hex 16
#   JWT_SECRET:      openssl rand -base64 32

# 3. 起動（PostgreSQL + バックエンド + フロントエンド）
docker compose up --build
#   または: make run
```

起動後、ブラウザで **http://localhost:3000** を開きます（API は `:8080`）。

---

## 基本的な使い方

1. **アカウント作成** — 初回アクセス時に「新規登録」からユーザーを作成します。
   この画面で **IMAP/SMTP のメールアカウントも同時に設定**できます（後から設定画面でも追加可能）。
2. **メールアカウントの追加** — 右上 ⚙️（設定）→「アカウント」から IMAP/SMTP を登録すると、バックグラウンド同期が始まります。
3. **AI 接続の設定** — 設定 →「AI 接続」で LLM のエンドポイントを登録します。
   プリセット（OpenAI / Ollama / LM Studio など）から選び、「接続テスト」で疎通確認できます。→ 詳細は [docs/llm-setup.md](docs/llm-setup.md)
4. **メールを読む・書く** — 左の一覧から選ぶと右に本文と AI 要約が表示されます。「作成」または返信欄の「AI で返信案」で下書きを生成できます。
5. **外観のカスタマイズ** — 設定 →「外観」でテーマ・キーカラー・表示密度・AI 要約の表示を変更できます。
6. **通知の有効化** — 設定 →「通知」でプッシュ通知を許可すると、重要メールが届いたときに AI 要約付きで通知されます。

詳しい操作は [docs/usage.md](docs/usage.md) を参照してください。

---

## ドキュメント

| ドキュメント | 内容 |
| :-- | :-- |
| [docs/getting-started.md](docs/getting-started.md) | セットアップ手順（鍵の生成・初回起動・最初のログイン） |
| [docs/usage.md](docs/usage.md) | 画面の使い方（一覧・本文・作成・ラベル・テーマ・通知・PWA） |
| [docs/configuration.md](docs/configuration.md) | 環境変数リファレンス |
| [docs/llm-setup.md](docs/llm-setup.md) | LLM（OpenAI 互換）接続の設定 |
| [docs/architecture.md](docs/architecture.md) | システム構成・データフロー |
| [docs/development.md](docs/development.md) | 開発環境・ビルド・ディレクトリ構成 |
| [docs/deployment.md](docs/deployment.md) | 本番デプロイとセキュリティ |

---

## 技術スタック

| レイヤ | 採用技術 |
| :-- | :-- |
| Backend | Go 1.25+ / Gin / pgx / golang-migrate |
| IMAP/SMTP | go-imap (IDLE) / go-message |
| LLM | OpenAI 互換 Chat Completions API |
| Realtime | WebSocket / Server-Sent Events |
| Web Push | VAPID (webpush-go) |
| Frontend | React 18 / TypeScript / Vite |
| State | TanStack Query v5 / Zustand |
| Style | Tailwind CSS v3（CSS 変数によるテーマ） |
| PWA | vite-plugin-pwa (Workbox) |
| DB | PostgreSQL 16+ |

---

## ライセンス

[Apache License 2.0](LICENSE)。詳細は [`NOTICE`](NOTICE) も参照してください。
