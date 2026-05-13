-- +migrate Up
ALTER TABLE users ADD COLUMN profile_photo VARCHAR(255);

-- +migrate Down
ALTER TABLE users DROP COLUMN profile_photo;