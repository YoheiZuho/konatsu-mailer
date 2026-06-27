-- Heuristic inbox category (primary / promotions / social / newsletters).
ALTER TABLE emails ADD COLUMN IF NOT EXISTS category VARCHAR(16) NOT NULL DEFAULT 'primary';
CREATE INDEX IF NOT EXISTS idx_emails_category ON emails(account_id, folder, category, date_sent DESC);

-- User-defined message filters (Thunderbird-style auto-classification rules).
CREATE TABLE IF NOT EXISTS filters (
  id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name        VARCHAR(128) NOT NULL,
  enabled     BOOLEAN NOT NULL DEFAULT true,
  match_type  VARCHAR(4) NOT NULL DEFAULT 'all',   -- all | any
  conditions  JSONB NOT NULL DEFAULT '[]',          -- [{field, op, value}]
  actions     JSONB NOT NULL DEFAULT '[]',          -- [{type, value}]
  position    INT NOT NULL DEFAULT 0,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_filters_user ON filters(user_id, position);
