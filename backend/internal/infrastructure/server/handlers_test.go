package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appbooking "Dormitory_Booking/internal/application/booking"
	"Dormitory_Booking/internal/infrastructure/memory"
	"Dormitory_Booking/internal/infrastructure/server"
)

func setupTestServer() http.Handler {
	repo := memory.NewInMemoryBookingRepo()
	svc := appbooking.NewService(repo)
	return server.NewRouter(svc)
}

func futureTimes() (string, string) {
	start := time.Now().Add(24 * time.Hour).Truncate(time.Minute)
	end := start.Add(1 * time.Hour)
	return start.Format(time.RFC3339), end.Format(time.RFC3339)
}

func TestCreateBooking(t *testing.T) {
	h := setupTestServer()

	start := time.Date(2099, 1, 5, 10, 0, 0, 0, time.UTC)
	end := start.Add(1 * time.Hour)

	body := map[string]any{
		"start":       start.Format(time.RFC3339),
		"end":         end.Format(time.RFC3339),
		"room":        21,
		"title":       "Test",
		"description": "desc",
		"telegramId":  "11",
		"isPrivate":   false,
	}

	raw, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/bookings", bytes.NewReader(raw))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("ожидали 200, получили %d, тело: %s", w.Code, w.Body.String())
	}
}

func TestListBookings(t *testing.T) {
	h := setupTestServer()

	req := httptest.NewRequest("GET", "/bookings", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("ожидали 200, получили %d", w.Code)
	}
}
