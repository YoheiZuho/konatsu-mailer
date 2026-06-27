DROP INDEX IF EXISTS idx_email_bodies_search_tsv;
ALTER TABLE email_bodies DROP COLUMN IF EXISTS search_tsv;
