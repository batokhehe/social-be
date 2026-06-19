ALTER TABLE donations
    DROP COLUMN IF EXISTS updated_by,
    DROP COLUMN IF EXISTS created_by;
