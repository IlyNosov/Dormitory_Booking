package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"Dormitory_Booking/internal/domain/auth"
)

// AuthRequestOTP — POST /auth/otp/request
// Тело: { "email": "user@edu.hse.ru" }
func (h *Handlers) AuthRequestOTP(w http.ResponseWriter, r *http.Request) {
	if h.authSvc == nil {
		http.Error(w, "auth not configured", http.StatusServiceUnavailable)
		return
	}
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := h.authSvc.RequestOTP(r.Context(), strings.TrimSpace(body.Email)); err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidEmail):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, auth.ErrTooManyRequests):
			http.Error(w, err.Error(), http.StatusTooManyRequests)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
}

// AuthVerifyOTP — POST /auth/otp/verify
// Тело: { "email", "code", "telegramId" (опционально, для бота) }
// Если telegramId задан — ответ JSON без Set-Cookie (бот-режим).
// Иначе — Set-Cookie: session=TOKEN + JSON.
func (h *Handlers) AuthVerifyOTP(w http.ResponseWriter, r *http.Request) {
	if h.authSvc == nil {
		http.Error(w, "auth not configured", http.StatusServiceUnavailable)
		return
	}
	var body struct {
		Email      string `json:"email"`
		Code       string `json:"code"`
		TelegramID string `json:"telegramId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	sess, user, err := h.authSvc.VerifyOTP(r.Context(), body.Email, body.Code, body.TelegramID)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidOTP):
			http.Error(w, err.Error(), http.StatusUnauthorized)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Бот-режим: telegramId передан → только JSON, без куки.
	if strings.TrimSpace(body.TelegramID) == "" {
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    sess.Token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Expires:  sess.ExpiresAt,
		})
	}

	writeJSON(w, map[string]any{
		"email":        user.Email,
		"isAdmin":      user.IsAdmin,
		"isSuperAdmin": user.IsSuperAdmin,
	})
}

// AuthLogout — POST /auth/logout
func (h *Handlers) AuthLogout(w http.ResponseWriter, r *http.Request) {
	if h.authSvc != nil {
		if c, err := r.Cookie("session"); err == nil && c.Value != "" {
			_ = h.authSvc.Logout(r.Context(), c.Value)
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name: "session", Value: "", Path: "/", MaxAge: -1, HttpOnly: true,
	})
	w.WriteHeader(http.StatusOK)
}

// AuthMe — GET /auth/me — возвращает текущего пользователя.
func (h *Handlers) AuthMe(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromCtx(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	writeJSON(w, map[string]any{
		"email":        user.Email,
		"isAdmin":      user.IsAdmin,
		"isSuperAdmin": user.IsSuperAdmin,
	})
}
