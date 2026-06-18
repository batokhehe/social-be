DROP INDEX IF EXISTS idx_donations_dedup;
DROP INDEX IF EXISTS idx_donations_import_batch_id;

ALTER TABLE donations
    DROP COLUMN IF EXISTS import_batch_id;
