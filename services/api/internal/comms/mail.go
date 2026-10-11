// Package comms sends outbound messages. Only email exists so far; SMS/push/WhatsApp
// providers plug in behind the same idea once accounts exist (T-190).
package comms

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
)

type Email struct {
	To      string
	Subject string
	Text    string
}

type Mailer interface {
	Send(ctx context.Context, e Email) error
}

// DevTokens makes token-bearing responses (verification, reset) include the raw token so the
// flow can be completed without a mail provider. Never set in production.
func DevTokens() bool { return os.Getenv("TAXIOS_DEV_TOKENS") == "1" }

// FromEnv returns an SMTP mailer when TAXIOS_SMTP_HOST is set, otherwise a logging mailer.
func FromEnv() Mailer {
	host := os.Getenv("TAXIOS_SMTP_HOST")
	if host == "" {
		return LogMailer{}
	}
	port := os.Getenv("TAXIOS_SMTP_PORT")
	if port == "" {
		port = "587"
	}
	return &SMTPMailer{Addr: net.JoinHostPort(host, port), User: os.Getenv("TAXIOS_SMTP_USER"),
		Password: os.Getenv("TAXIOS_SMTP_PASSWORD"), From: os.Getenv("TAXIOS_EMAIL_FROM")}
}

// LogMailer records that a message would have been sent. It never logs the body (tokens).
type LogMailer struct{}

func (LogMailer) Send(_ context.Context, e Email) error {
	slog.Warn("email_not_delivered_no_provider", "subject", e.Subject, "to_domain", domain(e.To))
	return nil
}

// SMTPMailer sends over STARTTLS (port 587) or implicit TLS (port 465) using the standard library.
type SMTPMailer struct {
	Addr, User, Password, From string
}

func (m *SMTPMailer) Send(_ context.Context, e Email) error {
	if _, err := mail.ParseAddress(e.To); err != nil {
		return fmt.Errorf("invalid recipient")
	}
	host, port, _ := net.SplitHostPort(m.Addr)
	msg := strings.NewReplacer("\r", "", "\n", "\r\n").Replace(fmt.Sprintf(
		"From: %s\nTo: %s\nSubject: %s\nMIME-Version: 1.0\nContent-Type: text/plain; charset=UTF-8\n\n%s\n",
		m.From, e.To, strings.ReplaceAll(e.Subject, "\n", " "), e.Text))
	auth := smtp.PlainAuth("", m.User, m.Password, host)
	if port == "465" {
		conn, err := tls.Dial("tcp", m.Addr, &tls.Config{ServerName: host})
		if err != nil {
			return err
		}
		c, err := smtp.NewClient(conn, host)
		if err != nil {
			return err
		}
		defer c.Close()
		return send(c, auth, m.From, e.To, msg)
	}
	c, err := smtp.Dial(m.Addr)
	if err != nil {
		return err
	}
	defer c.Close()
	if err := c.StartTLS(&tls.Config{ServerName: host}); err != nil {
		return err
	}
	return send(c, auth, m.From, e.To, msg)
}

func send(c *smtp.Client, auth smtp.Auth, from, to, msg string) error {
	if err := c.Auth(auth); err != nil {
		return err
	}
	if err := c.Mail(from); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

func domain(addr string) string {
	if i := strings.LastIndex(addr, "@"); i >= 0 {
		return addr[i+1:]
	}
	return ""
}
