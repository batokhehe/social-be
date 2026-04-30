package levelarea

type Service struct {
	Repo *Repository
}

func (s *Service) Create(req CreateRequest, actorID int) (*LevelArea, error) {
	return s.Repo.Create(req, actorID)
}

func (s *Service) GetAll() ([]LevelArea, error) {
	return s.Repo.GetAll()
}

func (s *Service) GetByID(id int) (*LevelArea, error) {
	return s.Repo.GetByID(id)
}

func (s *Service) Update(id int, req UpdateRequest, actorID int) (*LevelArea, error) {
	return s.Repo.Update(id, req, actorID)
}

func (s *Service) SoftDelete(id int, actorID int) error {
	return s.Repo.SoftDelete(id, actorID)
}
