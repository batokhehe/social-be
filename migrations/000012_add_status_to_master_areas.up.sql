ALTER TABLE master_areas
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'active';

UPDATE master_areas
    SET status = 'active'
    WHERE status IS NULL;
