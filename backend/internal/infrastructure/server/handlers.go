package server

// В этом файле HTTP-обработчики для бронирований, администратора и привязки Telegram.

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	appbooking "Dormitory_Booking/internal/application/booking"
	apptglink "Dormitory_Booking/internal/application/tglink"
	domain "Dormitory_Booking/internal/domain/booking"
	"Dormitory_Booking/internal/domain/tglink"
)

type Handlers struct {
	svc           *appbooking.Service
	linkSvc       *apptglink.Service
	adminPassword string
}

func normalizeTG(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "@")
	s = strings.ToLower(s)
	return s
}

func NewHandlers(svc *appbooking.Service, linkSvc *apptglink.Service) *Handlers {
	return &Handlers{
		svc:           svc,
		linkSvc:       linkSvc,
		adminPassword: os.Getenv("ADMIN_PASSWORD"),
	}
}

// Логин в админку

func (h *Handlers) AdminLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if h.adminPassword == "" || body.Password != h.adminPassword {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "admin_token",
		Value:    "1",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) AdminLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) isAdmin(r *http.Request) bool {
	if tok := r.Header.Get("X-Admin-Token"); tok != "" && tok == os.Getenv("ADMIN_TOKEN") {
		return true
	}
	c, err := r.Cookie("admin_token")
	return err == nil && c.Value == "1"
}

// sessionID извлекает идентификатор сессии из заголовка X-Session-ID.
func sessionID(r *http.Request) string {
	return strings.TrimSpace(r.Header.Get("X-Session-ID"))
}

// Бронирования

func (h *Handlers) GetAll(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListBookings(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sid := sessionID(r)
	var currentTgID string
	if sid != "" && h.linkSvc != nil {
		currentTgID, _ = h.linkSvc.GetLinkedTelegramID(sid)
	}

	out := make([]appbooking.BookingDTO, 0, len(list))
	for _, b := range list {
		out = append(out, appbooking.ToDTO(b, currentTgID, h.isAdmin(r)))
	}
	writeJSON(w, out)
}

func (h *Handlers) GetOne(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	b, err := h.svc.GetBooking(r.Context(), id)
	if err != nil {
		if err == domain.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, appbooking.ToDTO(b, "", h.isAdmin(r)))
}

func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Start       string `json:"start"`
		End         string `json:"end"`
		Room        int    `json:"room"`
		Title       string `json:"title"`
		Description string `json:"description"`
		TelegramID  string `json:"telegramId"`
		IsPrivate   bool   `json:"isPrivate"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Если TelegramID не передан с фронта — пробуем взять из привязки сессии.
	telegramID := strings.TrimSpace(body.TelegramID)
	if telegramID == "" && h.linkSvc != nil {
		if sid := sessionID(r); sid != "" {
			if linked, ok := h.linkSvc.GetLinkedTelegramID(sid); ok {
				telegramID = linked
			}
		}
	}

	start, err := time.Parse(time.RFC3339, body.Start)
	if err != nil {
		http.Error(w, "invalid start time", http.StatusBadRequest)
		return
	}
	end, err := time.Parse(time.RFC3339, body.End)
	if err != nil {
		http.Error(w, "invalid end time", http.StatusBadRequest)
		return
	}

	input := appbooking.CreateBookingInput{
		Start:       start,
		End:         end,
		Room:        domain.Room(body.Room),
		Title:       body.Title,
		Description: body.Description,
		TelegramID:  telegramID,
		IsPrivate:   body.IsPrivate,
	}

	b, err := h.svc.CreateBooking(r.Context(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, appbooking.ToDTO(b, telegramID, h.isAdmin(r)))
}

func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	isAdmin := h.isAdmin(r)

	requesterID := r.URL.Query().Get("tg")
	if requesterID == "" {
		requesterID = r.Header.Get("X-User-TelegramID")
	}
	// Fallback: по сессии
	if requesterID == "" && h.linkSvc != nil {
		if sid := sessionID(r); sid != "" {
			if linked, ok := h.linkSvc.GetLinkedTelegramID(sid); ok {
				requesterID = linked
			}
		}
	}

	err := h.svc.DeleteBooking(r.Context(), id, requesterID, isAdmin)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Привязка Telegram

// LinkStatus возвращает статус привязки для текущей сессии.
func (h *Handlers) LinkStatus(w http.ResponseWriter, r *http.Request) {
	if h.linkSvc == nil {
		writeJSON(w, map[string]any{"linked": false, "botDisabled": true})
		return
	}
	sid := sessionID(r)
	if sid == "" {
		http.Error(w, "X-Session-ID required", http.StatusBadRequest)
		return
	}

	l, _ := h.linkSvc.GetOrCreate(sid)
	writeJSON(w, map[string]any{
		"linked":     l.Confirmed,
		"telegramId": l.TelegramID,
	})
}

// LinkGenerate создаёт новый токен привязки и возвращает его.
func (h *Handlers) LinkGenerate(w http.ResponseWriter, r *http.Request) {
	if h.linkSvc == nil {
		http.Error(w, "bot not configured", http.StatusServiceUnavailable)
		return
	}
	sid := sessionID(r)
	if sid == "" {
		http.Error(w, "X-Session-ID required", http.StatusBadRequest)
		return
	}

	token, err := h.linkSvc.GenerateToken(sid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"token": token})
}

// LinkConfirm вызывается ботом: подтверждает токен и записывает telegramID.
// Защищён секретом BOT_INTERNAL_SECRET.
func (h *Handlers) LinkConfirm(w http.ResponseWriter, r *http.Request) {
	if h.linkSvc == nil {
		http.Error(w, "bot not configured", http.StatusServiceUnavailable)
		return
	}

	secret := os.Getenv("BOT_INTERNAL_SECRET")
	if secret == "" || r.Header.Get("X-Bot-Secret") != secret {
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

	if err := h.linkSvc.ConfirmLink(body.Token, body.TelegramID); err != nil {
		if errors.Is(err, tglink.ErrTokenExpired) {
			http.Error(w, "token not found or expired", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// LinkUnlink сбрасывает привязку.
func (h *Handlers) LinkUnlink(w http.ResponseWriter, r *http.Request) {
	if h.linkSvc == nil {
		http.Error(w, "bot not configured", http.StatusServiceUnavailable)
		return
	}
	sid := sessionID(r)
	if sid == "" {
		http.Error(w, "X-Session-ID required", http.StatusBadRequest)
		return
	}
	_ = h.linkSvc.Unlink(sid)
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
