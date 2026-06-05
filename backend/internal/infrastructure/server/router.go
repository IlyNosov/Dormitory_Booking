package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	appbooking "Dormitory_Booking/internal/application/booking"
	appauth "Dormitory_Booking/internal/application/auth"
	apptglink "Dormitory_Booking/internal/application/tglink"
)

func NewRouter(svc *appbooking.Service, linkSvc *apptglink.Service, authSvc *appauth.Service) http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Admin-Token", "X-Session-ID", "X-User-TelegramID", "X-Bot-Secret"},
		AllowCredentials: true,
	}))

	// Сессионный middleware: читает cookie `session`, кладёт User в контекст.
	r.Use(SessionMiddleware(authSvc))

	h := NewHandlers(svc, linkSvc, authSvc)

	// ── Аутентификация ───────────────────────────────────────────────────────
	r.Post("/auth/otp/request", h.AuthRequestOTP)
	r.Post("/auth/otp/verify", h.AuthVerifyOTP)
	r.Post("/auth/logout", h.AuthLogout)
	r.Get("/auth/me", h.AuthMe)

	// ── Legacy admin login ────────────────────────────────────────────────────
	r.Post("/admin/login", h.AdminLogin)
	r.Post("/admin/logout", h.AdminLogout)

	// ── Бронирования ─────────────────────────────────────────────────────────
	r.Get("/bookings", h.GetAll)
	r.Get("/bookings/{id}", h.GetOne)
	r.Post("/bookings", h.Create)
	r.Delete("/bookings/{id}", h.Delete)

	// ── Привязка Telegram ─────────────────────────────────────────────────────
	r.Get("/link/telegram", h.LinkStatus)
	r.Post("/link/telegram", h.LinkGenerate)
	r.Post("/link/telegram/confirm", h.LinkConfirm)
	r.Delete("/link/telegram", h.LinkUnlink)

	return r
}
