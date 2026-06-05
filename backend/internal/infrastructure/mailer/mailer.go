package mailer

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
)

type Mailer struct {
	host     string
	port     string
	user     string
	from     string
	password string
}

func New() *Mailer {
	from := os.Getenv("SMTP_FROM")
	user := os.Getenv("SMTP_USER")
	if user == "" {
		user = from
	}
	return &Mailer{
		host:     os.Getenv("SMTP_HOST"),
		port:     getEnv("SMTP_PORT", "587"),
		user:     user,
		from:     from,
		password: os.Getenv("SMTP_PASSWORD"),
	}
}

func (m *Mailer) SendOTP(to, code string) error {
	if m.host == "" {
		log.Printf("[DEV MAILER] OTP for %s: %s", to, code)
		return nil
	}

	subject := "Код подтверждения — Бронирование досуговых"
	body := fmt.Sprintf(
		"Ваш код подтверждения: %s\n\nКод действует 10 минут.\nЕсли вы не запрашивали код — проигнорируйте это письмо.",
		code,
	)

	msg := strings.Join([]string{
		"From: " + m.from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	addr := m.host + ":" + m.port
	auth := smtp.PlainAuth("", m.user, m.password, m.host)
	return smtp.SendMail(addr, auth, m.from, []string{to}, []byte(msg))
}

func (m *Mailer) Send(to, subject, body string) error {
	if m.host == "" {
		log.Printf("[DEV MAILER] To=%s Subject=%s Body=%s", to, subject, body)
		return nil
	}
	msg := strings.Join([]string{
		"From: " + m.from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")
	addr := m.host + ":" + m.port
	auth := smtp.PlainAuth("", m.user, m.password, m.host)
	return smtp.SendMail(addr, auth, m.from, []string{to}, []byte(msg))
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
