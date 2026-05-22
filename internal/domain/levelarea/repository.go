package levelarea

import (
	"context"
	"errors"
	"social-be/internal/pkg/pagination"
	"social-be/internal/pkg/query"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, req CreateRequest, actorID int) (*LevelArea, error)
	GetAll(ctx context.Context) ([]LevelArea, error)
	GetPaginated(ctx context.Context, page pagination.Query, filters query.Filters) ([]LevelArea, int64, error)
	GetByID(ctx context.Context, id int) (*LevelArea, error)
	Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*LevelArea, error)
	SoftDelete(ctx context.Context, id int, actorID int) error
}

type GormRepository struct {
	DB     *gorm.DB
	Logger *logrus.Logger
}

func NewGormRepository(db *gorm.DB, logger *logrus.Logger) Repository {
	return &GormRepository{
		DB:     db,
		Logger: logger,
	}
}

type levelAreaModel struct {
	ID          int        `gorm:"column:id"`
	Level       int        `gorm:"column:level"`
	Name        string     `gorm:"column:name"`
	Description string     `gorm:"column:description"`
	Status      string     `gorm:"column:status"`
	CreatedBy   *int       `gorm:"column:created_by"`
	UpdatedBy   *int       `gorm:"column:updated_by"`
	DeletedBy   *int       `gorm:"column:deleted_by"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
}

func (levelAreaModel) TableName() string {
	return "level_areas"
}

func toEntity(item levelAreaModel) LevelArea {
	return LevelArea{
		ID:          item.ID,
		Level:       item.Level,
		Name:        item.Name,
		Description: item.Description,
		Status:      item.Status,
		CreatedBy:   item.CreatedBy,
		UpdatedBy:   item.UpdatedBy,
		DeletedBy:   item.DeletedBy,
	}
}

func (r *GormRepository) Create(ctx context.Context, req CreateRequest, actorID int) (*LevelArea, error) {
	r.Logger.Infof(
		"[LEVEL_AREA][CREATE] start create actor=%d",
		actorID,
	)

	item := levelAreaModel{
		Name:        req.Name,
		Description: req.Description,
		Status:      defaultString(req.Status, "active"),
		CreatedBy:   &actorID,
		UpdatedBy:   &actorID,
	}

	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		insertLevel := 1

		// kalau insert di bawah item tertentu
		if req.BelowLevelAreaID != nil {

			var anchor levelAreaModel

			if err := tx.
				Where("id = ? AND deleted_at IS NULL", *req.BelowLevelAreaID).
				First(&anchor).Error; err != nil {

				r.Logger.WithError(err).Error(
					"[LEVEL_AREA][CREATE] anchor not found",
				)

				return err
			}

			insertLevel = anchor.Level + 1
		} else {

			// append ke bawah
			var maxLevel int

			tx.Model(&levelAreaModel{}).
				Select("COALESCE(MAX(level),0)").
				Where("deleted_at IS NULL").
				Scan(&maxLevel)

			insertLevel = maxLevel + 1
		}

		r.Logger.Infof(
			"[LEVEL_AREA][CREATE] insert level=%d",
			insertLevel,
		)

		// geser level setelah posisi insert
		if err := tx.Exec(`
			UPDATE level_areas
			SET level = level + 1
			WHERE deleted_at IS NULL
			AND level >= ?
		`, insertLevel).Error; err != nil {

			r.Logger.WithError(err).Error(
				"[LEVEL_AREA][CREATE] failed shift levels",
			)

			return err
		}

		item.Level = insertLevel

		if err := tx.Create(&item).Error; err != nil {

			r.Logger.WithError(err).Error(
				"[LEVEL_AREA][CREATE] failed insert database",
			)

			return err
		}

		r.Logger.Infof(
			"[LEVEL_AREA][CREATE] success create id=%d level=%d",
			item.ID,
			item.Level,
		)

		return nil
	})

	if err != nil {
		r.Logger.WithError(err).Error("Create LevelArea failed")
		return nil, err
	}

	out := toEntity(item)

	return &out, nil
}

func (r *GormRepository) GetAll(ctx context.Context) ([]LevelArea, error) {
	r.Logger.Info("[LEVEL_AREA][GET_ALL] fetching all data")

	var rows []levelAreaModel

	err := r.DB.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("level ASC").
		Find(&rows).Error

	if err != nil {

		r.Logger.WithError(err).Error(
			"[LEVEL_AREA][GET_ALL] failed fetch data",
		)

		return nil, err
	}

	items := make([]LevelArea, 0, len(rows))

	for _, row := range rows {
		items = append(items, toEntity(row))
	}

	r.Logger.Infof(
		"[LEVEL_AREA][GET_ALL] total=%d",
		len(items),
	)

	return items, nil
}

func (r *GormRepository) GetPaginated(ctx context.Context, page pagination.Query, filters query.Filters) ([]LevelArea, int64, error) {
	r.Logger.WithFields(logrus.Fields{
		"page":  page.Page,
		"limit": page.Limit,
	}).Info("[LEVEL_AREA][GET_PAGINATED] fetching data")

	var total int64
	baseQuery := r.DB.WithContext(ctx).Model(&levelAreaModel{}).Where("deleted_at IS NULL")
	baseQuery = query.ApplyFilters(baseQuery, filters)

	if err := baseQuery.Count(&total).Error; err != nil {
		r.Logger.WithError(err).Error("[LEVEL_AREA][GET_PAGINATED] failed count data")
		return nil, 0, err
	}

	var rows []levelAreaModel
	if err := baseQuery.
		Order("level ASC").
		Limit(page.Limit).
		Offset(page.Offset).
		Find(&rows).Error; err != nil {
		r.Logger.WithError(err).Error("[LEVEL_AREA][GET_PAGINATED] failed fetch data")
		return nil, 0, err
	}

	items := make([]LevelArea, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEntity(row))
	}

	return items, total, nil
}

func (r *GormRepository) GetByID(ctx context.Context, id int) (*LevelArea, error) {
	r.Logger.Infof(
		"[LEVEL_AREA][GET_BY_ID] id=%d",
		id,
	)

	var item levelAreaModel

	err := r.DB.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&item).Error

	if err != nil {

		r.Logger.WithError(err).Error(
			"[LEVEL_AREA][GET_BY_ID] failed fetch data",
		)

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		return nil, err
	}

	out := toEntity(item)

	return &out, nil
}

func (r *GormRepository) Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*LevelArea, error) {
	r.Logger.Infof(
		"[LEVEL_AREA][UPDATE] start update id=%d actor=%d",
		id,
		actorID,
	)

	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		var current levelAreaModel

		if err := tx.
			Where("id = ? AND deleted_at IS NULL", id).
			First(&current).Error; err != nil {

			r.Logger.WithError(err).Error(
				"[LEVEL_AREA][UPDATE] failed get current data",
			)

			return err
		}

		newLevel := current.Level

		// kalau pindah posisi
		if req.BelowLevelAreaID != nil {

			var anchor levelAreaModel

			if err := tx.
				Where("id = ? AND deleted_at IS NULL", *req.BelowLevelAreaID).
				First(&anchor).Error; err != nil {

				r.Logger.WithError(err).Error(
					"[LEVEL_AREA][UPDATE] anchor not found",
				)

				return err
			}

			newLevel = anchor.Level + 1
		}

		r.Logger.Infof(
			"[LEVEL_AREA][UPDATE] current=%d new=%d",
			current.Level,
			newLevel,
		)

		// kalau level berubah
		if newLevel != current.Level {

			// kosongin sementara level item sekarang
			if err := tx.Model(&levelAreaModel{}).
				Where("id = ?", current.ID).
				Update("level", 0).Error; err != nil {

				r.Logger.WithError(err).Error(
					"[LEVEL_AREA][UPDATE] failed clear current level",
				)

				return err
			}

			// kalau pindah ke atas
			if newLevel < current.Level {

				if err := tx.Exec(`
					UPDATE level_areas
					SET level = level + 1
					WHERE deleted_at IS NULL
					AND id <> ?
					AND level >= ?
					AND level < ?
				`, current.ID, newLevel, current.Level).Error; err != nil {

					r.Logger.WithError(err).Error(
						"[LEVEL_AREA][UPDATE] failed shift up",
					)

					return err
				}
			}

			// kalau pindah ke bawah
			if newLevel > current.Level {

				if err := tx.Exec(`
					UPDATE level_areas
					SET level = level - 1
					WHERE deleted_at IS NULL
					AND id <> ?
					AND level <= ?
					AND level > ?
				`, current.ID, newLevel, current.Level).Error; err != nil {

					r.Logger.WithError(err).Error(
						"[LEVEL_AREA][UPDATE] failed shift down",
					)

					return err
				}
			}
		}

		updates := map[string]any{
			"level":       newLevel,
			"name":        req.Name,
			"description": req.Description,
			"status":      defaultString(req.Status, current.Status),
			"updated_by":  actorID,
			"updated_at":  time.Now(),
		}

		if err := tx.Model(&levelAreaModel{}).
			Where("id = ?", id).
			Updates(updates).Error; err != nil {

			r.Logger.WithError(err).Error(
				"[LEVEL_AREA][UPDATE] failed update database",
			)

			return err
		}

		r.Logger.Infof(
			"[LEVEL_AREA][UPDATE] success update id=%d",
			id,
		)

		return nil
	})

	if err != nil {
		r.Logger.WithError(err).Error("Update LevelArea failed")
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *GormRepository) SoftDelete(ctx context.Context, id int, actorID int) error {
	r.Logger.Infof(
		"[LEVEL_AREA][DELETE] start delete id=%d actor=%d",
		id,
		actorID,
	)

	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		var current levelAreaModel

		if err := tx.
			Where("id = ? AND deleted_at IS NULL", id).
			First(&current).Error; err != nil {

			r.Logger.WithError(err).Error(
				"[LEVEL_AREA][DELETE] failed get current data",
			)

			return err
		}

		if err := tx.Model(&levelAreaModel{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"deleted_at": time.Now(),
				"deleted_by": actorID,
				"status":     "inactive",
			}).Error; err != nil {

			r.Logger.WithError(err).Error(
				"[LEVEL_AREA][DELETE] failed soft delete",
			)

			return err
		}

		// normalize level
		if err := tx.Exec(`
			UPDATE level_areas
			SET level = level - 1
			WHERE deleted_at IS NULL
			AND level > ?
		`, current.Level).Error; err != nil {

			r.Logger.WithError(err).Error(
				"[LEVEL_AREA][DELETE] failed normalize levels",
			)

			return err
		}

		r.Logger.Infof(
			"[LEVEL_AREA][DELETE] success delete id=%d",
			id,
		)

		return nil
	})
}

func defaultString(val, def string) string {
	if val == "" {
		return def
	}

	return val
}
