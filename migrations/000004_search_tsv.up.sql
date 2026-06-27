-- Full-text search vector over the stored, searchable text (subject + sender +
-- body preview). Bodies themselves are not persisted (design §2.1), so search
-- covers the preview text. 'simple' config avoids language-specific stemming;
-- Japanese substring matching is handled by ILIKE/pg_trgm in the query.
ALTER TABLE emails ADD COLUMN IF NOT EXISTS search_tsv tsvector
  GENERATED ALWAYS AS (
    to_tsvector('simple',
      coalesce(subject, '') || ' ' ||
      coalesce(sender_name, '') || ' ' ||
      coalesce(sender_addr, '') || ' ' ||
      coalesce(body_preview, ''))
  ) STORED;

CREATE INDEX IF NOT EXISTS idx_emails_search_tsv ON emails USING gin (search_tsv);
