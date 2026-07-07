CREATE TABLE IF NOT EXISTS expenses (
    id SERIAL PRIMARY KEY,
    expense_no VARCHAR(30) NOT NULL UNIQUE,
    expense_date TIMESTAMP NOT NULL,
    category_id INT NOT NULL,
    volunteer_id INT NOT NULL,
    amount NUMERIC(15, 2) NOT NULL CHECK (amount >= 0),
    description TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'paid', 'cancelled')),
    created_by INT NULL,
    updated_by INT NULL,
    deleted_by INT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    -- Financial data: RESTRICT (never cascade-delete expense records).
    CONSTRAINT fk_expenses_category   FOREIGN KEY (category_id)  REFERENCES master_expense_categories(id),
    CONSTRAINT fk_expenses_volunteer  FOREIGN KEY (volunteer_id) REFERENCES volunteers(id),
    CONSTRAINT fk_expenses_created_by FOREIGN KEY (created_by)   REFERENCES users(id),
    CONSTRAINT fk_expenses_updated_by FOREIGN KEY (updated_by)   REFERENCES users(id),
    CONSTRAINT fk_expenses_deleted_by FOREIGN KEY (deleted_by)   REFERENCES users(id)
);

-- List filter/sort/join support (partial: only live rows).
CREATE INDEX IF NOT EXISTS idx_expenses_category_id  ON expenses(category_id)  WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_expenses_volunteer_id ON expenses(volunteer_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_expenses_status       ON expenses(status)       WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_expenses_expense_date ON expenses(expense_date) WHERE deleted_at IS NULL;
-- expense_no already has a unique btree index; the LIKE 'EXP-YYYYMM-%' running
-- number lookup uses it via prefix match, so no extra index is needed.
