package server

import (
	"encoding/json"
	"net/http"
	"time"

	authsvc "Dormitory_Booking/internal/application/auth"
	adminsvc "Dormitory_Booking/internal/application/admin"
	admindomain "Dormitory_Booking/internal/domain/admin"
)

type AuthHandlers struct {
	auth  *authsvc.Service
	admin *adminsvc.Service
}

func NewAuthHandlers(auth *authsvc.Service, admin *adminsvc.Service) *AuthHandlers {
	return &AuthHandlers{auth: auth, admin: admin}
}

func (h *AuthHandlers) RequestOTP(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := h.auth.RequestOTP(r.Context(), body.Email); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandlers) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	token, u, err := h.auth.VerifyOTP(r.Context(), body.Email, body.Code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	isAdmin, _ := h.admin.IsAdmin(r.Context(), u.Email)

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})

	writeJSON(w, map[string]any{
		"email":        u.Email,
		"isAdmin":      isAdmin,
		"isSuperAdmin": admindomain.IsSuperAdmin(u.Email),
	})
}

func (h *AuthHandlers) Me(w http.ResponseWriter, r *http.Request) {
	u, ok := userFromCtx(r)
	if !ok {
		http.Error(w, "не авторизован", http.StatusUnauthorized)
		return
	}
	isAdmin := isAdminFromCtx(r)
	writeJSON(w, map[string]any{
		"email":        u.Email,
		"isAdmin":      isAdmin,
		"isSuperAdmin": admindomain.IsSuperAdmin(u.Email),
	})
}

func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		_ = h.auth.Logout(r.Context(), cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name: "session", Value: "", Path: "/", MaxAge: -1, HttpOnly: true,
	})
	w.WriteHeader(http.StatusOK)
}
