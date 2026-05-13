CREATE TABLE IF NOT EXISTS master_areas (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    level_area_id VARCHAR(20) NOT NULL,
    description VARCHAR(200) NOT NULL,
    parent_id INT NULL,
    location VARCHAR(100) NOT NULL,
    created_by INT NULL,
    updated_by INT NULL,
    deleted_by INT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    CONSTRAINT fk_master_areas_parent
        FOREIGN KEY (parent_id)
        REFERENCES master_areas(id)
);

CREATE INDEX IF NOT EXISTS idx_master_areas_level_area_id
    ON master_areas(level_area_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_master_areas_parent_id
    ON master_areas(parent_id)
    WHERE deleted_at IS NULL;
