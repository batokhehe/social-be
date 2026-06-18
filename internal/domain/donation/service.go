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
	otherItems := req.OtherItems
	item := &Donation{
		DonaturID:          req.DonaturID,
		DonaturGroupID:     &req.DonaturGroupID,
		AreaID:             req.AreaID,
		DonationCategoryID: req.DonationCategoryID,
		Type:               req.Type,
		Currency:           req.Currency,
		Amount:             req.Amount,
		OtherItems:         &otherItems,
	}
	applyDonationTypeRules(item)
	return s.Repo.Create(ctx, item)
}

func (s *Service) GetAll(ctx context.Context, donationType *int) ([]Donation, error) {
	return s.Repo.GetAll(ctx, donationType)
}

func (s *Service) GetByID(ctx context.Context, id int) (*Donation, error) {
	return s.Repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int, req UpdateRequest) (*Donation, error) {
	otherItems := req.OtherItems
	item := &Donation{
		ID:                 id,
		DonaturID:          req.DonaturID,
		DonaturGroupID:     &req.DonaturGroupID,
		AreaID:             req.AreaID,
		DonationCategoryID: req.DonationCategoryID,
		Type:               req.Type,
		Currency:           req.Currency,
		Amount:             req.Amount,
		OtherItems:         &otherItems,
	}
	applyDonationTypeRules(item)
	return s.Repo.Update(ctx, item)
}

func (s *Service) Delete(ctx context.Context, id int) error {
	return s.Repo.Delete(ctx, id)
}

// applyDonationTypeRules enforces the type-based field rules:
//   - money (0): no goods description -> other_items NULL
//   - goods (1): no monetary amount   -> amount 0
func applyDonationTypeRules(item *Donation) {
	switch item.Type {
	case DonationTypeMoney:
		item.OtherItems = nil
	case DonationTypeGoods:
		item.Amount = 0
	}
}
