package smtpmailer

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

type Mailer struct {
	host string
	port string
	user string
	pass string
	from string
}

func New() *Mailer {
	from := os.Getenv("SMTP_FROM")
	if from == "" {
		from = os.Getenv("SMTP_USER")
	}
	return &Mailer{
		host: getEnv("SMTP_HOST", "smtp.gmail.com"),
		port: getEnv("SMTP_PORT", "587"),
		user: os.Getenv("SMTP_USER"),
		pass: os.Getenv("SMTP_PASS"),
		from: from,
	}
}

func (m *Mailer) SendOTP(email, code string) error {
	if m.user == "" || m.pass == "" {
		// Dev-mode: выводим код в лог вместо реальной отправки.
		log.Printf("[AUTH DEV] OTP для %s: %s", email, code)
		return nil
	}

	auth := smtp.PlainAuth("", m.user, m.pass, m.host)
	body := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: Код входа в систему бронирования\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n"+
			"Ваш код подтверждения: %s\r\n\r\nКод действителен 10 минут. Не передавайте его никому.",
		m.from, email, code,
	)

	addr := fmt.Sprintf("%s:%s", m.host, m.port)
	if err := smtp.SendMail(addr, auth, m.from, []string{email}, []byte(body)); err != nil {
		log.Printf("SMTP send error to %s: %v", email, err)
		return fmt.Errorf("не удалось отправить письмо")
	}
	return nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
