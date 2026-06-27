DROP INDEX IF EXISTS idx_emails_search_tsv;
ALTER TABLE emails DROP COLUMN IF EXISTS search_tsv;
