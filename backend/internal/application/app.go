package app

import (
	appbooking "Dormitory_Booking/internal/application/booking"
	appadmin "Dormitory_Booking/internal/application/admin"
	appauth "Dormitory_Booking/internal/application/auth"
	approom "Dormitory_Booking/internal/application/room"
	apptglink "Dormitory_Booking/internal/application/tglink"
	domainbooking "Dormitory_Booking/internal/domain/booking"
	tgbot "Dormitory_Booking/internal/infrastructure/booking_bot"
	"Dormitory_Booking/internal/infrastructure/botuser"
	"Dormitory_Booking/internal/infrastructure/mailer"
	"Dormitory_Booking/internal/infrastructure/memory"
	notifier "Dormitory_Booking/internal/infrastructure/notifier_bot"
	pgrepo "Dormitory_Booking/internal/infrastructure/postgres"
	"Dormitory_Booking/internal/infrastructure/server"
	tglinkstore "Dormitory_Booking/internal/infrastructure/tglink"
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func Run(ctx context.Context) error {
	_ = godotenv.Load()

	addr := getEnv("HTTP_ADDR", ":8080")
	dbURL := os.Getenv("DB_URL")
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	var pool *pgxpool.Pool
	var bookingRepo domainbooking.Repository

	if dbURL != "" {
		log.Printf("connecting to Postgres: %s", dbURL)
		var err error
		pool, err = pgxpool.New(ctx, dbURL)
		if err != nil {
			return err
		}
		defer pool.Close()
		bookingRepo = pgrepo.NewBookingPostgresRepo(pool)
	} else {
		log.Println("DB_URL not set, using in-memory repo")
		bookingRepo = memory.NewInMemoryBookingRepo()
	}

	// Repos
	var (
		userRepo    *pgrepo.UserRepo
		sessionRepo *pgrepo.SessionRepo
		otpRepo     *pgrepo.OTPRepo
		adminRepo   *pgrepo.AdminRepo
		roomRepo    *pgrepo.RoomRepo
	)
	if pool != nil {
		userRepo = pgrepo.NewUserRepo(pool)
		sessionRepo = pgrepo.NewSessionRepo(pool)
		otpRepo = pgrepo.NewOTPRepo(pool)
		adminRepo = pgrepo.NewAdminRepo(pool)
		roomRepo = pgrepo.NewRoomRepo(pool)
	}

	// Services
	ml := mailer.New()
	authSvc := appauth.NewService(userRepo, sessionRepo, otpRepo, ml)
	adminSvc := appadmin.NewService(adminRepo)
	roomSvc := approom.NewService(roomRepo)

	// TG link service
	var linkSvc *apptglink.Service
	if pool != nil {
		linkSvc = apptglink.NewService(tglinkstore.NewPostgresStore(pool))
	} else {
		linkSvc = apptglink.NewService(tglinkstore.NewMemoryStore())
	}

	// Bot user store (persists chatID → email mappings)
	var botUserStore tgbot.UserStore
	if pool != nil {
		botUserStore = botuser.New(pool)
	}

	// Telegram bot + notifiers
	// DirectNotifier removed: bot already confirms bookings directly to the user,
	// sending a DM via DirectNotifier would cause double notifications.
	var bookingSvc *appbooking.Service
	var groupNotifier *notifier.GroupNotifier
	if botToken != "" {
		api, apiErr := tgbot.NewBotAPI(botToken)
		if apiErr != nil {
			log.Printf("telegram bot API init error: %v — running without bot", apiErr)
			bookingSvc = appbooking.NewService(bookingRepo, roomSvc)
		} else {
			var parts []interface {
				NotifyNewBooking(context.Context, domainbooking.Booking) error
				NotifyDeletedBooking(context.Context, domainbooking.Booking) error
			}
			// Group chat / channel notifications only
			if chat := os.Getenv("TELEGRAM_CHAT_ID"); chat != "" {
				var setter func(context.Context, string, int)
				if pgRepo, ok := bookingRepo.(*pgrepo.BookingPostgresRepo); ok {
					setter = func(ctx context.Context, id string, msgID int) {
						if err := pgRepo.SetTgMsgID(ctx, id, msgID); err != nil {
							log.Printf("SetTgMsgID error: %v", err)
						}
					}
				}
				if gn := notifier.NewTelegramNotifier(botToken, chat, setter); gn != nil {
					parts = append(parts, gn)
					groupNotifier = gn
				}
			}

			if len(parts) > 0 {
				bookingSvc = appbooking.NewServiceWithNotifier(bookingRepo, roomSvc, notifier.NewCompositeNotifier(parts...))
			} else {
				bookingSvc = appbooking.NewService(bookingRepo, roomSvc)
			}

			tgbot.StartBookingBot(ctx, api, bookingSvc, roomSvc, authSvc, botUserStore)
		}
	} else {
		bookingSvc = appbooking.NewService(bookingRepo, roomSvc)
	}

	hub := server.NewHub()

	handler := server.NewRouter(server.Services{
		Booking:    bookingSvc,
		Auth:       authSvc,
		Admin:      adminSvc,
		Room:       roomSvc,
		Mailer:     ml,
		Hub:        hub,
		LinkSvc:    linkSvc,
		TGNotifier: groupNotifier,
	})

	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("HTTP server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("shutting down...")
	case err := <-errCh:
		log.Printf("server error: %v", err)
	}

	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(shutCtx)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
