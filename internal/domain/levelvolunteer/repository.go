package levelvolunteer

import (
	"context"
	"errors"
	"social-be/internal/pkg/pagination"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, req CreateRequest, actorID int) (*LevelVolunteer, error)
	GetPaginated(ctx context.Context, page pagination.Query) ([]LevelVolunteer, int64, error)
	GetByID(ctx context.Context, id int) (*LevelVolunteer, error)
	Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*LevelVolunteer, error)
	SoftDelete(ctx context.Context, id int, actorID int) error
}

type GormRepository struct {
	DB     *gorm.DB
	Logger *logrus.Logger
}

func NewGormRepository(db *gorm.DB, logger *logrus.Logger) Repository {
	return &GormRepository{DB: db, Logger: logger}
}

type levelVolunteerModel struct {
	ID          int        `gorm:"column:id"`
	Level       int        `gorm:"column:level"`
	Name        string     `gorm:"column:name"`
	Description string     `gorm:"column:description"`
	Status      string     `gorm:"column:status"`
	CreatedBy   *int       `gorm:"column:created_by"`
	UpdatedBy   *int       `gorm:"column:updated_by"`
	DeletedBy   *int       `gorm:"column:deleted_by"`
	CreatedAt   *time.Time `gorm:"column:created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
}

func (levelVolunteerModel) TableName() string {
	return "level_volunteers"
}

func toEntity(row levelVolunteerModel) LevelVolunteer {
	return LevelVolunteer{
		ID:          row.ID,
		Level:       row.Level,
		Name:        row.Name,
		Description: row.Description,
		Status:      row.Status,
		CreatedBy:   row.CreatedBy,
		UpdatedBy:   row.UpdatedBy,
		DeletedBy:   row.DeletedBy,
	}
}

func (r *GormRepository) Create(ctx context.Context, req CreateRequest, actorID int) (*LevelVolunteer, error) {
	r.Logger.Infof(
		"[LEVEL_VOLUNTEER][CREATE] start create actor=%d",
		actorID,
	)

	row := levelVolunteerModel{
		Name:        req.Name,
		Description: req.Description,
		Status:      defaultString(req.Status, "active"),
		CreatedBy:   &actorID,
		UpdatedBy:   &actorID,
	}

	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		insertLevel := 1

		if req.BelowLevelVolunteerID != nil {
			var anchor levelVolunteerModel

			if err := tx.
				Where("id = ? AND deleted_at IS NULL", *req.BelowLevelVolunteerID).
				First(&anchor).Error; err != nil {
				r.Logger.WithError(err).Error(
					"[LEVEL_VOLUNTEER][CREATE] anchor not found",
				)

				return err
			}

			insertLevel = anchor.Level + 1
		} else {
			var maxLevel int

			tx.Model(&levelVolunteerModel{}).
				Select("COALESCE(MAX(level),0)").
				Where("deleted_at IS NULL").
				Scan(&maxLevel)

			insertLevel = maxLevel + 1
		}

		r.Logger.Infof(
			"[LEVEL_VOLUNTEER][CREATE] insert level=%d",
			insertLevel,
		)

		if err := tx.Exec(`
			UPDATE level_volunteers
			SET level = level + 1
			WHERE deleted_at IS NULL
			AND level >= ?
		`, insertLevel).Error; err != nil {
			r.Logger.WithError(err).Error(
				"[LEVEL_VOLUNTEER][CREATE] failed shift levels",
			)

			return err
		}

		row.Level = insertLevel

		if err := tx.Create(&row).Error; err != nil {
			r.Logger.WithError(err).Error(
				"[LEVEL_VOLUNTEER][CREATE] failed insert database",
			)

			return err
		}

		r.Logger.Infof(
			"[LEVEL_VOLUNTEER][CREATE] success create id=%d level=%d",
			row.ID,
			row.Level,
		)

		return nil
	})

	if err != nil {
		r.Logger.WithError(err).Error("Create LevelVolunteer failed")
		return nil, err
	}

	out := toEntity(row)
	return &out, nil
}

func (r *GormRepository) GetPaginated(ctx context.Context, page pagination.Query) ([]LevelVolunteer, int64, error) {
	r.Logger.WithFields(logrus.Fields{
		"page":  page.Page,
		"limit": page.Limit,
	}).Info("[LEVEL_VOLUNTEER][GET_PAGINATED] fetching data")

	var total int64
	baseQuery := r.DB.WithContext(ctx).Model(&levelVolunteerModel{}).Where("deleted_at IS NULL")
	if err := baseQuery.Count(&total).Error; err != nil {
		r.Logger.WithError(err).Error("[LEVEL_VOLUNTEER][GET_PAGINATED] failed count data")
		return nil, 0, err
	}

	var rows []levelVolunteerModel
	if err := baseQuery.Order("level ASC").Limit(page.Limit).Offset(page.Offset).Find(&rows).Error; err != nil {
		r.Logger.WithError(err).Error("[LEVEL_VOLUNTEER][GET_PAGINATED] failed fetch data")
		return nil, 0, err
	}

	items := make([]LevelVolunteer, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEntity(row))
	}

	return items, total, nil
}

func (r *GormRepository) GetByID(ctx context.Context, id int) (*LevelVolunteer, error) {
	r.Logger.Infof(
		"[LEVEL_VOLUNTEER][GET_BY_ID] id=%d",
		id,
	)

	var row levelVolunteerModel

	err := r.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&row).Error
	if err != nil {
		r.Logger.WithError(err).Error(
			"[LEVEL_VOLUNTEER][GET_BY_ID] failed fetch data",
		)

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		return nil, err
	}

	out := toEntity(row)
	return &out, nil
}

func (r *GormRepository) Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*LevelVolunteer, error) {
	r.Logger.Infof(
		"[LEVEL_VOLUNTEER][UPDATE] start update id=%d actor=%d",
		id,
		actorID,
	)

	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current levelVolunteerModel

		if err := tx.
			Where("id = ? AND deleted_at IS NULL", id).
			First(&current).Error; err != nil {
			r.Logger.WithError(err).Error(
				"[LEVEL_VOLUNTEER][UPDATE] failed get current data",
			)

			return err
		}

		newLevel := current.Level

		if req.BelowLevelVolunteerID != nil {
			var anchor levelVolunteerModel

			if err := tx.
				Where("id = ? AND deleted_at IS NULL", *req.BelowLevelVolunteerID).
				First(&anchor).Error; err != nil {
				r.Logger.WithError(err).Error(
					"[LEVEL_VOLUNTEER][UPDATE] anchor not found",
				)

				return err
			}

			newLevel = anchor.Level + 1
		}

		r.Logger.Infof(
			"[LEVEL_VOLUNTEER][UPDATE] current=%d new=%d",
			current.Level,
			newLevel,
		)

		if newLevel != current.Level {
			if err := tx.Model(&levelVolunteerModel{}).
				Where("id = ?", current.ID).
				Update("level", 0).Error; err != nil {
				r.Logger.WithError(err).Error(
					"[LEVEL_VOLUNTEER][UPDATE] failed clear current level",
				)

				return err
			}

			if newLevel < current.Level {
				if err := tx.Exec(`
					UPDATE level_volunteers
					SET level = level + 1
					WHERE deleted_at IS NULL
					AND id <> ?
					AND level >= ?
					AND level < ?
				`, current.ID, newLevel, current.Level).Error; err != nil {
					r.Logger.WithError(err).Error(
						"[LEVEL_VOLUNTEER][UPDATE] failed shift up",
					)

					return err
				}
			}

			if newLevel > current.Level {
				if err := tx.Exec(`
					UPDATE level_volunteers
					SET level = level - 1
					WHERE deleted_at IS NULL
					AND id <> ?
					AND level <= ?
					AND level > ?
				`, current.ID, newLevel, current.Level).Error; err != nil {
					r.Logger.WithError(err).Error(
						"[LEVEL_VOLUNTEER][UPDATE] failed shift down",
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

		if err := tx.Model(&levelVolunteerModel{}).
			Where("id = ?", id).
			Updates(updates).Error; err != nil {
			r.Logger.WithError(err).Error(
				"[LEVEL_VOLUNTEER][UPDATE] failed update database",
			)

			return err
		}

		r.Logger.Infof(
			"[LEVEL_VOLUNTEER][UPDATE] success update id=%d",
			id,
		)

		return nil
	})

	if err != nil {
		r.Logger.WithError(err).Error("Update LevelVolunteer failed")
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *GormRepository) SoftDelete(ctx context.Context, id int, actorID int) error {
	r.Logger.Infof(
		"[LEVEL_VOLUNTEER][DELETE] start delete id=%d actor=%d",
		id,
		actorID,
	)

	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current levelVolunteerModel

		if err := tx.
			Where("id = ? AND deleted_at IS NULL", id).
			First(&current).Error; err != nil {
			r.Logger.WithError(err).Error(
				"[LEVEL_VOLUNTEER][DELETE] failed get current data",
			)

			return err
		}

		if err := tx.Model(&levelVolunteerModel{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"deleted_at": time.Now(),
				"deleted_by": actorID,
				"status":     "inactive",
			}).Error; err != nil {
			r.Logger.WithError(err).Error(
				"[LEVEL_VOLUNTEER][DELETE] failed soft delete",
			)

			return err
		}

		if err := tx.Exec(`
			UPDATE level_volunteers
			SET level = level - 1
			WHERE deleted_at IS NULL
			AND level > ?
		`, current.Level).Error; err != nil {
			r.Logger.WithError(err).Error(
				"[LEVEL_VOLUNTEER][DELETE] failed normalize levels",
			)

			return err
		}

		r.Logger.Infof(
			"[LEVEL_VOLUNTEER][DELETE] success delete id=%d",
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
