-- +migrate Up
ALTER TABLE users DROP COLUMN profile_photo;

-- +migrate Down
ALTER TABLE users ADD COLUMN profile_photo VARCHAR(255);