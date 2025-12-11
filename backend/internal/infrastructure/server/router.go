package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	appbooking "Dormitory_Booking/internal/application/booking"
)

func NewRouter(svc *appbooking.Service) http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	h := NewHandlers(svc)

	// логин в админку
	r.Post("/admin/login", h.AdminLogin)
	r.Post("/admin/logout", h.AdminLogout)

	// брони
	r.Get("/bookings", h.GetAll)
	r.Get("/bookings/{id}", h.GetOne)
	r.Post("/bookings", h.Create)
	r.Delete("/bookings/{id}", h.Delete)

	return r
}
