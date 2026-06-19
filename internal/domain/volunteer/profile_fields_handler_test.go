package volunteer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"social-be/internal/middleware"

	"github.com/gin-gonic/gin"
)

// validCreateBody returns a fully valid CreateRequest JSON payload. Callers
// override individual fields (e.g. gender) to exercise specific validations.
func validCreateBody(gender string) string {
	return `{
		"vis_id":"V001",
		"indonesian_name":"Budi Santoso",
		"birth_place":"Jakarta",
		"birth_date":"1990-01-01",
		"gender":"` + gender + `",
		"master_area_id":1,
		"level_volunteer_id":1,
		"religion_id":1,
		"blood_type":"O",
		"rhesus":"positive",
		"last_education":"S1",
		"marital_status":"single",
		"profession":"Engineer",
		"field":"IT",
		"residential_address":"Jl. Mawar No. 1",
		"ktp_address":"Jl. KTP No. 1, RT 01 RW 02, Jakarta Pusat",
		"district":"Menteng",
		"postal_code":"12345",
		"phone":"08123456789",
		"email":"budi@example.com",
		"interested_activity_ids":[1],
		"languages":"Indonesian",
		"free_times":[{"day_of_week":"monday","period":"morning"}]
	}`
}

// newRouter wires the handler through the same ErrorMiddleware used in
// production so that AbortError-recorded errors are written to the response,
// and injects an authenticated actor.
func newRouter(h *Handler) *gin.Engine {
	r := gin.New()
	r.Use(middleware.ErrorMiddleware())
	r.Use(func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Set("request_id", "test-request-id")
		c.Next()
	})
	r.POST("/volunteers", h.Create)
	r.PUT("/volunteers/:id", h.Update)
	return r
}

func TestCreateVolunteer_WithNewProfileFields(t *testing.T) {
	var captured CreateRequest
	repo := &mockVolunteerRepo{
		createFn: func(ctx context.Context, req CreateRequest, actorID int) (*Volunteer, error) {
			captured = req
			return &Volunteer{
				ID:         10,
				Gender:     req.Gender,
				KTPAddress: req.KTPAddress,
				District:   req.District,
			}, nil
		},
	}
	h := &Handler{Service: &Service{Repo: repo}}
	r := newRouter(h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/volunteers", strings.NewReader(validCreateBody("male")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d (body=%s)", w.Code, w.Body.String())
	}

	// New fields must reach the repository.
	if captured.Gender != "male" {
		t.Fatalf("expected gender 'male' passed to repo, got %q", captured.Gender)
	}
	if captured.District != "Menteng" {
		t.Fatalf("expected district 'Menteng' passed to repo, got %q", captured.District)
	}
	if !strings.HasPrefix(captured.KTPAddress, "Jl. KTP") {
		t.Fatalf("unexpected ktp_address passed to repo: %q", captured.KTPAddress)
	}

	// New fields must be returned in the response.
	var got resetPasswordEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}
	data, ok := got.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %#v", got.Data)
	}
	if data["gender"] != "male" {
		t.Fatalf("expected gender 'male' in response, got %#v", data["gender"])
	}
	if data["district"] != "Menteng" {
		t.Fatalf("expected district 'Menteng' in response, got %#v", data["district"])
	}
}

func TestUpdateVolunteer_WithNewProfileFields(t *testing.T) {
	var captured UpdateRequest
	repo := &mockVolunteerRepo{
		updateFn: func(ctx context.Context, id int, req UpdateRequest, actorID int) (*Volunteer, error) {
			captured = req
			return &Volunteer{
				ID:         id,
				Gender:     req.Gender,
				KTPAddress: req.KTPAddress,
				District:   req.District,
			}, nil
		},
	}
	h := &Handler{Service: &Service{Repo: repo}}
	r := newRouter(h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/volunteers/10", strings.NewReader(validCreateBody("female")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d (body=%s)", w.Code, w.Body.String())
	}
	if captured.Gender != "female" {
		t.Fatalf("expected gender 'female' passed to repo, got %q", captured.Gender)
	}

	var got resetPasswordEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}
	data, ok := got.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %#v", got.Data)
	}
	if data["gender"] != "female" {
		t.Fatalf("expected gender 'female' in response, got %#v", data["gender"])
	}
}

func TestCreateVolunteer_EmailOptional(t *testing.T) {
	// A fully valid payload with no email field at all.
	const bodyNoEmail = `{
		"vis_id":"V002",
		"indonesian_name":"Siti",
		"birth_place":"Bandung",
		"birth_date":"1992-05-05",
		"gender":"female",
		"master_area_id":1,
		"level_volunteer_id":1,
		"religion_id":1,
		"blood_type":"A",
		"rhesus":"positive",
		"last_education":"S1",
		"marital_status":"single",
		"profession":"Teacher",
		"field":"Education",
		"residential_address":"Jl. Melati No. 2",
		"postal_code":"40123",
		"phone":"08129998887",
		"interested_activity_ids":[1],
		"languages":"Indonesian",
		"free_times":[{"day_of_week":"tuesday","period":"evening"}]
	}`

	var captured CreateRequest
	repo := &mockVolunteerRepo{
		createFn: func(ctx context.Context, req CreateRequest, actorID int) (*Volunteer, error) {
			captured = req
			return &Volunteer{ID: 11, Email: req.Email}, nil
		},
	}
	h := &Handler{Service: &Service{Repo: repo}}
	r := newRouter(h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/volunteers", strings.NewReader(bodyNoEmail))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for missing email, got %d (body=%s)", w.Code, w.Body.String())
	}
	if captured.Email != "" {
		t.Fatalf("expected empty email passed to repo, got %q", captured.Email)
	}
}

func TestCreateVolunteer_InvalidGender_Returns400(t *testing.T) {
	repo := &mockVolunteerRepo{
		createFn: func(ctx context.Context, req CreateRequest, actorID int) (*Volunteer, error) {
			t.Fatalf("repository must not be called when validation fails")
			return nil, nil
		},
	}
	h := &Handler{Service: &Service{Repo: repo}}
	r := newRouter(h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/volunteers", strings.NewReader(validCreateBody("abc")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d (body=%s)", w.Code, w.Body.String())
	}
}
