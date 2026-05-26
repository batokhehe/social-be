package email

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"social-be/internal/pkg/apperror"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
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
	Detail  interface{} `json:"detail,omitempty"`
}

func TestSendDummy_InvalidBody(t *testing.T) {
	h := &Handler{Service: &Service{sender: func(ctx context.Context, to, subject, body string) error {
		return nil
	}}}
	c, w := newTestContext(http.MethodPost, "/email/test", `{"email":`)

	h.Send(c)

	if !c.IsAborted() {
		t.Fatal("expected context to be aborted for invalid JSON")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected recorder code to stay 200 before middleware writes error, got %d", w.Code)
	}
}

func TestSendDummy_Success(t *testing.T) {
	h := &Handler{Service: &Service{sender: func(ctx context.Context, to, subject, body string) error {
		return nil
	}}}
	c, w := newTestContext(http.MethodPost, "/email/test", `{"email":"tester@example.com"}`)

	h.Send(c)

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

	data, ok := got.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be an object, got %#v", got.Data)
	}
	if data["to"] != "tester@example.com" {
		t.Fatalf("expected email to be echoed, got %#v", data)
	}
	if data["status"] != "queued" {
		t.Fatalf("expected status queued, got %#v", data)
	}
	if data["dummy"] != false {
		t.Fatalf("expected dummy flag false, got %#v", data)
	}
	if data["message"] != "email queued for async delivery" {
		t.Fatalf("expected queued message, got %#v", data)
	}
}
