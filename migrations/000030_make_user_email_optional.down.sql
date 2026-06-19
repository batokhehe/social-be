-- Restore the full column-level UNIQUE constraint. NOTE: this will fail if more
-- than one user currently has an empty email ('') -- deduplicate those rows
-- before rolling back.
DROP INDEX IF EXISTS users_email_unique;

ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);
