-- Donation type (0 = money, 1 = goods) and donation period (datetime).
ALTER TABLE donations
    ADD COLUMN IF NOT EXISTS type SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS period TIMESTAMPTZ NULL;
