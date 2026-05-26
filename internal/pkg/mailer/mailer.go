package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"

	"social-be/internal/pkg/logger"
)

func SendPasswordReset(ctx context.Context, to, name, resetLink string) error {
	subject := "Set your volunteer account password"
	body := fmt.Sprintf("Hi %s,\n\nYour volunteer account has been created.\nPlease set your password using this link:\n\n%s\n\nThis link will expire in 24 hours.\n", name, resetLink)
	return SendEmail(ctx, to, subject, body)
}

func SendEmail(ctx context.Context, to, subject, body string) error {
	host := os.Getenv("SMTP_HOST")
	from := os.Getenv("SMTP_FROM")
	if host == "" || from == "" {
		logger.FromContext(ctx).
			WithField("email", to).
			WithField("subject", subject).
			Info("smtp not configured; email skipped")
		return nil
	}

	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}

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

	return sendSMTP(host, port, auth, from, to, []byte(message))
}

func sendSMTP(host, port string, auth smtp.Auth, from, to string, message []byte) error {
	addr := net.JoinHostPort(host, port)
	if port == "465" {
		conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
		if err != nil {
			return err
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, host)
		if err != nil {
			return err
		}
		defer client.Close()
		return sendSMTPMessage(client, auth, from, to, message)
	}

	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: host}); err != nil {
			return err
		}
	}

	return sendSMTPMessage(client, auth, from, to, message)
}

func sendSMTPMessage(client *smtp.Client, auth smtp.Auth, from, to string, message []byte) error {
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}

	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	return client.Quit()
}
