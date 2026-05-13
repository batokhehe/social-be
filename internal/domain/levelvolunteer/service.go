package levelvolunteer

import (
	"context"
	"social-be/internal/pkg/pagination"
)

type Service struct {
	Repo Repository
}

func (s *Service) Create(ctx context.Context, req CreateRequest, actorID int) (*LevelVolunteer, error) {
	return s.Repo.Create(ctx, req, actorID)
}

func (s *Service) GetPaginated(ctx context.Context, page pagination.Query) ([]LevelVolunteer, pagination.Meta, error) {
	items, total, err := s.Repo.GetPaginated(ctx, page)
	if err != nil {
		return nil, pagination.Meta{}, err
	}

	return items, pagination.NewMeta(page.Page, page.Limit, total), nil
}

func (s *Service) GetByID(ctx context.Context, id int) (*LevelVolunteer, error) {
	return s.Repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*LevelVolunteer, error) {
	return s.Repo.Update(ctx, id, req, actorID)
}

func (s *Service) SoftDelete(ctx context.Context, id int, actorID int) error {
	return s.Repo.SoftDelete(ctx, id, actorID)
}
