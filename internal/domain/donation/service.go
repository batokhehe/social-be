package donation

import (
	"context"
)

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*Donation, error) {
	item := &Donation{
		DonaturID:          req.DonaturID,
		DonaturGroupID:     req.DonaturGroupID,
		AreaID:             req.AreaID,
		DonationCategoryID: req.DonationCategoryID,
		Currency:           req.Currency,
		Amount:             req.Amount,
		OtherItems:         req.OtherItems,
	}
	return s.Repo.Create(ctx, item)
}

func (s *Service) GetAll(ctx context.Context) ([]Donation, error) {
	return s.Repo.GetAll(ctx)
}

func (s *Service) GetByID(ctx context.Context, id int) (*Donation, error) {
	return s.Repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int, req UpdateRequest) (*Donation, error) {
	item := &Donation{
		ID:                 id,
		DonaturID:          req.DonaturID,
		DonaturGroupID:     req.DonaturGroupID,
		AreaID:             req.AreaID,
		DonationCategoryID: req.DonationCategoryID,
		Currency:           req.Currency,
		Amount:             req.Amount,
		OtherItems:         req.OtherItems,
	}
	return s.Repo.Update(ctx, item)
}

func (s *Service) Delete(ctx context.Context, id int) error {
	return s.Repo.Delete(ctx, id)
}
