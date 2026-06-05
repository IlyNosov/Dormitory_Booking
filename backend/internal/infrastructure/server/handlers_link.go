package server

import (
	"encoding/json"
	"net/http"
	"os"

	apptglink "Dormitory_Booking/internal/application/tglink"
	"Dormitory_Booking/internal/domain/tglink"
)

type LinkHandlers struct {
	svc       *apptglink.Service
	botSecret string
}

func NewLinkHandlers(svc *apptglink.Service) *LinkHandlers {
	return &LinkHandlers{svc: svc, botSecret: os.Getenv("BOT_INTERNAL_SECRET")}
}

// sessionID extracts the value of the "session" cookie used for auth.
func sessionID(r *http.Request) string {
	c, err := r.Cookie("session")
	if err != nil {
		return ""
	}
	return c.Value
}

// Status returns the Telegram link state for the current session.
func (h *LinkHandlers) Status(w http.ResponseWriter, r *http.Request) {
	sid := sessionID(r)
	if sid == "" {
		http.Error(w, "no session", http.StatusUnauthorized)
		return
	}
	link, err := h.svc.GetOrCreate(sid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{
		"confirmed":  link.Confirmed,
		"telegramId": link.TelegramID,
	})
}

// Generate creates a one-time link token for the current session.
func (h *LinkHandlers) Generate(w http.ResponseWriter, r *http.Request) {
	sid := sessionID(r)
	if sid == "" {
		http.Error(w, "no session", http.StatusUnauthorized)
		return
	}
	token, err := h.svc.GenerateToken(sid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"token": token})
}

// Confirm is called by the Telegram bot after the user sends the token.
// Validated via X-Bot-Secret header rather than user session.
func (h *LinkHandlers) Confirm(w http.ResponseWriter, r *http.Request) {
	if h.botSecret != "" && r.Header.Get("X-Bot-Secret") != h.botSecret {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	var body struct {
		Token      string `json:"token"`
		TelegramID string `json:"telegramId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := h.svc.ConfirmLink(body.Token, body.TelegramID); err != nil {
		if err == tglink.ErrTokenExpired || err == tglink.ErrNotFound {
			http.Error(w, "token not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Unlink removes the Telegram association for the current session.
func (h *LinkHandlers) Unlink(w http.ResponseWriter, r *http.Request) {
	sid := sessionID(r)
	if sid == "" {
		http.Error(w, "no session", http.StatusUnauthorized)
		return
	}
	if err := h.svc.Unlink(sid); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
