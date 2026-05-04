package user

import (
	"errors"

	"gorm.io/gorm"
)

type Repository struct {
	DB *gorm.DB
}

type userModel struct {
	ID           int     `gorm:"column:id"`
	Name         string  `gorm:"column:name"`
	Email        string  `gorm:"column:email"`
	PasswordHash string  `gorm:"column:password_hash"`
	Role         int     `gorm:"column:role"`
	DeletedAt    *string `gorm:"column:deleted_at"`
}

func (userModel) TableName() string {
	return "users"
}

func (r *Repository) Create(name, email, password string) error {
	payload := map[string]any{
		"name":          name,
		"email":         email,
		"password_hash": password,
	}

	return r.DB.Table("users").Create(payload).Error
}

func (r *Repository) GetAll() ([]User, error) {
	var rows []userModel
	if err := r.DB.Where("deleted_at IS NULL").Find(&rows).Error; err != nil {
		return nil, err
	}

	users := make([]User, 0, len(rows))
	for _, row := range rows {
		users = append(users, User{ID: row.ID, Name: row.Name, Email: row.Email})
	}

	return users, nil
}

func (r *Repository) GetByEmail(email string) (*User, string, int, error) {
	var row userModel
	err := r.DB.Where("email = ? AND deleted_at IS NULL", email).First(&row).Error
	if err != nil {
		return nil, "", 0, err
	}

	return &User{ID: row.ID, Name: row.Name, Email: row.Email}, row.PasswordHash, row.Role, nil
}

func (r *Repository) GetByID(id int) (*User, error) {
	var row userModel
	err := r.DB.Where("id = ? AND deleted_at IS NULL", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return &User{ID: row.ID, Name: row.Name, Email: row.Email}, nil
}
