CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

CREATE TABLE users (
  id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  email       VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  display_name VARCHAR(255),
  theme       VARCHAR(8)  NOT NULL DEFAULT 'system',
  brand_color VARCHAR(16) NOT NULL DEFAULT '#ffd20a',
  prefs       JSONB NOT NULL DEFAULT '{}',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE accounts (
  id                 UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id            UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  email              VARCHAR(255) NOT NULL,
  imap_host          VARCHAR(255) NOT NULL,
  imap_port          INT NOT NULL DEFAULT 993,
  imap_use_tls       BOOLEAN NOT NULL DEFAULT true,
  smtp_host          VARCHAR(255) NOT NULL,
  smtp_port          INT NOT NULL DEFAULT 587,
  smtp_use_starttls  BOOLEAN NOT NULL DEFAULT true,
  auth_user          VARCHAR(255) NOT NULL,
  password_encrypted BYTEA NOT NULL,
  sync_state         JSONB NOT NULL DEFAULT '{}',
  is_active          BOOLEAN NOT NULL DEFAULT true,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE llm_configs (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name            VARCHAR(100) NOT NULL,
  base_url        VARCHAR(512) NOT NULL DEFAULT 'https://api.openai.com/v1',
  api_key_encrypted BYTEA,
  model           VARCHAR(128) NOT NULL,
  temperature     REAL NOT NULL DEFAULT 0.2,
  max_tokens      INT NOT NULL DEFAULT 512,
  supports_json_schema BOOLEAN NOT NULL DEFAULT true,
  request_timeout_ms   INT NOT NULL DEFAULT 30000,
  is_default      BOOLEAN NOT NULL DEFAULT false,
  is_active       BOOLEAN NOT NULL DEFAULT true,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX uq_llm_default ON llm_configs(user_id) WHERE is_default;

CREATE TABLE threads (
  id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  account_id  UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  thread_key  VARCHAR(998) NOT NULL,
  subject     TEXT,
  last_date   TIMESTAMPTZ,
  UNIQUE(account_id, thread_key)
);

CREATE TABLE emails (
  id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  account_id   UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  thread_id    UUID REFERENCES threads(id) ON DELETE SET NULL,
  folder       VARCHAR(255) NOT NULL DEFAULT 'INBOX',
  imap_uid     BIGINT NOT NULL,
  message_id   VARCHAR(998),
  in_reply_to  VARCHAR(998),
  subject      TEXT,
  sender_name  VARCHAR(255),
  sender_addr  VARCHAR(320) NOT NULL,
  recipients   JSONB NOT NULL DEFAULT '[]',
  body_preview TEXT,
  ai_summary   TEXT,
  ai_priority  SMALLINT,
  analysis_status VARCHAR(16) NOT NULL DEFAULT 'pending',
  has_attachment BOOLEAN NOT NULL DEFAULT false,
  date_sent    TIMESTAMPTZ NOT NULL,
  is_read      BOOLEAN NOT NULL DEFAULT false,
  is_starred   BOOLEAN NOT NULL DEFAULT false,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(account_id, folder, imap_uid)
);
CREATE INDEX idx_emails_list   ON emails(account_id, folder, date_sent DESC);
CREATE INDEX idx_emails_thread ON emails(thread_id, date_sent);
CREATE INDEX idx_emails_unread ON emails(account_id, is_read) WHERE is_read = false;
CREATE INDEX idx_emails_subj_trgm ON emails USING gin (subject gin_trgm_ops);

CREATE TABLE labels (
  id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  account_id  UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  name        VARCHAR(64) NOT NULL,
  color       VARCHAR(32) NOT NULL DEFAULT 'oklch(0.55 0.13 165)',
  is_system   BOOLEAN NOT NULL DEFAULT false,
  UNIQUE(account_id, name)
);

CREATE TABLE email_labels (
  email_id    UUID NOT NULL REFERENCES emails(id) ON DELETE CASCADE,
  label_id    UUID NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
  source      VARCHAR(8) NOT NULL DEFAULT 'ai',
  confidence  REAL,
  PRIMARY KEY(email_id, label_id)
);

CREATE TABLE attachments (
  id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  email_id    UUID NOT NULL REFERENCES emails(id) ON DELETE CASCADE,
  filename    VARCHAR(512),
  content_type VARCHAR(128),
  size_bytes  BIGINT,
  storage_key VARCHAR(512)
);

CREATE TABLE push_subscriptions (
  id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  endpoint    TEXT NOT NULL,
  p256dh      TEXT NOT NULL,
  auth        TEXT NOT NULL,
  user_agent  VARCHAR(255),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(user_id, endpoint)
);
