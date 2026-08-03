ALTER TABLE users
    DROP COLUMN IF EXISTS last_login_user_agent,
    DROP COLUMN IF EXISTS last_login_ip,
    DROP COLUMN IF EXISTS last_login_at;
