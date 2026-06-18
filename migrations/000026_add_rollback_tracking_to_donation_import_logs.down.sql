ALTER TABLE donation_import_logs
    DROP CONSTRAINT IF EXISTS fk_donation_import_logs_rolled_back_by;

ALTER TABLE donation_import_logs
    DROP COLUMN IF EXISTS rolled_back_at,
    DROP COLUMN IF EXISTS rolled_back_by;
