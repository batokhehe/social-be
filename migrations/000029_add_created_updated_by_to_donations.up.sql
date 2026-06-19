-- Audit columns: who created / last updated a donation.
ALTER TABLE donations
    ADD COLUMN IF NOT EXISTS created_by INT NULL,
    ADD COLUMN IF NOT EXISTS updated_by INT NULL;
