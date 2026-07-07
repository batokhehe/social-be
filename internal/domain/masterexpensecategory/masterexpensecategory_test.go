package masterexpensecategory

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"social-be/internal/middleware"
	"social-be/internal/pkg/pagination"
	"social-be/internal/pkg/query"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

type mockRepo struct {
	createFn func(ctx context.Context, req CreateRequest) (*MasterExpenseCategory, error)
	updateFn func(ctx context.Context, id int, req UpdateRequest) (*MasterExpenseCategory, error)
	deleteFn func(ctx context.Context, id int) error
	selectFn func(ctx context.Context) ([]SelectItem, error)
}

func (m *mockRepo) Create(ctx context.Context, req CreateRequest) (*MasterExpenseCategory, error) {
	return m.createFn(ctx, req)
}
func (m *mockRepo) GetPaginated(ctx context.Context, page pagination.Query, filters query.Filters, sort Sort) ([]MasterExpenseCategory, int64, error) {
	return nil, 0, nil
}
func (m *mockRepo) GetSelect(ctx context.Context) ([]SelectItem, error) { return m.selectFn(ctx) }
func (m *mockRepo) GetByID(ctx context.Context, id int) (*MasterExpenseCategory, error) {
	return nil, nil
}
func (m *mockRepo) Update(ctx context.Context, id int, req UpdateRequest) (*MasterExpenseCategory, error) {
	return m.updateFn(ctx, id, req)
}
func (m *mockRepo) SoftDelete(ctx context.Context, id int) error { return m.deleteFn(ctx, id) }

func newRouter(h *Handler) *gin.Engine {
	r := gin.New()
	r.Use(middleware.ErrorMiddleware())
	r.Use(func(c *gin.Context) { c.Set("request_id", "test"); c.Next() })
	r.POST("/cat", h.Create)
	r.PUT("/cat/:id", h.Update)
	r.DELETE("/cat/:id", h.Delete)
	r.GET("/cat/select", h.GetSelect)
	return r
}

func do(r *gin.Engine, method, target, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

const validBody = `{"code":"OPERASIONAL","name":"Operasional","active":true}`

// --- create ---

func TestCreate_Success(t *testing.T) {
	var captured CreateRequest
	repo := &mockRepo{createFn: func(ctx context.Context, req CreateRequest) (*MasterExpenseCategory, error) {
		captured = req
		return &MasterExpenseCategory{ID: 1, Code: req.Code, Name: req.Name, Active: *req.Active}, nil
	}}
	h := NewHandler(NewService(repo))
	w := do(newRouter(h), http.MethodPost, "/cat", validBody)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}
	if captured.Code != "OPERASIONAL" || captured.Active == nil || *captured.Active != true {
		t.Fatalf("request not passed through: %+v", captured)
	}
}

func TestCreate_MissingActive_400(t *testing.T) {
	repo := &mockRepo{createFn: func(ctx context.Context, req CreateRequest) (*MasterExpenseCategory, error) {
		t.Fatalf("repo must not be called on validation failure")
		return nil, nil
	}}
	h := NewHandler(NewService(repo))
	// active omitted -> *bool nil -> required fails
	w := do(newRouter(h), http.MethodPost, "/cat", `{"code":"X","name":"Y"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing active, got %d", w.Code)
	}
}

func TestCreate_CodeTooLong_400(t *testing.T) {
	repo := &mockRepo{createFn: func(ctx context.Context, req CreateRequest) (*MasterExpenseCategory, error) {
		t.Fatalf("repo must not be called on validation failure")
		return nil, nil
	}}
	h := NewHandler(NewService(repo))
	long := strings.Repeat("A", 31)
	w := do(newRouter(h), http.MethodPost, "/cat", `{"code":"`+long+`","name":"Y","active":true}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for code > 30, got %d", w.Code)
	}
}

func TestCreate_DuplicateCode_409(t *testing.T) {
	repo := &mockRepo{createFn: func(ctx context.Context, req CreateRequest) (*MasterExpenseCategory, error) {
		return nil, ErrCodeExists
	}}
	h := NewHandler(NewService(repo))
	w := do(newRouter(h), http.MethodPost, "/cat", validBody)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409 for duplicate code, got %d", w.Code)
	}
}

func TestCreate_DuplicateName_409(t *testing.T) {
	repo := &mockRepo{createFn: func(ctx context.Context, req CreateRequest) (*MasterExpenseCategory, error) {
		return nil, ErrNameExists
	}}
	h := NewHandler(NewService(repo))
	w := do(newRouter(h), http.MethodPost, "/cat", validBody)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409 for duplicate name, got %d", w.Code)
	}
}

// --- update ---

func TestUpdate_Success(t *testing.T) {
	repo := &mockRepo{updateFn: func(ctx context.Context, id int, req UpdateRequest) (*MasterExpenseCategory, error) {
		return &MasterExpenseCategory{ID: id, Code: req.Code, Name: req.Name, Active: *req.Active}, nil
	}}
	h := NewHandler(NewService(repo))
	w := do(newRouter(h), http.MethodPut, "/cat/5", validBody)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestUpdate_NotFound_404(t *testing.T) {
	repo := &mockRepo{updateFn: func(ctx context.Context, id int, req UpdateRequest) (*MasterExpenseCategory, error) {
		return nil, ErrCategoryNotFound
	}}
	h := NewHandler(NewService(repo))
	w := do(newRouter(h), http.MethodPut, "/cat/99", validBody)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- delete ---

func TestDelete_Success(t *testing.T) {
	var gotID int
	repo := &mockRepo{deleteFn: func(ctx context.Context, id int) error { gotID = id; return nil }}
	h := NewHandler(NewService(repo))
	w := do(newRouter(h), http.MethodDelete, "/cat/7", "")
	if w.Code != http.StatusOK || gotID != 7 {
		t.Fatalf("expected 200 & id 7, got %d id=%d", w.Code, gotID)
	}
}

func TestDelete_NotFound_404(t *testing.T) {
	repo := &mockRepo{deleteFn: func(ctx context.Context, id int) error { return ErrCategoryNotFound }}
	h := NewHandler(NewService(repo))
	w := do(newRouter(h), http.MethodDelete, "/cat/99", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- select ---

func TestGetSelect(t *testing.T) {
	repo := &mockRepo{selectFn: func(ctx context.Context) ([]SelectItem, error) {
		return []SelectItem{{ID: 1, Code: "OPERASIONAL", Name: "Operasional"}}, nil
	}}
	h := NewHandler(NewService(repo))
	w := do(newRouter(h), http.MethodGet, "/cat/select", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var env struct {
		Data []SelectItem `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(env.Data) != 1 || env.Data[0].Code != "OPERASIONAL" {
		t.Fatalf("unexpected select payload: %s", w.Body.String())
	}
}
