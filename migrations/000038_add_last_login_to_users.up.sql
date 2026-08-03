-- Track only the latest successful login (admin visibility). Nullable, additive;
-- no existing column changed, no new table.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS last_login_at         TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS last_login_ip         VARCHAR(45) NULL,
    ADD COLUMN IF NOT EXISTS last_login_user_agent TEXT NULL;
