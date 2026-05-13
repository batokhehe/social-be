package auth

import (
	"context"

	"social-be/internal/domain/user"
	"social-be/internal/domain/volunteer"
	"social-be/internal/pkg/security"
)

type Service struct {
	UserService      *user.Service
	VolunteerService *volunteer.Service
}

func NewService(userService *user.Service, volunteerService *volunteer.Service) *Service {
	return &Service{
		UserService:      userService,
		VolunteerService: volunteerService,
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) error {
	return s.UserService.Register(ctx, user.CreateRequest{
		Name:         req.Name,
		Email:        req.Email,
		Password:     req.Password,
		ProfilePhoto: req.ProfilePhoto,
	})
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*TokenResponse, error) {
	accessToken, refreshToken, userID, role, err := s.UserService.LoginEx(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	response := &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	// get user data
	user, err := s.UserService.GetByID(ctx, userID)
	if err == nil && user != nil {
		response.User = user
	}

	// If role is volunteer (role=2), fetch volunteer data
	if role == 2 && s.VolunteerService != nil {
		volunteer, err := s.VolunteerService.GetByUserID(ctx, userID)
		if err == nil && volunteer != nil {
			response.Data = volunteer
		}
	}

	return response, nil
}

func (s *Service) Refresh(ctx context.Context, req RefreshRequest) (*TokenResponse, error) {
	claims, err := security.ParseToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	if claims["type"] != "refresh" {
		return nil, errInvalidTokenType
	}

	userID, err := security.GetIntClaim(claims, "user_id")
	if err != nil {
		return nil, err
	}

	user, err := s.UserService.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	newAccess, err := security.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{AccessToken: newAccess}, nil
}
