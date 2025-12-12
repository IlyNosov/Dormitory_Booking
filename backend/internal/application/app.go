package app

// В этом файле основная точка запуска backend-приложения.

import (
	appbooking "Dormitory_Booking/internal/application/booking"
	domainbooking "Dormitory_Booking/internal/domain/booking"
	bot "Dormitory_Booking/internal/infrastructure/booking_bot"
	"Dormitory_Booking/internal/infrastructure/memory"
	notifier "Dormitory_Booking/internal/infrastructure/notifier_bot"
	pgrepo "Dormitory_Booking/internal/infrastructure/postgres"
	"Dormitory_Booking/internal/infrastructure/server"
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
// Здесь настраиваем окружение, репозитории, сервисы и HTTP-сервер.
func Run(ctx context.Context) error {
	_ = godotenv.Load()

	addr := getEnv("HTTP_ADDR", ":8080")
	dbURL := os.Getenv("DB_URL") // если пусто — работаем в in-memory режиме
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	var repo domainbooking.Repository
	var pool *pgxpool.Pool
	var err error

	if dbURL != "" {
		log.Printf("using Postgres repo: %s\n", dbURL)
		pool, err = pgxpool.New(ctx, dbURL)
		if err != nil {
			return err
		}
		defer pool.Close()
		repo = pgrepo.NewBookingPostgresRepo(pool)
	} else {
		log.Println("DB_URL не задан, используем in-memory репозиторий (dev mode)")
		repo = memory.NewInMemoryBookingRepo()
	}

	var n appbooking.Notifier
	if botToken != "" {
		chat := os.Getenv("TELEGRAM_CHAT_ID")
		chatFile := os.Getenv("TELEGRAM_CHAT_ID_FILE")
		if chat == "" && chatFile != "" {
			if loaded, err := notifier.LoadChatID(chatFile); err == nil {
				chat = loaded
			}
		}
		if chat != "" {
			n = notifier.NewTelegramNotifier(botToken, chat)
		} else if chatFile != "" {
			go startTelegramPoller(context.Background(), botToken, chatFile)
		}
	}

	if n != nil {
		svc := appbooking.NewServiceWithNotifier(repo, n)
		// Запускаем бота для бронирований в личных сообщениях
		if botToken != "" {
			_ = bot.StartBookingBot(ctx, botToken, svc)
		}

		handler := server.NewRouter(svc)
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
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	}

	svc := appbooking.NewService(repo)
	if botToken != "" {
		_ = bot.StartBookingBot(ctx, botToken, svc)
	}
	handler := server.NewRouter(svc)

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

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	return nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func startTelegramPoller(ctx context.Context, token string, chatFile string) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Printf("telegram poller init error: %v", err)
		return
	}
	_, _ = bot.Request(tgbotapi.DeleteWebhookConfig{})

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	log.Printf("telegram poller started for @%s", bot.Self.UserName)
	for {
		select {
		case <-ctx.Done():
			log.Printf("telegram poller stopped: context done")
			return
		case upd := <-updates:
			if upd.Message != nil && upd.Message.Chat != nil {
				id := upd.Message.Chat.ID
				if err := notifier.SaveChatID(chatFile, formatInt64(id)); err == nil {
					log.Printf("saved chat id %d to %s via message", id, chatFile)
				}
				continue
			}
			if upd.MyChatMember != nil {
				id := upd.MyChatMember.Chat.ID
				if err := notifier.SaveChatID(chatFile, formatInt64(id)); err == nil {
					log.Printf("saved chat id %d to %s via member update", id, chatFile)
				}
				continue
			}
		}
	}
}

func formatInt64(v int64) string { return fmt.Sprintf("%d", v) }
