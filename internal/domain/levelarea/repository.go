package levelarea

import "database/sql"

type Repository struct {
	DB *sql.DB
}

func (r *Repository) Create(req CreateRequest, actorID int) (*LevelArea, error) {
	status := req.Status
	if status == "" {
		status = "active"
	}

	var out LevelArea
	err := r.DB.QueryRow(`
		INSERT INTO level_areas (level, tingkatan_area, description, status, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING id_level_area, level, tingkatan_area, description, status, created_by, updated_by, deleted_by
	`, req.Level, req.TingkatanArea, req.Description, status, actorID).Scan(
		&out.IDLevelArea,
		&out.Level,
		&out.TingkatanArea,
		&out.Description,
		&out.Status,
		&out.CreatedBy,
		&out.UpdatedBy,
		&out.DeletedBy,
	)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

func (r *Repository) GetAll() ([]LevelArea, error) {
	rows, err := r.DB.Query(`
		SELECT id_level_area, level, tingkatan_area, description, status, created_by, updated_by, deleted_by
		FROM level_areas
		WHERE deleted_at IS NULL
		ORDER BY level ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []LevelArea
	for rows.Next() {
		var item LevelArea
		if err := rows.Scan(&item.IDLevelArea, &item.Level, &item.TingkatanArea, &item.Description, &item.Status, &item.CreatedBy, &item.UpdatedBy, &item.DeletedBy); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) GetByID(id int) (*LevelArea, error) {
	var item LevelArea
	err := r.DB.QueryRow(`
		SELECT id_level_area, level, tingkatan_area, description, status, created_by, updated_by, deleted_by
		FROM level_areas
		WHERE id_level_area = $1 AND deleted_at IS NULL
	`, id).Scan(&item.IDLevelArea, &item.Level, &item.TingkatanArea, &item.Description, &item.Status, &item.CreatedBy, &item.UpdatedBy, &item.DeletedBy)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) Update(id int, req UpdateRequest, actorID int) (*LevelArea, error) {
	status := req.Status
	if status == "" {
		status = "active"
	}

	var out LevelArea
	err := r.DB.QueryRow(`
		UPDATE level_areas
		SET level = $1,
		    tingkatan_area = $2,
		    description = $3,
		    status = $4,
		    updated_by = $5,
		    updated_at = NOW()
		WHERE id_level_area = $6 AND deleted_at IS NULL
		RETURNING id_level_area, level, tingkatan_area, description, status, created_by, updated_by, deleted_by
	`, req.Level, req.TingkatanArea, req.Description, status, actorID, id).Scan(
		&out.IDLevelArea,
		&out.Level,
		&out.TingkatanArea,
		&out.Description,
		&out.Status,
		&out.CreatedBy,
		&out.UpdatedBy,
		&out.DeletedBy,
	)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *Repository) SoftDelete(id int, actorID int) error {
	_, err := r.DB.Exec(`
		UPDATE level_areas
		SET deleted_at = NOW(),
		    deleted_by = $2,
		    updated_by = $2,
		    updated_at = NOW(),
		    status = 'inactive'
		WHERE id_level_area = $1 AND deleted_at IS NULL
	`, id, actorID)
	return err
}
