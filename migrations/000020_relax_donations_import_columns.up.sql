-- Excel-imported donations only carry donatur, category, amount and note.
-- Relax the columns that are not present in the import sheet so imported rows
-- can be created without a group/area. Existing API create still supplies them.
ALTER TABLE donations
    ALTER COLUMN donatur_group_id DROP NOT NULL,
    ALTER COLUMN area_id DROP NOT NULL;

ALTER TABLE donations
    ALTER COLUMN currency SET DEFAULT 'IDR';
