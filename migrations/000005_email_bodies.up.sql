-- Cache of full message bodies, populated lazily on first open so the reading
-- pane reads from the DB instead of fetching from IMAP every time.
CREATE TABLE IF NOT EXISTS email_bodies (
  email_id    UUID PRIMARY KEY REFERENCES emails(id) ON DELETE CASCADE,
  text        TEXT NOT NULL DEFAULT '',
  html        TEXT NOT NULL DEFAULT '',
  attachments JSONB NOT NULL DEFAULT '[]',
  fetched_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
