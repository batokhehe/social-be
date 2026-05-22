CREATE TABLE IF NOT EXISTS speaks (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    pic_user_id INT NOT NULL,
    category_id INT NULL,
    is_anonymous BOOLEAN NOT NULL DEFAULT FALSE,
    description TEXT NULL,
    status INT NOT NULL DEFAULT 0,
    progress_at TIMESTAMP NULL,
    progress_notes TEXT NULL,
    finish_at TIMESTAMP NULL,
    finish_notes TEXT NULL,
    created_by INT NULL,
    updated_by INT NULL,
    deleted_by INT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS speak_attachments (
    id SERIAL PRIMARY KEY,
    speak_id INT NOT NULL,
    file_path VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    created_by INT NULL,
    updated_by INT NULL,
    deleted_by INT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    CONSTRAINT fk_speak_attachments_speak FOREIGN KEY (speak_id) REFERENCES speaks(id)
);
