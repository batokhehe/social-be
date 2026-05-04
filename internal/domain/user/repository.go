package user

import "gorm.io/gorm"

type Repository interface {
	Create(name, email, password string) error
	GetAll() ([]User, error)
	GetByEmail(email string) (*User, string, int, error)
	GetByID(id int) (*User, error)
}

type GormRepository struct {
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

func (r *GormRepository) Create(name, email, password string) error {
	row := userModel{Name: name, Email: email, PasswordHash: password}
	return r.DB.Select("name", "email", "password_hash").Create(&row).Error
}

func (r *GormRepository) GetAll() ([]User, error) {
	var rows []userModel
	if err := r.DB.Select("id", "name", "email").
		Where("deleted_at IS NULL").
		Order("id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	users := make([]User, 0, len(rows))
	for _, row := range rows {
		users = append(users, User{ID: row.ID, Name: row.Name, Email: row.Email})
	}

	return users, nil
}

func (r *GormRepository) GetByEmail(email string) (*User, string, int, error) {
	var row userModel
	err := r.DB.Select("id", "name", "email", "password_hash", "role").
		Where("email = ? AND deleted_at IS NULL", email).
		Take(&row).Error
	if err != nil {
		return nil, "", 0, err
	}

	return &User{ID: row.ID, Name: row.Name, Email: row.Email}, row.PasswordHash, row.Role, nil
}

func (r *GormRepository) GetByID(id int) (*User, error) {
	var row userModel
	err := r.DB.Select("id", "name", "email").
		Where("id = ? AND deleted_at IS NULL", id).
		Take(&row).Error
	if err != nil {
		return nil, err
	}

	return &User{ID: row.ID, Name: row.Name, Email: row.Email}, nil
}
