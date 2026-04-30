CREATE TABLE IF NOT EXISTS level_areas (
    id SERIAL PRIMARY KEY,
    level INT NOT NULL,
    name VARCHAR(20) NOT NULL,
    description VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_by INT NULL,
    updated_by INT NULL,
    deleted_by INT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_level_areas_level ON level_areas(level) WHERE deleted_at IS NULL;
