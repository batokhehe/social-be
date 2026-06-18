-- Soft delete for donations so import rollbacks preserve an audit trail.
ALTER TABLE donations
    ADD COLUMN IF NOT EXISTS deleted_by INT NULL,
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;

-- Restrict dedup lookups (and the duplicate preload) to live rows only, so a
-- rolled-back donation no longer blocks re-importing the same record.
DROP INDEX IF EXISTS idx_donations_dedup;
CREATE INDEX IF NOT EXISTS idx_donations_dedup
    ON donations(donatur_id, donation_category_id, amount)
    WHERE deleted_at IS NULL;
