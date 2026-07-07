-- Name uniqueness is enforced at the application layer against ALL rows
-- (including soft-deleted), so this lookup runs on every create/update. code is
-- already covered by its UNIQUE constraint; name was unindexed. Full (non-
-- partial) index because the check intentionally scans soft-deleted rows too.
CREATE INDEX IF NOT EXISTS idx_master_expense_categories_name
    ON master_expense_categories (name);
