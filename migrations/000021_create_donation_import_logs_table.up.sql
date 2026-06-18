CREATE TABLE IF NOT EXISTS donation_import_logs (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(36) NOT NULL UNIQUE,
    filename VARCHAR(255) NOT NULL,
    -- Header metadata extracted from the uploaded sheet (per-file, not per-row).
    komisariat_id VARCHAR(50) NULL,
    komisariat_name VARCHAR(150) NULL,
    period VARCHAR(20) NULL,
    uploaded_by INT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    total_rows INT NOT NULL DEFAULT 0,
    success_rows INT NOT NULL DEFAULT 0,
    failed_rows INT NOT NULL DEFAULT 0,
    skipped_rows INT NOT NULL DEFAULT 0,
    status VARCHAR(30) NOT NULL DEFAULT 'completed',
    CONSTRAINT fk_donation_import_logs_uploaded_by
        FOREIGN KEY (uploaded_by) REFERENCES users(id) ON DELETE SET NULL
);

-- batch_id is already indexed by its UNIQUE constraint; only add uploaded_at.
CREATE INDEX IF NOT EXISTS idx_donation_import_logs_uploaded_at ON donation_import_logs(uploaded_at);
