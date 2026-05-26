package email

import (
	"context"
	"fmt"
	"time"

	"social-be/internal/pkg/logger"
	"social-be/internal/pkg/mailer"
)

type Sender func(ctx context.Context, to, subject, body string) error

type Service struct {
	sender Sender
}

func NewService() *Service {
	return &Service{sender: mailer.SendEmail}
}

func (s *Service) Send(ctx context.Context, req SendRequest) (*SendResponse, error) {
	sender := s.sender
	if sender == nil {
		sender = mailer.SendEmail
	}

	sendCtx := context.WithoutCancel(ctx)
	go func() {
		if err := sender(sendCtx, req.Email, "SMTP check", fmt.Sprintf("This is a test email sent from social-be for %s.\n", req.Email)); err != nil {
			logger.FromContext(sendCtx).WithError(err).WithField("email", req.Email).Error("failed to send async email")
			return
		}
		logger.FromContext(sendCtx).WithField("email", req.Email).Info("async email sent")
	}()

	return &SendResponse{
		To:          req.Email,
		Status:      "queued",
		Dummy:       false,
		Message:     "email queued for async delivery",
		SimulatedAt: time.Now().UTC(),
	}, nil
}
