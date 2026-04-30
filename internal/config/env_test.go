package config

import "testing"

func TestRequireEnv(t *testing.T) {
	t.Setenv("A", "1")
	t.Setenv("B", "2")

	if err := RequireEnv("A", "B"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := RequireEnv("A", "C"); err == nil {
		t.Fatal("expected error for missing env")
	}
}
