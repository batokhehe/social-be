ALTER TABLE volunteers
    DROP CONSTRAINT IF EXISTS chk_volunteers_gender;

ALTER TABLE volunteers
    DROP COLUMN IF EXISTS gender,
    DROP COLUMN IF EXISTS district,
    DROP COLUMN IF EXISTS ktp_address;
