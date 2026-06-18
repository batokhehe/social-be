package speak

import (
	"context"
	"social-be/internal/pkg/pagination"
)

type Service struct {
	Repo Repository
}

func (s *Service) Create(ctx context.Context, req CreateRequest, actorID int) (*Speak, error) {
	return s.Repo.Create(ctx, req, actorID)
}

func (s *Service) GetAll(ctx context.Context, page pagination.Query) ([]Speak, pagination.Meta, error) {
	items, total, err := s.Repo.GetPaginated(ctx, page)
	if err != nil {
		return nil, pagination.Meta{}, err
	}

	return items, pagination.NewMeta(page.Page, page.Limit, total), nil
}

func (s *Service) GetAllAsReporter(ctx context.Context, page pagination.Query, actorID int) ([]Speak, pagination.Meta, error) {
	items, total, err := s.Repo.GetPaginatedAsReporter(ctx, page, actorID)
	if err != nil {
		return nil, pagination.Meta{}, err
	}

	return items, pagination.NewMeta(page.Page, page.Limit, total), nil
}

func (s *Service) GetAllAsRespondent(ctx context.Context, page pagination.Query, actorID int) ([]Speak, pagination.Meta, error) {
	items, total, err := s.Repo.GetPaginatedAsRespondent(ctx, page, actorID)
	if err != nil {
		return nil, pagination.Meta{}, err
	}

	return items, pagination.NewMeta(page.Page, page.Limit, total), nil
}

func (s *Service) GetByID(ctx context.Context, id int) (*Speak, error) {
	return s.Repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*Speak, error) {
	return s.Repo.Update(ctx, id, req, actorID)
}

func (s *Service) SoftDelete(ctx context.Context, id int, actorID int) error {
	return s.Repo.SoftDelete(ctx, id, actorID)
}

func (s *Service) Action(ctx context.Context, id int, req ActionRequest, actorID int) (*Speak, error) {
	return s.Repo.Action(ctx, id, req, actorID)
}

func (s *Service) AddAttachment(ctx context.Context, speakID int, filePath string, originalName string, actorID int, attachmentType int) (*SpeakAttachment, error) {
	return s.Repo.AddAttachment(ctx, speakID, filePath, originalName, actorID, attachmentType)
}

func (s *Service) GetAttachments(ctx context.Context, speakID int, attachmentType int) ([]SpeakAttachment, error) {
	return s.Repo.GetAttachments(ctx, speakID, attachmentType)
}

func (s *Service) SoftDeleteAttachment(ctx context.Context, speakID int, attachmentID int, actorID int) error {
	return s.Repo.SoftDeleteAttachment(ctx, speakID, attachmentID, actorID)
}
