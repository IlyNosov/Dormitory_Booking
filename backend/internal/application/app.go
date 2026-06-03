package app

// В этом файле основная точка запуска backend-приложения.

import (
	appbooking "Dormitory_Booking/internal/application/booking"
	apptglink "Dormitory_Booking/internal/application/tglink"
	domainbooking "Dormitory_Booking/internal/domain/booking"
	tgbot "Dormitory_Booking/internal/infrastructure/booking_bot"
	"Dormitory_Booking/internal/infrastructure/memory"
	notifier "Dormitory_Booking/internal/infrastructure/notifier_bot"
	pgrepo "Dormitory_Booking/internal/infrastructure/postgres"
	"Dormitory_Booking/internal/infrastructure/server"
	tglinkstore "Dormitory_Booking/internal/infrastructure/tglink"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// Run — логика запуска backend-приложения.
func Run(ctx context.Context) error {
	_ = godotenv.Load()

	addr := getEnv("HTTP_ADDR", ":8080")
	dbURL := os.Getenv("DB_URL")
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	// ── Репозитории ──────────────────────────────────────────────────────────

	var bookingRepo domainbooking.Repository
	var pool *pgxpool.Pool
	var err error

	if dbURL != "" {
		log.Printf("using Postgres repo: %s\n", dbURL)
		pool, err = pgxpool.New(ctx, dbURL)
		if err != nil {
			return err
		}
		defer pool.Close()
		bookingRepo = pgrepo.NewBookingPostgresRepo(pool)
	} else {
		log.Println("DB_URL не задан, используем in-memory репозиторий (dev mode)")
		bookingRepo = memory.NewInMemoryBookingRepo()
	}

	// ── Link-сервис ───────────────────────────────────────────────────────────

	var linkSvc *apptglink.Service
	if pool != nil {
		linkSvc = apptglink.NewService(tglinkstore.NewPostgresStore(pool))
	} else {
		linkSvc = apptglink.NewService(tglinkstore.NewMemoryStore())
	}

	// ── Telegram: BotAPI → нотификаторы → сервис → booking-бот ───────────────
	//
	// Порядок инициализации важен: BotAPI создаётся один раз и используется
	// и для booking-бота (polling), и для DirectNotifier (отправка DM).

	var bookingSvc *appbooking.Service

	if botToken != "" {
		api, apiErr := tgbot.NewBotAPI(botToken)
		if apiErr != nil {
			log.Printf("telegram bot API init error: %v — работаем без бота", apiErr)
			bookingSvc = appbooking.NewService(bookingRepo)
		} else {
			var parts []interface {
				NotifyNewBooking(context.Context, domainbooking.Booking) error
				NotifyDeletedBooking(context.Context, domainbooking.Booking) error
			}

			// DM-нотификатор (личные сообщения владельцу брони)
			parts = append(parts, notifier.NewDirectNotifier(api))

			// Групповой нотификатор (чат общежития)
			chat := os.Getenv("TELEGRAM_CHAT_ID")
			chatFile := os.Getenv("TELEGRAM_CHAT_ID_FILE")
			if chat == "" && chatFile != "" {
				if loaded, loadErr := notifier.LoadChatID(chatFile); loadErr == nil {
					chat = loaded
				}
			}
			if chat != "" {
				if gn := notifier.NewTelegramNotifier(botToken, chat); gn != nil {
					parts = append(parts, gn)
				}
			} else if chatFile != "" {
				go startTelegramPoller(context.Background(), botToken, chatFile)
			}

			comp := notifier.NewCompositeNotifier(parts...)
			bookingSvc = appbooking.NewServiceWithNotifier(bookingRepo, comp)

			// Запускаем booking-бота на том же API
			tgbot.StartBookingBot(ctx, api, bookingSvc)
		}
	} else {
		bookingSvc = appbooking.NewService(bookingRepo)
	}

	// ── HTTP-сервер ───────────────────────────────────────────────────────────

	handler := server.NewRouter(bookingSvc, linkSvc)
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("HTTP server listening on %s\n", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("context cancelled, shutting down server...")
	case err := <-errCh:
		log.Printf("server error: %v\n", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func startTelegramPoller(ctx context.Context, token string, chatFile string) {
	b, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Printf("telegram poller init error: %v", err)
		return
	}
	_, _ = b.Request(tgbotapi.DeleteWebhookConfig{})

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.GetUpdatesChan(u)

	log.Printf("telegram poller started for @%s", b.Self.UserName)
	for {
		select {
		case <-ctx.Done():
			return
		case upd := <-updates:
			if upd.Message != nil && upd.Message.Chat != nil {
				id := upd.Message.Chat.ID
				if err := notifier.SaveChatID(chatFile, formatInt64(id)); err == nil {
					log.Printf("saved chat id %d to %s", id, chatFile)
				}
				continue
			}
			if upd.MyChatMember != nil {
				id := upd.MyChatMember.Chat.ID
				if err := notifier.SaveChatID(chatFile, formatInt64(id)); err == nil {
					log.Printf("saved chat id %d to %s", id, chatFile)
				}
				continue
			}
		}
	}
}

func formatInt64(v int64) string { return fmt.Sprintf("%d", v) }
