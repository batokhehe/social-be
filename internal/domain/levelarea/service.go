package levelarea

import "context"

type Service struct {
	Repo Repository
}

func (s *Service) Create(ctx context.Context, req CreateRequest, actorID int) (*LevelArea, error) {
	return s.Repo.Create(ctx, req, actorID)
}

func (s *Service) GetAll(ctx context.Context) ([]LevelArea, error) {
	return s.Repo.GetAll(ctx)
}

func (s *Service) GetByID(ctx context.Context, id int) (*LevelArea, error) {
	return s.Repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*LevelArea, error) {
	return s.Repo.Update(ctx, id, req, actorID)
}

func (s *Service) SoftDelete(ctx context.Context, id int, actorID int) error {
	return s.Repo.SoftDelete(ctx, id, actorID)
}
