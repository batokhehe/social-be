CREATE TABLE IF NOT EXISTS volunteers (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL UNIQUE,
    tzuhi_app_id VARCHAR(9) NOT NULL UNIQUE,
    vis_id VARCHAR(50) NOT NULL,
    indonesian_name VARCHAR(100) NOT NULL,
    mandarin_name VARCHAR(100) NOT NULL DEFAULT '',
    birth_place VARCHAR(100) NOT NULL,
    birth_date DATE NOT NULL,
    master_area_id INT NOT NULL,
    level_volunteer_id INT NOT NULL,
    attribute_volunteer_id INT NOT NULL,
    religion_id INT NOT NULL,
    blood_type VARCHAR(5) NOT NULL,
    rhesus VARCHAR(20) NOT NULL,
    last_education VARCHAR(10) NOT NULL,
    marital_status VARCHAR(20) NOT NULL,
    profession VARCHAR(100) NOT NULL,
    field VARCHAR(100) NOT NULL,
    residential_address VARCHAR(255) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    home_phone VARCHAR(30) NOT NULL DEFAULT '',
    office_phone VARCHAR(30) NOT NULL DEFAULT '',
    phone VARCHAR(30) NOT NULL,
    email VARCHAR(100) NOT NULL,
    languages VARCHAR(255) NOT NULL,
    private_vehicle BOOLEAN NOT NULL DEFAULT FALSE,
    regular_donor BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_by INT NULL,
    updated_by INT NULL,
    deleted_by INT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    CONSTRAINT fk_volunteers_user
        FOREIGN KEY (user_id)
        REFERENCES users(id),
    CONSTRAINT fk_volunteers_master_area
        FOREIGN KEY (master_area_id)
        REFERENCES master_areas(id),
    CONSTRAINT fk_volunteers_level_volunteer
        FOREIGN KEY (level_volunteer_id)
        REFERENCES level_volunteers(id),
    CONSTRAINT fk_volunteers_attribute_volunteer
        FOREIGN KEY (attribute_volunteer_id)
        REFERENCES attribute_volunteers(id),
    CONSTRAINT fk_volunteers_religion
        FOREIGN KEY (religion_id)
        REFERENCES religions(id)
);

CREATE TABLE IF NOT EXISTS volunteer_activities (
    volunteer_id INT NOT NULL,
    master_activity_id INT NOT NULL,
    PRIMARY KEY (volunteer_id, master_activity_id),
    CONSTRAINT fk_volunteer_activities_registration
        FOREIGN KEY (volunteer_id)
        REFERENCES volunteers(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_volunteer_activities_master_activity
        FOREIGN KEY (master_activity_id)
        REFERENCES master_activities(id)
);

CREATE TABLE IF NOT EXISTS volunteer_free_times (
    volunteer_id INT NOT NULL,
    day_of_week VARCHAR(20) NOT NULL,
    period VARCHAR(20) NOT NULL,
    PRIMARY KEY (volunteer_id, day_of_week, period),
    CONSTRAINT fk_volunteer_free_times_registration
        FOREIGN KEY (volunteer_id)
        REFERENCES volunteers(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_volunteers_tzuhi_app_id
    ON volunteers(tzuhi_app_id);

CREATE INDEX IF NOT EXISTS idx_volunteers_email
    ON volunteers(email)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    token VARCHAR(128) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_password_reset_tokens_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_token
    ON password_reset_tokens(token);
