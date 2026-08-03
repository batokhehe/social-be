-- Reverses 000037. NOTE: restoring volunteer_id NOT NULL fails if any expense
-- was created without a PIC while the new model was live -- assign a PIC to
-- those rows (or keep the column nullable) before rolling back.
ALTER TABLE expenses
    ALTER COLUMN volunteer_id SET NOT NULL;

DROP INDEX IF EXISTS idx_expenses_master_area_date;
DROP INDEX IF EXISTS idx_expenses_master_area_id;

ALTER TABLE expenses
    DROP CONSTRAINT IF EXISTS fk_expenses_master_area;

ALTER TABLE expenses
    DROP COLUMN IF EXISTS master_area_id;
