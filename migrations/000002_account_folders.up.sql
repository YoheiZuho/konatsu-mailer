-- Cache the IMAP mailbox list (with special-use roles) discovered during sync,
-- so the UI can show the account's real folders (Inbox, Sent, Junk, Trash, ...).
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS folders JSONB NOT NULL DEFAULT '[]';
