-- Track who rolled a batch back and when (audit). These columns are mutable by
-- design (set during rollback) and are intentionally NOT covered by the
-- donation_import_logs immutability trigger.
ALTER TABLE donation_import_logs
    ADD COLUMN IF NOT EXISTS rolled_back_by INT NULL,
    ADD COLUMN IF NOT EXISTS rolled_back_at TIMESTAMPTZ NULL;

ALTER TABLE donation_import_logs
    ADD CONSTRAINT fk_donation_import_logs_rolled_back_by
        FOREIGN KEY (rolled_back_by) REFERENCES users(id) ON DELETE SET NULL;
