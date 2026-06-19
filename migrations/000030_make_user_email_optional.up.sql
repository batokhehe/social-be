-- Volunteers may now be created without an email. The volunteer create flow
-- still inserts a users row (login by vis_id), storing an empty email ('') when
-- none is provided. The original column-level UNIQUE constraint would reject a
-- second emailless user, so it is replaced with a partial unique index that
-- enforces uniqueness only for non-empty emails. The column stays NOT NULL
-- (empty string, never NULL), so no application model changes are required.
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;

CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique
    ON users (email)
    WHERE email <> '';
