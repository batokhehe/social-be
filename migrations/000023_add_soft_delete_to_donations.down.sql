DROP INDEX IF EXISTS idx_donations_dedup;
CREATE INDEX IF NOT EXISTS idx_donations_dedup
    ON donations(donatur_id, donation_category_id, amount);

ALTER TABLE donations
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS deleted_by;
