package security

import "testing"

func TestGetIntClaim(t *testing.T) {
	claims := map[string]interface{}{"user_id": float64(10)}

	id, err := GetIntClaim(claims, "user_id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != 10 {
		t.Fatalf("expected 10 got %d", id)
	}
}

func TestGetStringClaim(t *testing.T) {
	claims := map[string]interface{}{"email": "a@b.com"}

	email, err := GetStringClaim(claims, "email")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if email != "a@b.com" {
		t.Fatalf("unexpected value: %s", email)
	}
}
