-- Persisted per-row failures for each import batch (audit trail of bad rows).
CREATE TABLE IF NOT EXISTS donation_import_errors (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(36) NOT NULL,
    row_number INT NOT NULL,
    donator_id VARCHAR(50) NULL,
    category_name VARCHAR(150) NULL,
    error_message TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_donation_import_errors_batch
        FOREIGN KEY (batch_id) REFERENCES donation_import_logs(batch_id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_donation_import_errors_batch_id
    ON donation_import_errors(batch_id);
