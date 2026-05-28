package donation

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, donation *Donation) (*Donation, error)
	GetAll(ctx context.Context) ([]Donation, error)
	GetByID(ctx context.Context, id int) (*Donation, error)
	Update(ctx context.Context, donation *Donation) (*Donation, error)
	Delete(ctx context.Context, id int) error
}

type GormRepository struct {
	DB *gorm.DB
}

func NewGormRepository(db *gorm.DB) Repository {
	return &GormRepository{DB: db}
}

func (r *GormRepository) Create(ctx context.Context, donation *Donation) (*Donation, error) {
	if err := r.DB.WithContext(ctx).Create(donation).Error; err != nil {
		return nil, err
	}
	return donation, nil
}

func (r *GormRepository) GetAll(ctx context.Context) ([]Donation, error) {
	var donations []Donation
	if err := r.DB.WithContext(ctx).Order("id ASC").Find(&donations).Error; err != nil {
		return nil, err
	}
	return donations, nil
}

func (r *GormRepository) GetByID(ctx context.Context, id int) (*Donation, error) {
	var donation Donation
	if err := r.DB.WithContext(ctx).First(&donation, id).Error; err != nil {
		return nil, err
	}
	return &donation, nil
}

func (r *GormRepository) Update(ctx context.Context, donation *Donation) (*Donation, error) {
	if err := r.DB.WithContext(ctx).Save(donation).Error; err != nil {
		return nil, err
	}
	return donation, nil
}

func (r *GormRepository) Delete(ctx context.Context, id int) error {
	if err := r.DB.WithContext(ctx).Delete(&Donation{}, id).Error; err != nil {
		return err
	}
	return nil
}
