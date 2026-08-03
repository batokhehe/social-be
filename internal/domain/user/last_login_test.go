package user

import (
	"testing"
	"time"
)

// toEntity must map the three last-login columns through to the API entity.
func TestToEntity_MapsLastLoginFields(t *testing.T) {
	at := time.Date(2026, time.July, 25, 8, 30, 0, 0, time.UTC)
	ip := "203.0.113.7"
	ua := "Mozilla/5.0 (test)"

	got := toEntity(userModel{
		ID: 5, Name: "Admin", Email: "a@x.io", Role: 1, Status: 1,
		LastLoginAt: &at, LastLoginIP: &ip, LastLoginUserAgent: &ua,
	})

	if got.LastLoginAt == nil || !got.LastLoginAt.Equal(at) {
		t.Fatalf("last_login_at = %v, want %v", got.LastLoginAt, at)
	}
	if got.LastLoginIP == nil || *got.LastLoginIP != ip {
		t.Fatalf("last_login_ip = %v, want %s", got.LastLoginIP, ip)
	}
	if got.LastLoginUserAgent == nil || *got.LastLoginUserAgent != ua {
		t.Fatalf("last_login_user_agent = %v, want %s", got.LastLoginUserAgent, ua)
	}
}

// A never-logged-in user maps to nil pointers (serialized as absent/null).
func TestToEntity_LastLoginNilByDefault(t *testing.T) {
	got := toEntity(userModel{ID: 1, Name: "New", Email: "n@x.io"})
	if got.LastLoginAt != nil || got.LastLoginIP != nil || got.LastLoginUserAgent != nil {
		t.Fatalf("expected nil last-login fields, got %+v", got)
	}
}
