CREATE TABLE IF NOT EXISTS master_donatur_groups (
    id SERIAL PRIMARY KEY,
    id_group_donatur VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    pic_name VARCHAR(100) NOT NULL,
    pic_phone VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_by INT NULL,
    updated_by INT NULL,
    deleted_by INT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);
