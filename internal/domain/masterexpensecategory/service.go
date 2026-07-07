package masterexpensecategory

import (
	"context"

	"social-be/internal/pkg/pagination"
	"social-be/internal/pkg/query"
)

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*MasterExpenseCategory, error) {
	return s.Repo.Create(ctx, req)
}

func (s *Service) GetPaginated(ctx context.Context, page pagination.Query, filters query.Filters, sort Sort) ([]MasterExpenseCategory, pagination.Meta, error) {
	items, total, err := s.Repo.GetPaginated(ctx, page, filters, sort)
	if err != nil {
		return nil, pagination.Meta{}, err
	}
	return items, pagination.NewMeta(page.Page, page.Limit, total), nil
}

func (s *Service) GetSelect(ctx context.Context) ([]SelectItem, error) {
	return s.Repo.GetSelect(ctx)
}

func (s *Service) GetByID(ctx context.Context, id int) (*MasterExpenseCategory, error) {
	return s.Repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int, req UpdateRequest) (*MasterExpenseCategory, error) {
	return s.Repo.Update(ctx, id, req)
}

func (s *Service) Delete(ctx context.Context, id int) error {
	return s.Repo.SoftDelete(ctx, id)
}
