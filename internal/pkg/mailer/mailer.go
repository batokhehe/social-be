package mailer

import (
	"context"
	"fmt"
	"net/smtp"
	"os"
	"strings"

	"social-be/internal/pkg/logger"
)

func SendPasswordReset(ctx context.Context, to, name, resetLink string) error {
	host := os.Getenv("SMTP_HOST")
	from := os.Getenv("SMTP_FROM")
	if host == "" || from == "" {
		logger.FromContext(ctx).
			WithField("email", to).
			WithField("reset_link", resetLink).
			Info("smtp not configured; password reset email skipped")
		return nil
	}

	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}

	subject := "Set your volunteer account password"
	body := fmt.Sprintf("Hi %s,\n\nYour volunteer account has been created.\nPlease set your password using this link:\n\n%s\n\nThis link will expire in 24 hours.\n", name, resetLink)
	message := strings.Join([]string{
		"From: " + from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=\"UTF-8\"",
		"",
		body,
	}, "\r\n")

	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	var auth smtp.Auth
	if username != "" || password != "" {
		auth = smtp.PlainAuth("", username, password, host)
	}

	return smtp.SendMail(host+":"+port, auth, from, []string{to}, []byte(message))
}
