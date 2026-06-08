package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"social-be/internal/pkg/cache"
	"social-be/internal/pkg/pagination"
	"social-be/internal/pkg/query"
	"social-be/internal/pkg/security"
	"time"
)

const (
	userListCacheKey       = "user:list"
	userByIDCacheKeyFormat = "user:%d"
)

type Service struct {
	Repo Repository
}

func (s *Service) Register(ctx context.Context, req CreateRequest) error {
	hash, err := security.HashPassword(req.Password)
	if err != nil {
		return err
	}

	if err := s.Repo.Create(ctx, req, hash); err != nil {
		return err
	}

	cache.RDB.Del(ctx, userListCacheKey)
	return nil
}

func (s *Service) Login(ctx context.Context, email, password string) (string, string, error) {
	user, hash, role, err := s.Repo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", err
	}

	if !security.CheckPassword(hash, password) {
		return "", "", errors.New("invalid credentials")
	}

	accessToken, err := security.GenerateAccessToken(user.ID, user.Email, role)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := security.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// LoginEx returns accessToken, refreshToken, userID, role, error
func (s *Service) LoginEx(ctx context.Context, email, password string) (string, string, int, int, error) {
	user, hash, role, err := s.Repo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", 0, 0, err
	}

	if !security.CheckPassword(hash, password) {
		return "", "", 0, 0, errors.New("invalid credentials")
	}

	accessToken, err := security.GenerateAccessToken(user.ID, user.Email, role)
	if err != nil {
		return "", "", 0, 0, err
	}

	refreshToken, err := security.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", 0, 0, err
	}

	return accessToken, refreshToken, user.ID, role, nil
}

func (s *Service) GetAll(ctx context.Context) ([]User, error) {
	val, err := cache.RDB.Get(ctx, userListCacheKey).Result()
	if err == nil {
		var users []User
		if err := json.Unmarshal([]byte(val), &users); err == nil {
			return users, nil
		}
	}

	users, err := s.Repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(users)
	cache.RDB.Set(ctx, userListCacheKey, data, time.Minute)

	return users, nil
}

func (s *Service) GetPaginated(ctx context.Context, page pagination.Query, filters query.Filters) ([]User, pagination.Meta, error) {
	users, total, err := s.Repo.GetPaginated(ctx, page, filters)
	if err != nil {
		return nil, pagination.Meta{}, err
	}

	return users, pagination.NewMeta(page.Page, page.Limit, total), nil
}

func (s *Service) GetByID(ctx context.Context, id int) (*User, error) {
	key := fmt.Sprintf(userByIDCacheKeyFormat, id)

	val, err := cache.RDB.Get(ctx, key).Result()
	if err == nil {
		var user User
		if err := json.Unmarshal([]byte(val), &user); err == nil {
			return &user, nil
		}
	}

	user, err := s.Repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(user)
	cache.RDB.Set(ctx, key, data, time.Minute)

	return user, nil
}

func (s *Service) UpdateProfilePhoto(
	ctx context.Context,
	userID int,
	profilePhoto string,
) error {
	if err := s.Repo.UpdateProfilePhoto(
		ctx,
		userID,
		profilePhoto,
	); err != nil {
		return err
	}

	// hapus cache user
	cache.RDB.Del(ctx, fmt.Sprintf(userByIDCacheKeyFormat, userID))
	cache.RDB.Del(ctx, userListCacheKey)

	return nil
}
