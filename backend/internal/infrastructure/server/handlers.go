package server

// В этом файле HTTP-обработчики для бронирований и простая админ-авторизация.

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"

	appbooking "Dormitory_Booking/internal/application/booking"
	domain "Dormitory_Booking/internal/domain/booking"
)

type Handlers struct {
	svc           *appbooking.Service
	adminPassword string
}

func NewHandlers(svc *appbooking.Service) *Handlers {
	return &Handlers{
		svc:           svc,
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
	c, err := r.Cookie("admin_token")
	return err == nil && c.Value == "1"
}

// Бронирования

func (h *Handlers) GetAll(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListBookings(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	out := make([]appbooking.BookingDTO, 0, len(list))
	for _, b := range list {
		out = append(out, appbooking.ToDTO(b, "", h.isAdmin(r)))
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
		TelegramID:  body.TelegramID,
		IsPrivate:   body.IsPrivate,
	}

	b, err := h.svc.CreateBooking(r.Context(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, appbooking.ToDTO(b, body.TelegramID, h.isAdmin(r)))
}

func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	isAdmin := h.isAdmin(r)

	// пользователь без логина удаляет "свою" бронь по telegramId, переданному в query
	requesterID := r.URL.Query().Get("tg")

	err := h.svc.DeleteBooking(r.Context(), id, requesterID, isAdmin)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
