CREATE TABLE IF NOT EXISTS master_donaturs (
    id SERIAL PRIMARY KEY,
    id_donatur VARCHAR(50) NOT NULL UNIQUE,
    telepon VARCHAR(20) NOT NULL,
    id_tzu_chi_app VARCHAR(50) NULL,
    id_vis_volunteer VARCHAR(50) NULL,
    id_group_donatur INT NULL,
    name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_by INT NULL,
    updated_by INT NULL,
    deleted_by INT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    CONSTRAINT fk_master_donaturs_donatur_group FOREIGN KEY (id_group_donatur) REFERENCES master_donatur_groups(id)
);
