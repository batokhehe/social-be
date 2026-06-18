-- Batch identity for Excel-imported donations: enables audit trail and rollback.
ALTER TABLE donations
    ADD COLUMN IF NOT EXISTS import_batch_id VARCHAR(36) NULL;

CREATE INDEX IF NOT EXISTS idx_donations_import_batch_id
    ON donations(import_batch_id);

-- Speeds up the duplicate pre-load query (lookup by donor used during import).
CREATE INDEX IF NOT EXISTS idx_donations_dedup
    ON donations(donatur_id, donation_category_id, amount);
