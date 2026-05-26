package email

import (
	"context"
	"testing"
	"time"
)

func TestService_Send_QueuesAsync(t *testing.T) {
	called := make(chan struct{}, 1)
	svc := &Service{sender: func(ctx context.Context, to, subject, body string) error {
		called <- struct{}{}
		return nil
	}}

	resp, err := svc.Send(context.Background(), SendRequest{Email: "tester@example.com"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Status != "queued" {
		t.Fatalf("expected queued status, got %s", resp.Status)
	}
	if resp.Dummy {
		t.Fatal("expected dummy flag to be false")
	}
	if resp.Message != "email queued for async delivery" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}

	select {
	case <-called:
	case <-time.After(time.Second):
		t.Fatal("expected async sender to be invoked")
	}
}
