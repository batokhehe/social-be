package expense

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

func (s *Service) Create(ctx context.Context, req CreateRequest, actorID int) (*Expense, error) {
	return s.Repo.Create(ctx, req, actorID)
}

func (s *Service) GetPaginated(ctx context.Context, page pagination.Query, filters query.Filters, sort ExpenseSort) ([]ExpenseListItem, pagination.Meta, error) {
	items, total, err := s.Repo.GetPaginated(ctx, page, filters, sort)
	if err != nil {
		return nil, pagination.Meta{}, err
	}
	return items, pagination.NewMeta(page.Page, page.Limit, total), nil
}

func (s *Service) GetByID(ctx context.Context, id int) (*Expense, error) {
	return s.Repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*Expense, error) {
	return s.Repo.Update(ctx, id, req, actorID)
}

func (s *Service) Delete(ctx context.Context, id int, actorID int) error {
	return s.Repo.SoftDelete(ctx, id, actorID)
}

// GenerateExpenseNumber previews the next expense number for the given date
// string. The authoritative number is allocated atomically during Create.
func (s *Service) GenerateExpenseNumber(ctx context.Context, expenseDate string) (string, error) {
	t, err := parseExpenseDate(expenseDate)
	if err != nil {
		return "", err
	}
	return s.Repo.PeekNextExpenseNumber(ctx, t)
}
