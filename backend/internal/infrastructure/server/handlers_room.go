package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	roomsvc "Dormitory_Booking/internal/application/room"
	roomdomain "Dormitory_Booking/internal/domain/room"

	"github.com/go-chi/chi/v5"
)

type RoomHandlers struct{ svc *roomsvc.Service }

func NewRoomHandlers(svc *roomsvc.Service) *RoomHandlers { return &RoomHandlers{svc: svc} }

func (h *RoomHandlers) List(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.svc.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, rooms)
}

func (h *RoomHandlers) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Number int    `json:"number"`
		Name   string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	rm, err := h.svc.Create(r.Context(), roomdomain.Room{
		Number: body.Number,
		Name:   body.Name,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, rm)
}

func (h *RoomHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	num, err := strconv.Atoi(chi.URLParam(r, "number"))
	if err != nil {
		http.Error(w, "invalid room number", http.StatusBadRequest)
		return
	}
	if err := h.svc.Delete(r.Context(), num); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
