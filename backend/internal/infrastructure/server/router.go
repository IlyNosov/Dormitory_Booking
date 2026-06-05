package server

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	appbooking "Dormitory_Booking/internal/application/booking"
	authsvc "Dormitory_Booking/internal/application/auth"
	adminsvc "Dormitory_Booking/internal/application/admin"
	approom "Dormitory_Booking/internal/application/room"
	apptglink "Dormitory_Booking/internal/application/tglink"
	"Dormitory_Booking/internal/infrastructure/mailer"
)

type Services struct {
	Booking    *appbooking.Service
	Auth       *authsvc.Service
	Admin      *adminsvc.Service
	Room       *approom.Service
	Mailer     *mailer.Mailer
	Hub        *Hub
	LinkSvc    *apptglink.Service
	TGNotifier BugNotifier
}

func NewRouter(svcs Services) http.Handler {
	r := chi.NewRouter()

	allowedOrigin := os.Getenv("CORS_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "*"
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{allowedOrigin},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "X-User-TelegramID"},
		AllowCredentials: true,
	}))

	r.Use(SessionMiddleware(svcs.Auth, svcs.Admin))

	bookingH := NewBookingHandlers(svcs.Booking, svcs.Mailer, svcs.Hub, svcs.TGNotifier)
	authH := NewAuthHandlers(svcs.Auth, svcs.Admin)
	roomH := NewRoomHandlers(svcs.Room)
	adminH := NewAdminHandlers(svcs.Admin)
	linkH := NewLinkHandlers(svcs.LinkSvc)

	// Bug reports — public, no auth
	r.Post("/bugs", bookingH.ReportBug)

	// Auth
	r.Post("/auth/otp/request", authH.RequestOTP)
	r.Post("/auth/otp/verify", authH.VerifyOTP)
	r.Post("/auth/logout", authH.Logout)
	r.Get("/auth/me", authH.Me)

	// Public booking reads — includes SSE stream
	r.Get("/bookings/events", svcs.Hub.Handler())
	r.Get("/bookings", bookingH.GetAll)
	r.Get("/bookings/{id}", bookingH.GetOne)

	// Authenticated: create/update/delete
	r.Group(func(r chi.Router) {
		r.Use(RequireAuth)
		r.Post("/bookings", bookingH.Create)
		r.Patch("/bookings/{id}", bookingH.Update)
		r.Delete("/bookings/{id}", bookingH.Delete)
	})

	// Rooms list — public
	r.Get("/rooms", roomH.List)

	// Telegram link — requires auth
	r.Group(func(r chi.Router) {
		r.Use(RequireAuth)
		r.Get("/link/telegram", linkH.Status)
		r.Post("/link/telegram", linkH.Generate)
		r.Delete("/link/telegram", linkH.Unlink)
	})
	// Called by the bot (validated via X-Bot-Secret header, not user session)
	r.Post("/link/telegram/confirm", linkH.Confirm)

	// Admin only
	r.Group(func(r chi.Router) {
		r.Use(RequireAdmin)
		r.Post("/bookings/force", bookingH.ForceCreate)
		r.Post("/rooms", roomH.Create)
		r.Delete("/rooms/{number}", roomH.Delete)
		r.Get("/admins", adminH.List)
		r.Post("/admins", adminH.Add)
		r.Delete("/admins", adminH.Remove)
	})

	return r
}
