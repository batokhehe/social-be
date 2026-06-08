CREATE TABLE IF NOT EXISTS donations (
    id SERIAL PRIMARY KEY,
    donatur_id INT NOT NULL,
    donatur_group_id INT NOT NULL,
    area_id INT NOT NULL,
    donation_category_id INT NOT NULL,
    currency VARCHAR(10) NOT NULL,
    amount NUMERIC(15, 2) NOT NULL,
    other_items TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_donations_donatur
        FOREIGN KEY (donatur_id)
        REFERENCES master_donaturs(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_donations_donatur_group
        FOREIGN KEY (donatur_group_id)
        REFERENCES master_donatur_groups(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_donations_area
        FOREIGN KEY (area_id)
        REFERENCES master_areas(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_donations_category
        FOREIGN KEY (donation_category_id)
        REFERENCES master_donation_categories(id)
        ON DELETE CASCADE
);