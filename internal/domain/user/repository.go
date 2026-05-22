package user

import (
	"context"
	"fmt"
	"social-be/internal/pkg/pagination"
	"social-be/internal/pkg/query"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, req CreateRequest, passwordHash string) error
	GetAll(ctx context.Context) ([]User, error)
	GetPaginated(ctx context.Context, page pagination.Query, filters query.Filters) ([]User, int64, error)
	GetByEmail(ctx context.Context, email string) (*User, string, int, error)
	GetByID(ctx context.Context, id int) (*User, error)
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

type userModel struct {
	ID           int        `gorm:"column:id"`
	Name         string     `gorm:"column:name"`
	Email        string     `gorm:"column:email"`
	PasswordHash string     `gorm:"column:password_hash"`
	Role         int        `gorm:"column:role"`
	Status       int        `gorm:"column:status"`
	ProfilePhoto string     `gorm:"column:profile_photo"`
	CreatedAt    *time.Time `gorm:"column:created_at"`
	CreatedBy    *int       `gorm:"column:created_by"`
	UpdatedAt    *time.Time `gorm:"column:updated_at"`
	UpdatedBy    *int       `gorm:"column:updated_by"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
	DeletedBy    *int       `gorm:"column:deleted_by"`
}

func (userModel) TableName() string {
	return "users"
}

func toEntity(row userModel) User {
	return User{
		ID:           row.ID,
		Name:         row.Name,
		Email:        row.Email,
		Role:         row.Role,
		Status:       row.Status,
		ProfilePhoto: row.ProfilePhoto,
	}
}

func (r *GormRepository) Create(ctx context.Context, req CreateRequest, passwordHash string) error {
	role := RoleVolunteer
	if req.Role != nil {
		role = *req.Role
	}
	row := userModel{Name: req.Name, Email: req.Email, PasswordHash: passwordHash, Role: role, ProfilePhoto: req.ProfilePhoto}
	if err := r.DB.WithContext(ctx).Select("name", "email", "password_hash", "role", "profile_photo").Create(&row).Error; err != nil {
		r.Logger.WithError(err).Error("Create User failed")
		return fmt.Errorf("failed create user: %w", err)
	}

	return nil
}

func (r *GormRepository) GetAll(ctx context.Context) ([]User, error) {
	var rows []userModel
	if err := r.DB.WithContext(ctx).Select("id", "name", "email", "role", "status", "profile_photo").
		Where("deleted_at IS NULL").
		Order("id ASC").
		Find(&rows).Error; err != nil {
		r.Logger.WithError(err).Error("GetAll User failed")
		return nil, fmt.Errorf("failed get all user: %w", err)
	}

	users := make([]User, 0, len(rows))
	for _, row := range rows {
		users = append(users, toEntity(row))
	}

	return users, nil
}

func (r *GormRepository) GetPaginated(ctx context.Context, page pagination.Query, filters query.Filters) ([]User, int64, error) {
	var total int64
	baseQuery := r.DB.WithContext(ctx).Model(&userModel{}).Where("deleted_at IS NULL")
	baseQuery = query.ApplyFilters(baseQuery, filters)
	if err := baseQuery.Count(&total).Error; err != nil {
		r.Logger.WithError(err).Error("Count User failed")
		return nil, 0, fmt.Errorf("failed count user: %w", err)
	}

	var rows []userModel
	if err := baseQuery.
		Select("id", "name", "email", "role", "status", "profile_photo").
		Order("id ASC").
		Limit(page.Limit).
		Offset(page.Offset).
		Find(&rows).Error; err != nil {
		r.Logger.WithError(err).Error("GetPaginated User failed")
		return nil, 0, fmt.Errorf("failed get paginated user: %w", err)
	}

	users := make([]User, 0, len(rows))
	for _, row := range rows {
		users = append(users, toEntity(row))
	}

	return users, total, nil
}

func (r *GormRepository) GetByEmail(ctx context.Context, email string) (*User, string, int, error) {
	var row userModel
	err := r.DB.WithContext(ctx).Select("id", "name", "email", "password_hash", "role", "status", "profile_photo").
		Where("email = ? AND deleted_at IS NULL", email).
		Take(&row).Error
	if err != nil {
		return nil, "", 0, err
	}

	user := toEntity(row)
	return &user, row.PasswordHash, row.Role, nil
}

func (r *GormRepository) GetByID(ctx context.Context, id int) (*User, error) {
	var row userModel
	err := r.DB.WithContext(ctx).Select("id", "name", "email", "role", "status", "profile_photo").
		Where("id = ? AND deleted_at IS NULL", id).
		Take(&row).Error
	if err != nil {
		return nil, err
	}

	user := toEntity(row)
	return &user, nil
}
