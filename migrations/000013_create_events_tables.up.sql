CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    start_at TIMESTAMP NOT NULL,
    end_at TIMESTAMP NOT NULL,
    category_activity_id INT NOT NULL,
    activity_id INT NOT NULL,
    pic_user_id INT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_by INT NULL,
    updated_by INT NULL,
    deleted_by INT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    CONSTRAINT fk_events_category_activity FOREIGN KEY (category_activity_id) REFERENCES category_activities(id),
    CONSTRAINT fk_events_activity FOREIGN KEY (activity_id) REFERENCES master_activities(id),
    CONSTRAINT fk_events_pic_user FOREIGN KEY (pic_user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS event_attachments (
    id SERIAL PRIMARY KEY,
    event_id INT NOT NULL,
    file_path VARCHAR(255) NOT NULL,
    description VARCHAR(255) NULL,
    created_by INT NULL,
    updated_by INT NULL,
    deleted_by INT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    CONSTRAINT fk_event_attachments_event FOREIGN KEY (event_id) REFERENCES events(id)
);

CREATE TABLE IF NOT EXISTS event_attendances (
    id SERIAL PRIMARY KEY,
    event_id INT NOT NULL,
    volunteer_id INT NOT NULL,
    status VARCHAR(50) NOT NULL,
    checkin_at TIMESTAMP NULL,
    checkout_at TIMESTAMP NULL,
    checkin_photo VARCHAR(255) NULL,
    checkout_photo VARCHAR(255) NULL,
    created_by INT NULL,
    updated_by INT NULL,
    deleted_by INT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    CONSTRAINT fk_event_attendances_event FOREIGN KEY (event_id) REFERENCES events(id),
    CONSTRAINT fk_event_attendances_volunteer FOREIGN KEY (volunteer_id) REFERENCES volunteers(id)
);
