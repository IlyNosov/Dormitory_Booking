package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	appbooking "Dormitory_Booking/internal/application/booking"
	apptglink "Dormitory_Booking/internal/application/tglink"
)

func NewRouter(svc *appbooking.Service, linkSvc *apptglink.Service) http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Admin-Token", "X-Session-ID", "X-User-TelegramID", "X-Bot-Secret"},
		AllowCredentials: true,
	}))

	h := NewHandlers(svc, linkSvc)

	// логин в админку
	r.Post("/admin/login", h.AdminLogin)
	r.Post("/admin/logout", h.AdminLogout)

	// брони
	r.Get("/bookings", h.GetAll)
	r.Get("/bookings/{id}", h.GetOne)
	r.Post("/bookings", h.Create)
	r.Delete("/bookings/{id}", h.Delete)

	// привязка Telegram
	r.Get("/link/telegram", h.LinkStatus)
	r.Post("/link/telegram", h.LinkGenerate)
	r.Post("/link/telegram/confirm", h.LinkConfirm)
	r.Delete("/link/telegram", h.LinkUnlink)

	return r
}
