# はじめに（セットアップ）

konatsu を Docker Compose で起動し、最初のログインまでを行う手順です。

## 1. 前提条件

- Docker / Docker Compose v2
- メールアカウントの IMAP/SMTP 情報（ホスト名・ポート・ユーザー名・パスワード）
- （任意）LLM 機能用の API キーまたはローカル LLM

## 2. 取得

```bash
git clone <repository-url> konatsu-mailer
cd konatsu-mailer
```

## 3. シークレットの生成

`.env.example` を `.env` にコピーし、最低限 2 つの秘密情報を生成して設定します。

```bash
cp .env.example .env
```

| 変数 | 生成コマンド | 説明 |
| :-- | :-- | :-- |
| `MASTER_ENC_KEY` | `openssl rand -hex 16` | メール／API パスワードを暗号化する AES-256 鍵。**ちょうど 32 バイト**である必要があります（`-hex 16` は 32 文字＝32 バイト）。 |
| `JWT_SECRET` | `openssl rand -base64 32` | セッショントークン（JWT）の署名鍵。 |

> ⚠️ `MASTER_ENC_KEY` は一度暗号化したデータの復号に必要です。**後から変更すると保存済みのパスワードが復号できなくなります。** 本番では安全に保管してください。

`.env` 例:

```dotenv
MASTER_ENC_KEY=3f9c1a...（32文字）
JWT_SECRET=Yk1n...（base64）
ALLOW_REGISTRATION=true
```

### （任意）Web Push の鍵

プッシュ通知を使う場合は VAPID 鍵ペアを生成して設定します。

```bash
npx web-push generate-vapid-keys
```

出力された Public/Private を `.env` の `VAPID_PUBLIC_KEY` / `VAPID_PRIVATE_KEY` に設定し、`VAPID_SUBJECT` を `mailto:あなたの連絡先` にします。未設定の場合、プッシュ通知は無効化されますが他の機能はそのまま動作します。

## 4. 起動

```bash
docker compose up --build
# または
make run
```

3 つのサービスが起動します。

| サービス | ポート | 説明 |
| :-- | :-- | :-- |
| `frontend` | `3000` | Web UI（nginx。`/api` をバックエンドへプロキシ） |
| `backend` | `8080` | REST API / WebSocket / 同期ワーカー |
| `db` | （内部のみ） | PostgreSQL。データは `./db_data` に永続化 |

DB マイグレーションはバックエンド起動時に自動適用されます。

## 5. 最初のログイン

1. ブラウザで **http://localhost:3000** を開きます。
2. 「新規登録」からユーザーを作成します。
   - 任意で「メールアカウントを設定」をオンにし、IMAP/SMTP 情報を入力すると、登録と同時にメールアカウントが追加されます。
3. ログイン後、右上 ⚙️ → 「アカウント」でメールアカウントを確認・追加できます。
4. 「AI 接続」で LLM を設定します（[llm-setup.md](llm-setup.md)）。

## 6. 停止・初期化

```bash
docker compose down            # 停止（データは残る）
docker compose down && rm -rf db_data   # データも削除して初期化
```

## トラブルシューティング

| 症状 | 対処 |
| :-- | :-- |
| `MASTER_ENC_KEY is required` で起動しない | `.env` に `MASTER_ENC_KEY` を設定。Compose 既定値でも起動はしますが本番では必ず変更。 |
| ログインできるがメールが表示されない | メールアカウントが未登録、または IMAP 接続情報が誤っている可能性。設定→アカウントを確認。 |
| 新規登録のリンクが出ない | `ALLOW_REGISTRATION=false` になっています（[configuration.md](configuration.md)）。 |
| AI 要約が出ない | LLM 接続が未設定、または接続テストに失敗。[llm-setup.md](llm-setup.md) を参照。 |

次は [使い方ガイド](usage.md) へ。
