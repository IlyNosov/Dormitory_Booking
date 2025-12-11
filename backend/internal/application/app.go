package app

// В этом файле основная точка запуска backend-приложения.

import (
	appbooking "Dormitory_Booking/internal/application/booking"
	domainbooking "Dormitory_Booking/internal/domain/booking"
	"Dormitory_Booking/internal/infrastructure/memory"
	pgrepo "Dormitory_Booking/internal/infrastructure/postgres"
	"Dormitory_Booking/internal/infrastructure/server"
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// Run — логика запуска backend-приложения.
// Здесь настраиваем окружение, репозитории, сервисы и HTTP-сервер.
func Run(ctx context.Context) error {
	_ = godotenv.Load()

	addr := getEnv("HTTP_ADDR", ":8080")
	dbURL := os.Getenv("DB_URL") // если пусто — работаем в in-memory режиме

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

	svc := appbooking.NewService(repo)
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
