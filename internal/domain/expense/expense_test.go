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

const validBody = `{"expense_date":"2026-07-06","master_area_id":3,"category_id":1,"volunteer_id":2,"amount":150000,"description":"beli ATK","status":"draft"}`

// bodyNoPIC omits volunteer_id entirely: an expense may exist without a PIC.
const bodyNoPIC = `{"expense_date":"2026-07-06","master_area_id":3,"category_id":1,"amount":150000,"description":"beli ATK","status":"draft"}`

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
	if captured.MasterAreaID != 3 || captured.CategoryID != 1 || captured.Amount != 150000 {
		t.Fatalf("request not passed through: %+v", captured)
	}
	if captured.VolunteerID == nil || *captured.VolunteerID != 2 {
		t.Fatalf("PIC volunteer not passed through: %+v", captured.VolunteerID)
	}
}

// An expense may exist without a PIC: volunteer_id omitted must still succeed.
func TestCreate_WithoutPIC_Success(t *testing.T) {
	var captured CreateRequest
	repo := &mockRepo{createFn: func(ctx context.Context, req CreateRequest, actorID int) (*Expense, error) {
		captured = req
		return &Expense{ID: 6, ExpenseNo: "EXP-202607-00002", MasterAreaID: req.MasterAreaID}, nil
	}}
	h := &Handler{Service: NewService(repo)}
	w := do(newRouter(h), http.MethodPost, "/expenses", bodyNoPIC)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 without PIC, got %d (%s)", w.Code, w.Body.String())
	}
	if captured.VolunteerID != nil {
		t.Fatalf("expected nil PIC, got %v", *captured.VolunteerID)
	}
	if captured.MasterAreaID != 3 {
		t.Fatalf("master_area_id = %d, want 3", captured.MasterAreaID)
	}
}

// bodyCrossAreaPIC owns the expense in area 3 while the PIC volunteer (id 2)
// belongs to a different area. This MUST be accepted.
const bodyCrossAreaPIC = `{"expense_date":"2026-07-06","master_area_id":3,"category_id":1,"volunteer_id":2,"amount":150000,"status":"draft"}`

// Business rule: ownership is master_area_id only; the PIC may belong to ANY
// Master Area. No validation may reject a PIC from a different area.
func TestCreate_CrossAreaPIC_Allowed(t *testing.T) {
	var captured CreateRequest
	repo := &mockRepo{createFn: func(ctx context.Context, req CreateRequest, actorID int) (*Expense, error) {
		captured = req
		// PIC volunteer 2 lives in area 99, the expense is owned by area 3.
		return &Expense{
			ID:           7,
			ExpenseNo:    "EXP-202607-00003",
			MasterAreaID: req.MasterAreaID,
			VolunteerID:  req.VolunteerID,
			MasterArea:   &MasterAreaInfo{ID: 3, Name: "Area Tiga"},
			Volunteer:    &VolunteerInfo{ID: 2, IndonesianName: "Budi", MasterAreaID: 99},
		}, nil
	}}
	h := &Handler{Service: NewService(repo)}
	w := do(newRouter(h), http.MethodPost, "/expenses", bodyCrossAreaPIC)

	if w.Code != http.StatusOK {
		t.Fatalf("cross-area PIC must be accepted, got %d (%s)", w.Code, w.Body.String())
	}
	if captured.MasterAreaID != 3 || captured.VolunteerID == nil || *captured.VolunteerID != 2 {
		t.Fatalf("both values must reach the repository unchanged: %+v", captured)
	}

	// The response must carry the two divergent areas side by side.
	var env struct {
		Data Expense `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Data.MasterAreaID != 3 {
		t.Fatalf("expense master_area_id = %d, want 3", env.Data.MasterAreaID)
	}
	if env.Data.Volunteer == nil || env.Data.Volunteer.MasterAreaID != 99 {
		t.Fatalf("PIC area must be independent of the expense area: %+v", env.Data.Volunteer)
	}
	if env.Data.MasterAreaID == env.Data.Volunteer.MasterAreaID {
		t.Fatalf("test is not exercising a cross-area PIC")
	}
}

// The same rule applies on update: reassigning to a PIC from another area is OK.
func TestUpdate_CrossAreaPIC_Allowed(t *testing.T) {
	var captured UpdateRequest
	repo := &mockRepo{updateFn: func(ctx context.Context, id int, req UpdateRequest, actorID int) (*Expense, error) {
		captured = req
		return &Expense{ID: id, MasterAreaID: req.MasterAreaID, VolunteerID: req.VolunteerID}, nil
	}}
	h := &Handler{Service: NewService(repo)}
	w := do(newRouter(h), http.MethodPut, "/expenses/7", bodyCrossAreaPIC)

	if w.Code != http.StatusOK {
		t.Fatalf("cross-area PIC must be accepted on update, got %d (%s)", w.Code, w.Body.String())
	}
	if captured.MasterAreaID != 3 || captured.VolunteerID == nil || *captured.VolunteerID != 2 {
		t.Fatalf("both values must reach the repository unchanged: %+v", captured)
	}
}

// master_area_id is the ownership field and is required.
func TestCreate_MissingMasterArea_400(t *testing.T) {
	repo := &mockRepo{createFn: func(ctx context.Context, req CreateRequest, actorID int) (*Expense, error) {
		t.Fatalf("repo must not be called on validation failure")
		return nil, nil
	}}
	h := &Handler{Service: NewService(repo)}
	body := `{"expense_date":"2026-07-06","category_id":1,"amount":150000,"status":"draft"}`
	w := do(newRouter(h), http.MethodPost, "/expenses", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing master_area_id, got %d", w.Code)
	}
}

func TestCreate_MasterAreaNotFound_400(t *testing.T) {
	repo := &mockRepo{createFn: func(ctx context.Context, req CreateRequest, actorID int) (*Expense, error) {
		return nil, ErrMasterAreaNotFound
	}}
	h := &Handler{Service: NewService(repo)}
	w := do(newRouter(h), http.MethodPost, "/expenses", validBody)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown master area, got %d", w.Code)
	}
}

func TestCreate_InvalidStatus_400(t *testing.T) {
	repo := &mockRepo{createFn: func(ctx context.Context, req CreateRequest, actorID int) (*Expense, error) {
		t.Fatalf("repo must not be called on validation failure")
		return nil, nil
	}}
	h := &Handler{Service: NewService(repo)}
	body := `{"expense_date":"2026-07-06","master_area_id":3,"category_id":1,"volunteer_id":2,"amount":150000,"status":"approved"}`
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
	body := `{"expense_date":"2026-07-06","master_area_id":3,"category_id":1,"volunteer_id":2,"amount":0,"status":"draft"}`
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
