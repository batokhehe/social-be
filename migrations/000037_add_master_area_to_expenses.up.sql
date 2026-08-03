-- Expense ownership moves from Volunteer to Master Area.
-- volunteer_id is KEPT as an optional PIC (nullable); its FK is left intact.

-- 1. Add the new ownership column (nullable for backfill).
ALTER TABLE expenses
    ADD COLUMN IF NOT EXISTS master_area_id INT NULL;

-- 2. Backfill from the PIC volunteer's area. expenses.volunteer_id is currently
--    NOT NULL with an FK, and volunteers.master_area_id is NOT NULL, so every
--    existing row resolves deterministically.
UPDATE expenses e
SET master_area_id = v.master_area_id
FROM volunteers v
WHERE v.id = e.volunteer_id
    AND e.master_area_id IS NULL;

-- 3. Verify the backfill covered every row; abort the migration otherwise.
DO $$
DECLARE
    missing BIGINT;
BEGIN
    SELECT COUNT(*) INTO missing FROM expenses WHERE master_area_id IS NULL;
    IF missing > 0 THEN
        RAISE EXCEPTION 'expenses.master_area_id backfill incomplete: % row(s) still NULL', missing;
    END IF;
END $$;

-- 4. Ownership is mandatory.
ALTER TABLE expenses
    ALTER COLUMN master_area_id SET NOT NULL;

-- 5. Financial data: RESTRICT (never cascade-delete expense records).
ALTER TABLE expenses
    ADD CONSTRAINT fk_expenses_master_area
    FOREIGN KEY (master_area_id) REFERENCES master_areas(id);

-- 6. Ownership lookups + dashboard aggregation (area + month) on live rows only.
CREATE INDEX IF NOT EXISTS idx_expenses_master_area_id
    ON expenses (master_area_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_expenses_master_area_date
    ON expenses (master_area_id, expense_date)
    WHERE deleted_at IS NULL;

-- 7. Volunteer becomes an optional PIC. Column and FK are intentionally kept.
ALTER TABLE expenses
    ALTER COLUMN volunteer_id DROP NOT NULL;
