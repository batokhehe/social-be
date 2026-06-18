ALTER TABLE donations
    ALTER COLUMN currency DROP DEFAULT;

-- NOTE: re-adding NOT NULL will fail if imported rows with NULL group/area exist.
ALTER TABLE donations
    ALTER COLUMN area_id SET NOT NULL,
    ALTER COLUMN donatur_group_id SET NOT NULL;
