package expense

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"social-be/internal/middleware"
	"social-be/internal/pkg/pagination"
	"social-be/internal/pkg/query"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

// --- mock repository ---

type mockRepo struct {
	createFn func(ctx context.Context, req CreateRequest, actorID int) (*Expense, error)
	updateFn func(ctx context.Context, id int, req UpdateRequest, actorID int) (*Expense, error)
	deleteFn func(ctx context.Context, id int, actorID int) error
}

func (m *mockRepo) Create(ctx context.Context, req CreateRequest, actorID int) (*Expense, error) {
	return m.createFn(ctx, req, actorID)
}
func (m *mockRepo) GetPaginated(ctx context.Context, page pagination.Query, filters query.Filters, sort ExpenseSort) ([]ExpenseListItem, int64, error) {
	return nil, 0, nil
}
func (m *mockRepo) GetByID(ctx context.Context, id int) (*Expense, error) { return nil, nil }
func (m *mockRepo) Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*Expense, error) {
	return m.updateFn(ctx, id, req, actorID)
}
func (m *mockRepo) SoftDelete(ctx context.Context, id int, actorID int) error {
	return m.deleteFn(ctx, id, actorID)
}

func newRouter(h *Handler) *gin.Engine {
	r := gin.New()
	r.Use(middleware.ErrorMiddleware())
	r.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Set("request_id", "test")
		c.Next()
	})
	r.POST("/expenses", h.Create)
	r.PUT("/expenses/:id", h.Update)
	r.DELETE("/expenses/:id", h.Delete)
	return r
}

const validBody = `{"expense_date":"2026-07-06","category_id":1,"volunteer_id":2,"amount":150000,"description":"beli ATK","status":"draft"}`

func do(r *gin.Engine, method, target, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

// --- expense number generation ---

func TestBuildExpenseNo(t *testing.T) {
	jul := time.Date(2026, time.July, 6, 0, 0, 0, 0, time.UTC)
	aug := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		t    time.Time
		seq  int
		want string
	}{
		{jul, 1, "EXP-202607-00001"},
		{jul, 2, "EXP-202607-00002"},
		{jul, 42, "EXP-202607-00042"},
		{aug, 1, "EXP-202608-00001"}, // resets in a new month
	}
	for _, tc := range cases {
		if got := buildExpenseNo(tc.t, tc.seq); got != tc.want {
			t.Fatalf("buildExpenseNo(%v,%d)=%s want %s", tc.t, tc.seq, got, tc.want)
		}
	}
}

func TestParseExpenseDate(t *testing.T) {
	for _, v := range []string{"2026-07-06", "2026-07-06 14:30:00", "2026-07-06T14:30:00Z"} {
		if _, err := parseExpenseDate(v); err != nil {
			t.Fatalf("parseExpenseDate(%q) unexpected error: %v", v, err)
		}
	}
	if _, err := parseExpenseDate("06/07/2026"); err == nil {
		t.Fatalf("expected error for invalid date format")
	}
}

// --- create ---

func TestCreate_Success(t *testing.T) {
	var captured CreateRequest
	repo := &mockRepo{createFn: func(ctx context.Context, req CreateRequest, actorID int) (*Expense, error) {
		captured = req
		return &Expense{ID: 5, ExpenseNo: "EXP-202607-00001", Status: req.Status}, nil
	}}
	h := &Handler{Service: NewService(repo)}
	w := do(newRouter(h), http.MethodPost, "/expenses", validBody)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}
	if captured.CategoryID != 1 || captured.VolunteerID != 2 || captured.Amount != 150000 {
		t.Fatalf("request not passed through: %+v", captured)
	}
}

func TestCreate_InvalidStatus_400(t *testing.T) {
	repo := &mockRepo{createFn: func(ctx context.Context, req CreateRequest, actorID int) (*Expense, error) {
		t.Fatalf("repo must not be called on validation failure")
		return nil, nil
	}}
	h := &Handler{Service: NewService(repo)}
	body := `{"expense_date":"2026-07-06","category_id":1,"volunteer_id":2,"amount":150000,"status":"approved"}`
	w := do(newRouter(h), http.MethodPost, "/expenses", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid status, got %d", w.Code)
	}
}

func TestCreate_ZeroAmount_400(t *testing.T) {
	repo := &mockRepo{createFn: func(ctx context.Context, req CreateRequest, actorID int) (*Expense, error) {
		t.Fatalf("repo must not be called on validation failure")
		return nil, nil
	}}
	h := &Handler{Service: NewService(repo)}
	body := `{"expense_date":"2026-07-06","category_id":1,"volunteer_id":2,"amount":0,"status":"draft"}`
	w := do(newRouter(h), http.MethodPost, "/expenses", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for zero amount, got %d", w.Code)
	}
}

func TestCreate_CategoryNotFound_400(t *testing.T) {
	repo := &mockRepo{createFn: func(ctx context.Context, req CreateRequest, actorID int) (*Expense, error) {
		return nil, ErrCategoryNotFound
	}}
	h := &Handler{Service: NewService(repo)}
	w := do(newRouter(h), http.MethodPost, "/expenses", validBody)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing category, got %d", w.Code)
	}
}

func TestCreate_InactiveCategory_400(t *testing.T) {
	repo := &mockRepo{createFn: func(ctx context.Context, req CreateRequest, actorID int) (*Expense, error) {
		return nil, ErrCategoryInactive
	}}
	h := &Handler{Service: NewService(repo)}
	w := do(newRouter(h), http.MethodPost, "/expenses", validBody)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for inactive category, got %d", w.Code)
	}
}

func TestUpdate_InactiveCategory_400(t *testing.T) {
	repo := &mockRepo{updateFn: func(ctx context.Context, id int, req UpdateRequest, actorID int) (*Expense, error) {
		return nil, ErrCategoryInactive
	}}
	h := &Handler{Service: NewService(repo)}
	w := do(newRouter(h), http.MethodPut, "/expenses/5", validBody)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for inactive category on update, got %d", w.Code)
	}
}

// --- update ---

func TestUpdate_NotFound_404(t *testing.T) {
	repo := &mockRepo{updateFn: func(ctx context.Context, id int, req UpdateRequest, actorID int) (*Expense, error) {
		return nil, ErrExpenseNotFound
	}}
	h := &Handler{Service: NewService(repo)}
	w := do(newRouter(h), http.MethodPut, "/expenses/99", validBody)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestUpdate_Success(t *testing.T) {
	repo := &mockRepo{updateFn: func(ctx context.Context, id int, req UpdateRequest, actorID int) (*Expense, error) {
		return &Expense{ID: id, Status: req.Status}, nil
	}}
	h := &Handler{Service: NewService(repo)}
	w := do(newRouter(h), http.MethodPut, "/expenses/5", validBody)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}
}

// --- delete ---

func TestDelete_Success(t *testing.T) {
	var gotID, gotActor int
	repo := &mockRepo{deleteFn: func(ctx context.Context, id int, actorID int) error {
		gotID, gotActor = id, actorID
		return nil
	}}
	h := &Handler{Service: NewService(repo)}
	w := do(newRouter(h), http.MethodDelete, "/expenses/7", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if gotID != 7 || gotActor != 1 {
		t.Fatalf("delete got id=%d actor=%d, want 7/1", gotID, gotActor)
	}

	var env struct {
		Data map[string]interface{} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &env)
	if env.Data["message"] != "expense deleted" {
		t.Fatalf("unexpected delete body: %s", w.Body.String())
	}
}

func TestDelete_NotFound_404(t *testing.T) {
	repo := &mockRepo{deleteFn: func(ctx context.Context, id int, actorID int) error {
		return ErrExpenseNotFound
	}}
	h := &Handler{Service: NewService(repo)}
	w := do(newRouter(h), http.MethodDelete, "/expenses/99", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
