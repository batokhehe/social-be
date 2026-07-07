-- Optional leadership markers. Nullable with no default: NULL = "not assigned /
-- unknown". Managed directly by administrators (no CRUD API).
ALTER TABLE volunteers
    ADD COLUMN IF NOT EXISTS is_hu_ai_leader  BOOLEAN NULL,
    ADD COLUMN IF NOT EXISTS is_hu_ai_deputy  BOOLEAN NULL,
    ADD COLUMN IF NOT EXISTS is_xie_li_leader BOOLEAN NULL,
    ADD COLUMN IF NOT EXISTS is_xie_li_deputy BOOLEAN NULL;
