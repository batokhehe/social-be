package volunteer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"social-be/internal/pkg/apperror"
	"social-be/internal/pkg/pagination"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newVolunteerTestContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Accept", "application/json")
	c.Request = c.Request.WithContext(context.Background())
	c.Set("request_id", "test-request-id")
	return c, w
}

type resetPasswordEnvelope struct {
	Success bool        `json:"success"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Detail  interface{} `json:"detail,omitempty"`
}

type mockVolunteerRepo struct {
	validateTokenFn func(ctx context.Context, token string) (*ResetPasswordValidation, error)
	resetPasswordFn func(ctx context.Context, token, password string) error
	createFn        func(ctx context.Context, req CreateRequest, actorID int) (*Volunteer, error)
	updateFn        func(ctx context.Context, id int, req UpdateRequest, actorID int) (*Volunteer, error)
}

func (m *mockVolunteerRepo) Create(ctx context.Context, req CreateRequest, actorID int) (*Volunteer, error) {
	if m.createFn != nil {
		return m.createFn(ctx, req, actorID)
	}
	return nil, nil
}

func (m *mockVolunteerRepo) GetPaginated(ctx context.Context, page pagination.Query) ([]Volunteer, int64, error) {
	return nil, 0, nil
}

func (m *mockVolunteerRepo) GetSelect(ctx context.Context) ([]Volunteer, int64, error) {
	return nil, 0, nil
}

func (m *mockVolunteerRepo) GetByID(ctx context.Context, id int) (*Volunteer, error) {
	return nil, nil
}

func (m *mockVolunteerRepo) GetByUserID(ctx context.Context, userID int) (*Volunteer, error) {
	return nil, nil
}

func (m *mockVolunteerRepo) Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*Volunteer, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, req, actorID)
	}
	return nil, nil
}

func (m *mockVolunteerRepo) SoftDelete(ctx context.Context, id int, actorID int) error {
	return nil
}

func (m *mockVolunteerRepo) ValidateResetToken(ctx context.Context, token string) (*ResetPasswordValidation, error) {
	return m.validateTokenFn(ctx, token)
}

func (m *mockVolunteerRepo) ResetPassword(ctx context.Context, token, password string) error {
	return m.resetPasswordFn(ctx, token, password)
}

func TestVerifyResetPassword_Success(t *testing.T) {
	repo := &mockVolunteerRepo{
		validateTokenFn: func(ctx context.Context, token string) (*ResetPasswordValidation, error) {
			if token != "valid-token" {
				t.Fatalf("unexpected token: %s", token)
			}
			return &ResetPasswordValidation{Email: "volunteer@example.com", Valid: true, ExpiresAt: time.Now().Add(time.Hour)}, nil
		},
	}

	h := &Handler{Service: &Service{Repo: repo}}
	c, w := newVolunteerTestContext(http.MethodPost, "/volunteers/reset-password/verify", `{"token":"valid-token"}`)

	h.VerifyResetPassword(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var got resetPasswordEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}
	if !got.Success || got.Code != apperror.CodeOK || got.Message != "success" {
		t.Fatalf("unexpected envelope: %#v", got)
	}
	data, ok := got.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %#v", got.Data)
	}
	if data["email"] != "volunteer@example.com" {
		t.Fatalf("unexpected email: %#v", data)
	}
	if data["valid"] != true {
		t.Fatalf("expected valid true, got %#v", data)
	}
}

func TestResetPassword_Success(t *testing.T) {
	repo := &mockVolunteerRepo{
		resetPasswordFn: func(ctx context.Context, token, password string) error {
			if token != "valid-token" {
				t.Fatalf("unexpected token: %s", token)
			}
			if password != "StrongPass123" {
				t.Fatalf("unexpected password: %s", password)
			}
			return nil
		},
	}

	h := &Handler{Service: &Service{Repo: repo}}
	c, w := newVolunteerTestContext(http.MethodPost, "/volunteers/reset-password", `{"token":"valid-token","password":"StrongPass123"}`)

	h.ResetPassword(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var got resetPasswordEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}
	if !got.Success || got.Code != apperror.CodeOK || got.Message != "success" {
		t.Fatalf("unexpected envelope: %#v", got)
	}
}
