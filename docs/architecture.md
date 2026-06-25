# アーキテクチャ

konatsu の全体構成の概要です。網羅的な設計は [基本設計.md](../基本設計.md) / [詳細設計.md](../詳細設計.md) を参照してください。

## コンポーネント

```
            ┌──────────────────────── Browser (PWA) ────────────────────────┐
            │  React + TanStack Query + Zustand                             │
            │   REST / WebSocket / SSE  ・  Service Worker（push・cache）    │
            └───────────────┬───────────────────────────────┬───────────────┘
                            │ (same origin)                 │ Web Push
                ┌───────────▼───────────┐                   │
                │   frontend (nginx)     │  /api を proxy     │
                └───────────┬───────────┘                   │
                            │                                │
   ┌────────────────────────▼─────────────────────────┐     │
   │                 backend (Go, :8080)               │     │
   │  ┌─────────────┐ ┌──────────────┐ ┌────────────┐  │     │
   │  │ Gin API +   │ │ SyncManager  │ │ Analysis   │  │     │
   │  │ WebSocket   │ │ (IMAP IDLE)  │ │ Pipeline   │──┼─────┘ (VAPID)
   │  │ Hub         │ │ per account  │ │ (LLM pool) │  │
   │  └──────┬──────┘ └──────┬───────┘ └─────┬──────┘  │
   └─────────┼───────────────┼───────────────┼─────────┘
             │               │               │
        ┌────▼────┐   ┌───────▼───────┐  ┌────▼─────────────┐
        │ Postgres │   │ Mail server   │  │ OpenAI 互換 LLM  │
        │          │   │ (IMAP/SMTP)   │  │ (cloud / local)  │
        └──────────┘   └───────────────┘  └──────────────────┘
```

単一の Go バイナリ内で 3 つのサブシステムを goroutine として動かします。

- **API Server (Gin)** — REST + WebSocket アップグレード
- **SyncManager** — `is_active` な各アカウントに 1 goroutine で IMAP IDLE 接続を保持し、新着を検知して DB 保存
- **AnalysisPipeline** — LLM 解析ジョブを消費するワーカープール

## データフロー

```
IMAP IDLE → 新着検知 → MIME パース → emails 保存
   → ws.Broadcast(NEW_MAIL)
   → JobQueue.Enqueue(Analyze)        ※フィルタ層を通過した場合のみ
AnalysisPipeline:
   → llm.Classify（OpenAI 互換 /chat/completions）
   → 要約・重要度・ラベルを保存
   → ws.Broadcast(MAIL_ANALYZED)
   → push.Notify（重要度 ≥ NOTIFY_THRESHOLD）
```

フロントエンドは WebSocket イベントを受けて TanStack Query のキャッシュを更新します。

- `NEW_MAIL` → 一覧を再取得（invalidate）
- `MAIL_ANALYZED` → 該当行をキャッシュ上で部分更新（再取得なし＝ちらつき防止）
- `SYNC_STATUS` → TopBar の同期バッジを更新

## リアルタイム通信

| 用途 | 仕組み |
| :-- | :-- |
| 新着・解析・同期状態の通知 | WebSocket `GET /api/ws`（トークンは `Sec-WebSocket-Protocol`） |
| AI 返信案／下書きの逐次生成 | Server-Sent Events `POST /api/ai/draft` |
| ブラウザを閉じても届く通知 | Web Push (VAPID) |

## データモデル（主要テーブル）

`users` / `accounts` / `llm_configs` / `threads` / `emails` / `labels` / `email_labels` / `attachments` / `push_subscriptions`。
DDL は [migrations/](../migrations) と [詳細設計.md §2](../詳細設計.md) を参照。

- 本文・添付の実体は保持せず、`emails` にはメタデータ・プレビュー・AI 要約のみ保存。全文は S3 互換ストレージ、または IMAP から UID 指定で遅延取得。
- パスワード/API キーは `BYTEA` に AES-256-GCM（`nonce || ciphertext`）で保存。

## フロントエンド

- レイアウトはリーディングペイン（案 B）、狭幅で単一カラムにレスポンシブ。
- テーマ（light/dark/system）とブランドキーカラーは CSS 変数で表現し、`localStorage`（`ui.theme` / `ui.brand` …）＋サーバー（`/me/preferences`）の二層で永続化。
- PWA は vite-plugin-pwa（Workbox）。Service Worker が静的アセットのプリキャッシュ、API の NetworkFirst、`push` / `notificationclick` を処理。

## セキュリティ

- 全リソースは `user_id` スコープでフィルタ。
- HTML メールは DOMPurify でサニタイズ、外部画像は既定でブロック。
- IMAP/SMTP は TLS/STARTTLS 前提。
- LLM の `base_url` をローカルに向ければメール本文を外部に出さずに解析可能。

詳細は [deployment.md](deployment.md) のセキュリティチェックリストも参照。
