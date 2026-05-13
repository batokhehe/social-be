CREATE TABLE IF NOT EXISTS master_activities (
    id SERIAL PRIMARY KEY,
    category_activity_id INT NOT NULL,
    name VARCHAR(50) NOT NULL,
    target INT NOT NULL,
    description VARCHAR(200) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_by INT NULL,
    updated_by INT NULL,
    deleted_by INT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    CONSTRAINT fk_master_activities_category_activity
        FOREIGN KEY (category_activity_id)
        REFERENCES category_activities(id)
);

CREATE INDEX IF NOT EXISTS idx_master_activities_category_activity_id
    ON master_activities(category_activity_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_master_activities_status
    ON master_activities(status)
    WHERE deleted_at IS NULL;
