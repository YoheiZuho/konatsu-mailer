-- Make the cached full body searchable (full-text). Combined with emails.search_tsv
-- (subject/sender/preview), search now covers the full body of opened messages.
ALTER TABLE email_bodies ADD COLUMN IF NOT EXISTS search_tsv tsvector
  GENERATED ALWAYS AS (to_tsvector('simple', coalesce(text, ''))) STORED;

CREATE INDEX IF NOT EXISTS idx_email_bodies_search_tsv ON email_bodies USING gin (search_tsv);
