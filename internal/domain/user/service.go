package user

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"social-be/internal/pkg/cache"
	"social-be/internal/pkg/pagination"
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

func (s *Service) UploadPhotoToNAS(ctx context.Context, file *multipart.FileHeader) (string, error) {
	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Read file content
	fileContent, err := io.ReadAll(src)
	if err != nil {
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	// Prepare multipart form data for NAS API
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Add file field
	fw, err := w.CreateFormFile("file", file.Filename)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := fw.Write(fileContent); err != nil {
		return "", fmt.Errorf("failed to write file content: %w", err)
	}

	// Close the writer
	w.Close()

	// Get NAS API URL from environment
	nasAPIURL := os.Getenv("NAS_API_URL")
	if nasAPIURL == "" {
		nasAPIURL = "http://localhost:8081/api/store-image" // Default NAS API URL
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", nasAPIURL, &b)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call NAS API: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("NAS API returned status %d", resp.StatusCode)
	}

	// Parse response
	var nasResponse struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&nasResponse); err != nil {
		return "", fmt.Errorf("failed to parse NAS response: %w", err)
	}

	if nasResponse.URL == "" {
		return "", errors.New("NAS API did not return a valid URL")
	}

	return nasResponse.URL, nil
}

func (s *Service) GetPaginated(ctx context.Context, page pagination.Query) ([]User, pagination.Meta, error) {
	users, total, err := s.Repo.GetPaginated(ctx, page)
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
