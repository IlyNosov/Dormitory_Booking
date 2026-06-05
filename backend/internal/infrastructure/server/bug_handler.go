package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type BugNotifier interface {
	Send(text string) error
}

type bugReport struct {
	Description string `json:"description"`
	URL         string `json:"url"`
	UserAgent   string `json:"userAgent"`
}

var mskBug = time.FixedZone("MSK", 3*60*60)

func (h *BookingHandlers) ReportBug(w http.ResponseWriter, r *http.Request) {
	var rep bugReport
	if err := json.NewDecoder(r.Body).Decode(&rep); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	userEmail := "не авторизован"
	if u, ok := userFromCtx(r); ok {
		userEmail = u.Email
	}

	now := time.Now().In(mskBug).Format("02.01.2006 15:04:05")
	tgText := fmt.Sprintf("🐛 Баг-репорт\n👤 %s\n🕐 %s\n🌐 %s\n🖥 %s\n\n%s",
		userEmail, now, rep.URL, rep.UserAgent, rep.Description,
	)

	if h.bugNotifier != nil {
		if err := h.bugNotifier.Send(tgText); err != nil {
			log.Printf("bug report telegram send error: %v", err)
		}
	}

	if h.mailer != nil {
		body := fmt.Sprintf(
			"Новый баг-репорт\n\nВремя: %s\nПользователь: %s\nURL: %s\nОписание: %s\nUA: %s",
			now, userEmail, rep.URL, rep.Description, rep.UserAgent,
		)
		if err := h.mailer.Send("irnosov@edu.hse.ru", "Баг на dubki.fun", body); err != nil {
			log.Printf("bug report email failed: %v", err)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}
