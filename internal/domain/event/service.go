package event

import (
	"context"
	"errors"
	"fmt"
	"time"

	"social-be/internal/domain/volunteer"
	"social-be/internal/pkg/pagination"

	"gorm.io/gorm"
)

type Service struct {
	Repo             Repository
	VolunteerService *volunteer.Service
}

func (s *Service) Create(ctx context.Context, req CreateRequest, actorID int) (*Event, error) {
	startAt, err := parseDateTime(req.StartAt)
	if err != nil {
		return nil, fmt.Errorf("invalid start_at: %w", err)
	}
	endAt, err := parseDateTime(req.EndAt)
	if err != nil {
		return nil, fmt.Errorf("invalid end_at: %w", err)
	}
	if endAt.Before(startAt) {
		return nil, fmt.Errorf("end_at must be after start_at")
	}
	return s.Repo.Create(ctx, req, startAt, endAt, actorID)
}

func (s *Service) GetAll(ctx context.Context, page pagination.Query) ([]Event, pagination.Meta, error) {
	items, total, err := s.Repo.GetPaginated(ctx, page)
	if err != nil {
		return nil, pagination.Meta{}, err
	}
	return items, pagination.NewMeta(page.Page, page.Limit, total), nil
}

func (s *Service) GetActiveEvents(
	ctx context.Context,
	userID int,
	page pagination.Query,
) ([]Event, pagination.Meta, error) {

	volunteerData, err := s.VolunteerService.GetByUserID(ctx, userID)

	volunteerID := 0

	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pagination.Meta{}, err
		}
	} else {
		volunteerID = volunteerData.ID
	}

	items, total, err := s.Repo.GetActiveEvents(
		ctx,
		volunteerID,
		time.Now(),
		page,
	)
	if err != nil {
		return nil, pagination.Meta{}, err
	}

	return items, pagination.NewMeta(page.Page, page.Limit, total), nil
}

func (s *Service) GetAppliedEvents(ctx context.Context, userID int, page pagination.Query) ([]Event, pagination.Meta, error) {
	volunteerData, err := s.VolunteerService.GetByUserID(ctx, userID)
	if err != nil {
		return nil, pagination.Meta{}, err
	}
	items, total, err := s.Repo.GetAppliedEventsByVolunteer(ctx, volunteerData.ID, page)
	if err != nil {
		return nil, pagination.Meta{}, err
	}
	return items, pagination.NewMeta(page.Page, page.Limit, total), nil
}

func (s *Service) GetCompletedEvents(ctx context.Context, userID int, page pagination.Query) ([]Event, pagination.Meta, error) {
	volunteerData, err := s.VolunteerService.GetByUserID(ctx, userID)
	if err != nil {
		return nil, pagination.Meta{}, err
	}
	items, total, err := s.Repo.GetCompletedEventsByVolunteer(ctx, volunteerData.ID, page)
	if err != nil {
		return nil, pagination.Meta{}, err
	}
	return items, pagination.NewMeta(page.Page, page.Limit, total), nil
}

func (s *Service) GetDetailEventsByVolunteer(ctx context.Context, userID int, page pagination.Query) ([]Event, pagination.Meta, error) {
	volunteerData, err := s.VolunteerService.GetByUserID(ctx, userID)
	if err != nil {
		return nil, pagination.Meta{}, err
	}
	items, total, err := s.Repo.GetDetailEventsByVolunteer(ctx, volunteerData.ID, page)
	if err != nil {
		return nil, pagination.Meta{}, err
	}
	return items, pagination.NewMeta(page.Page, page.Limit, total), nil
}

func (s *Service) GetByID(ctx context.Context, id int) (*Event, error) {
	return s.Repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*Event, error) {
	startAt, err := parseDateTime(req.StartAt)
	if err != nil {
		return nil, fmt.Errorf("invalid start_at: %w", err)
	}
	endAt, err := parseDateTime(req.EndAt)
	if err != nil {
		return nil, fmt.Errorf("invalid end_at: %w", err)
	}
	if endAt.Before(startAt) {
		return nil, fmt.Errorf("end_at must be after start_at")
	}
	return s.Repo.Update(ctx, id, req, startAt, endAt, actorID)
}

func (s *Service) SoftDelete(ctx context.Context, id int, actorID int) error {
	return s.Repo.SoftDelete(ctx, id, actorID)
}

func (s *Service) AddAttachment(ctx context.Context, eventID int, filePath string, description string, actorID int) (*EventAttachment, error) {
	return s.Repo.AddAttachment(ctx, eventID, filePath, description, actorID)
}

func (s *Service) GetAttachments(ctx context.Context, eventID int) ([]EventAttachment, error) {
	return s.Repo.GetEventAttachments(ctx, eventID)
}

func (s *Service) ApplyToEvent(ctx context.Context, eventID int, userID int, actorID int) (*EventAttendance, error) {
	volunteerData, err := s.VolunteerService.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.Repo.Apply(ctx, eventID, volunteerData.ID, actorID)
}

func (s *Service) CheckIn(ctx context.Context, eventID int, userID int, photoPath *string, actorID int) (*EventAttendance, error) {
	volunteerData, err := s.VolunteerService.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.Repo.CheckIn(ctx, eventID, volunteerData.ID, photoPath, actorID)
}

func (s *Service) CheckOut(ctx context.Context, eventID int, userID int, photoPath *string, actorID int) (*EventAttendance, error) {
	volunteerData, err := s.VolunteerService.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.Repo.CheckOut(ctx, eventID, volunteerData.ID, photoPath, actorID)
}

func parseDateTime(value string) (time.Time, error) {
	formats := []string{
		"02012006 15:04",
		"02-01-2006 15:04",
		"02/01/2006 15:04",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, value); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("must use ddmmyyyy hh:mm or dd-mm-yyyy hh:mm format")
}
