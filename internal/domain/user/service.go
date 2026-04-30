package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"social-be/internal/pkg/cache"
	"social-be/internal/pkg/security"
	"time"
)

type Service struct {
	Repo *Repository
}

func (s *Service) Register(name, email, password string) error {
	hash, err := security.HashPassword(password)
	if err != nil {
		return err
	}
	cache.RDB.Del(cache.Ctx, "users:all")
	return s.Repo.Create(name, email, hash)
}

func (s *Service) Login(email, password string) (string, string, error) {
	user, hash, role, err := s.Repo.GetByEmail(email)
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

func (s *Service) GetUsers() ([]User, error) {
	val, err := cache.RDB.Get(cache.Ctx, "user:list").Result()
	if err == nil {
		var users []User
		if err := json.Unmarshal([]byte(val), &users); err == nil {
			return users, nil
		}
	}

	users, err := s.Repo.GetAll()
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(users)
	cache.RDB.Set(cache.Ctx, "user:list", data, time.Minute)

	return users, nil
}

func (s *Service) GetByID(id int) (*User, error) {
	key := fmt.Sprintf("user:%d", id)

	// 🔹 cek cache
	val, err := cache.RDB.Get(cache.Ctx, key).Result()
	if err == nil {
		var user User
		if err := json.Unmarshal([]byte(val), &user); err == nil {
			return &user, nil
		}
	}

	// 🔹 ambil dari DB
	user, err := s.Repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 🔹 simpan ke cache
	data, _ := json.Marshal(user)
	cache.RDB.Set(cache.Ctx, key, data, time.Minute)

	return user, nil
}
