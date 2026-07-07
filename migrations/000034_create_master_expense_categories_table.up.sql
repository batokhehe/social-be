CREATE TABLE IF NOT EXISTS master_expense_categories (
    id SERIAL PRIMARY KEY,
    code VARCHAR(30) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);

-- Default seed categories (idempotent).
INSERT INTO master_expense_categories (code, name) VALUES
    ('OPERASIONAL', 'Operasional'),
    ('TRANSPORT',   'Transport'),
    ('KONSUMSI',    'Konsumsi'),
    ('ATK',         'ATK'),
    ('PERALATAN',   'Peralatan'),
    ('KESEHATAN',   'Kesehatan'),
    ('SOSIAL',      'Sosial'),
    ('LAINNYA',     'Lainnya')
ON CONFLICT (code) DO NOTHING;
