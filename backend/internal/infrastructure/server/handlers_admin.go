package server

import (
	"encoding/json"
	"net/http"

	adminsvc "Dormitory_Booking/internal/application/admin"
	admindomain "Dormitory_Booking/internal/domain/admin"
)

type AdminHandlers struct{ svc *adminsvc.Service }

func NewAdminHandlers(svc *adminsvc.Service) *AdminHandlers { return &AdminHandlers{svc: svc} }

func (h *AdminHandlers) List(w http.ResponseWriter, r *http.Request) {
	admins, err := h.svc.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, admins)
}

func (h *AdminHandlers) Add(w http.ResponseWriter, r *http.Request) {
	u, ok := userFromCtx(r)
	if !ok || !admindomain.IsSuperAdmin(u.Email) {
		http.Error(w, "только суперадмин может управлять администраторами", http.StatusForbidden)
		return
	}

	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := h.svc.Add(r.Context(), body.Email, u.Email); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *AdminHandlers) Remove(w http.ResponseWriter, r *http.Request) {
	u, ok := userFromCtx(r)
	if !ok || !admindomain.IsSuperAdmin(u.Email) {
		http.Error(w, "только суперадмин может управлять администраторами", http.StatusForbidden)
		return
	}

	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := h.svc.Remove(r.Context(), body.Email, u.Email); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
