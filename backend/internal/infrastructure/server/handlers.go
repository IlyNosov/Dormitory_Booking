package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	appbooking "Dormitory_Booking/internal/application/booking"
	domain "Dormitory_Booking/internal/domain/booking"
	"Dormitory_Booking/internal/infrastructure/mailer"
)

type BookingHandlers struct {
	svc         *appbooking.Service
	mailer      *mailer.Mailer
	hub         *Hub
	bugNotifier BugNotifier
}

func NewBookingHandlers(svc *appbooking.Service, ml *mailer.Mailer, hub *Hub, bugNotifier BugNotifier) *BookingHandlers {
	return &BookingHandlers{svc: svc, mailer: ml, hub: hub, bugNotifier: bugNotifier}
}

func (h *BookingHandlers) GetAll(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListBookings(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	u, _ := userFromCtx(r)
	isAdmin := isAdminFromCtx(r)

	out := make([]appbooking.BookingDTO, 0, len(list))
	for _, b := range list {
		out = append(out, appbooking.ToDTO(b, u.Email, "", isAdmin))
	}
	writeJSON(w, out)
}

func (h *BookingHandlers) GetOne(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	b, err := h.svc.GetBooking(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	u, _ := userFromCtx(r)
	writeJSON(w, appbooking.ToDTO(b, u.Email, "", isAdminFromCtx(r)))
}

func (h *BookingHandlers) Create(w http.ResponseWriter, r *http.Request) {
	u, ok := userFromCtx(r)
	if !ok {
		http.Error(w, "требуется авторизация", http.StatusUnauthorized)
		return
	}

	var body struct {
		Start       string `json:"start"`
		End         string `json:"end"`
		Room        int    `json:"room"`
		Title       string `json:"title"`
		Description string `json:"description"`
		IsPrivate   bool   `json:"isPrivate"`
		TelegramID  string `json:"telegram_id"`
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

	b, err := h.svc.CreateBooking(r.Context(), appbooking.CreateBookingInput{
		Start: start, End: end, Room: body.Room,
		Title: body.Title, Description: body.Description,
		UserEmail: u.Email, IsPrivate: body.IsPrivate,
		TelegramID: body.TelegramID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.hub.Notify()
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, appbooking.ToDTO(b, u.Email, "", isAdminFromCtx(r)))
}

func (h *BookingHandlers) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	u, ok := userFromCtx(r)
	if !ok {
		http.Error(w, "требуется авторизация", http.StatusUnauthorized)
		return
	}

	var body struct {
		Start       string `json:"start"`
		End         string `json:"end"`
		Room        int    `json:"room"`
		Title       string `json:"title"`
		Description string `json:"description"`
		IsPrivate   bool   `json:"isPrivate"`
		TelegramID  string `json:"telegram_id"`
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

	isAdmin := isAdminFromCtx(r)
	b, err := h.svc.UpdateBooking(r.Context(), appbooking.UpdateBookingInput{
		ID: id, Start: start, End: end, Room: body.Room,
		Title: body.Title, Description: body.Description,
		TelegramID: body.TelegramID, IsPrivate: body.IsPrivate,
	}, u.Email, isAdmin)
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
	h.hub.Notify()
	writeJSON(w, appbooking.ToDTO(b, u.Email, "", isAdmin))
}

// ForceCreate is admin-only: bypasses business rules.
func (h *BookingHandlers) ForceCreate(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromCtx(r)

	var body struct {
		Start       string `json:"start"`
		End         string `json:"end"`
		Room        int    `json:"room"`
		Title       string `json:"title"`
		Description string `json:"description"`
		IsPrivate   bool   `json:"isPrivate"`
		TelegramID  string `json:"telegram_id"`
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

	b, err := h.svc.CreateBooking(r.Context(), appbooking.CreateBookingInput{
		Start: start, End: end, Room: body.Room,
		Title: body.Title, Description: body.Description,
		UserEmail: u.Email, IsPrivate: body.IsPrivate,
		TelegramID: body.TelegramID,
		Force: true,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.hub.Notify()
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, appbooking.ToDTO(b, u.Email, "", true))
}

func (h *BookingHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	u, _ := userFromCtx(r)
	isAdmin := isAdminFromCtx(r)

	tgID := r.Header.Get("X-User-TelegramID")

	err := h.svc.DeleteBooking(r.Context(), id, u.Email, tgID, isAdmin)
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
	h.hub.Notify()
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
