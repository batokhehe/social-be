ALTER TABLE volunteers
    ADD COLUMN IF NOT EXISTS ktp_address TEXT NULL,
    ADD COLUMN IF NOT EXISTS district VARCHAR(100) NULL,
    ADD COLUMN IF NOT EXISTS gender VARCHAR(10) NULL;

ALTER TABLE volunteers
    ADD CONSTRAINT chk_volunteers_gender
    CHECK (gender IS NULL OR gender IN ('male', 'female'));
