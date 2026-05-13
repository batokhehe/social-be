package levelvolunteer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"social-be/internal/pkg/apperror"
	"social-be/internal/pkg/pagination"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type mockRepository struct {
	createFn       func(ctx context.Context, req CreateRequest, actorID int) (*LevelVolunteer, error)
	getPaginatedFn func(ctx context.Context, page pagination.Query) ([]LevelVolunteer, int64, error)
	getByIDFn      func(ctx context.Context, id int) (*LevelVolunteer, error)
	updateFn       func(ctx context.Context, id int, req UpdateRequest, actorID int) (*LevelVolunteer, error)
	softDeleteFn   func(ctx context.Context, id int, actorID int) error
}

func (m *mockRepository) Create(ctx context.Context, req CreateRequest, actorID int) (*LevelVolunteer, error) {
	return m.createFn(ctx, req, actorID)
}

func (m *mockRepository) GetPaginated(ctx context.Context, page pagination.Query) ([]LevelVolunteer, int64, error) {
	return m.getPaginatedFn(ctx, page)
}

func (m *mockRepository) GetByID(ctx context.Context, id int) (*LevelVolunteer, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockRepository) Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*LevelVolunteer, error) {
	return m.updateFn(ctx, id, req, actorID)
}

func (m *mockRepository) SoftDelete(ctx context.Context, id int, actorID int) error {
	return m.softDeleteFn(ctx, id, actorID)
}

func newTestContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Accept", "application/json")
	c.Request = c.Request.WithContext(context.Background())
	c.Set("request_id", "test-request-id")
	return c, w
}

type envelope struct {
	Success bool        `json:"success"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
	Detail  interface{} `json:"detail,omitempty"`
}

func TestBindAndValidate_InvalidBody(t *testing.T) {
	c, _ := newTestContext(http.MethodPost, "/level-volunteers", "{invalid-json")
	var req CreateRequest
	if bindAndValidate(c, &req) {
		t.Fatal("expected bindAndValidate to return false for invalid JSON")
	}
	if !c.IsAborted() {
		t.Fatal("expected context to be aborted when bind fails")
	}
}

func TestActorID_NoUser(t *testing.T) {
	c, _ := newTestContext(http.MethodGet, "/level-volunteers", "")
	_, ok := actorID(c)
	if ok {
		t.Fatal("expected actorID to return false when user_id is missing")
	}
	if !c.IsAborted() {
		t.Fatal("expected context to be aborted when actor is missing")
	}
}

func TestCreate_Success(t *testing.T) {
	expected := &LevelVolunteer{ID: 1, Name: "Junior", Description: "Junior volunteer", Status: "active"}
	mockRepo := &mockRepository{
		createFn: func(ctx context.Context, req CreateRequest, actorID int) (*LevelVolunteer, error) {
			if actorID != 42 {
				t.Fatalf("expected actorID 42, got %d", actorID)
			}
			if req.Name != "Junior" || req.Description != "Junior volunteer" {
				t.Fatalf("unexpected request values: %#v", req)
			}
			return expected, nil
		},
	}

	h := &Handler{Service: &Service{Repo: mockRepo}}
	c, w := newTestContext(http.MethodPost, "/level-volunteers", `{"name":"Junior","description":"Junior volunteer"}`)
	c.Set("user_id", 42)

	h.Create(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var got envelope
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}
	if !got.Success || got.Code != apperror.CodeOK || got.Message != "success" {
		t.Fatalf("unexpected response envelope: %#v", got)
	}
}

func TestGetByID_InvalidID(t *testing.T) {
	mockRepo := &mockRepository{
		getByIDFn: func(ctx context.Context, id int) (*LevelVolunteer, error) {
			t.Fatalf("GetByID should not be called for invalid ID")
			return nil, nil
		},
	}

	h := &Handler{Service: &Service{Repo: mockRepo}}
	c, w := newTestContext(http.MethodGet, "/level-volunteers/abc", "")
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	h.GetByID(c)

	if !c.IsAborted() {
		t.Fatal("expected context to be aborted for invalid ID")
	}
	if w.Body.Len() != 0 {
		t.Fatalf("expected no response body for aborted context, got %q", w.Body.String())
	}
}

func TestDelete_Success(t *testing.T) {
	mockRepo := &mockRepository{
		softDeleteFn: func(ctx context.Context, id int, actorID int) error {
			if id != 7 {
				t.Fatalf("expected delete id 7, got %d", id)
			}
			if actorID != 11 {
				t.Fatalf("expected actorID 11, got %d", actorID)
			}
			return nil
		},
	}

	h := &Handler{Service: &Service{Repo: mockRepo}}
	c, w := newTestContext(http.MethodDelete, "/level-volunteers/7", "")
	c.Params = gin.Params{{Key: "id", Value: "7"}}
	c.Set("user_id", 11)

	h.Delete(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var got envelope
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}
	if got.Data == nil {
		t.Fatal("expected response body data to be present")
	}
}
