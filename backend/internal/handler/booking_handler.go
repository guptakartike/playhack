package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"playhack/internal/middleware"
	"playhack/internal/repository"
	"playhack/internal/service"
)

type BookingHandler struct {
	bookingService *service.BookingService
	sseHub         *service.SSEHub
}

func NewBookingHandler(bookingService *service.BookingService, sseHub *service.SSEHub) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
		sseHub:         sseHub,
	}
}

type CreateBookingPayload struct {
	SlotID      string `json:"slot_id"`
	PlayerCount int    `json:"player_count"`
}

func (h *BookingHandler) HandleCreateBooking(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	var req CreateBookingPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SlotID == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	booking, err := h.bookingService.BookSlot(r.Context(), userID, req.SlotID, req.PlayerCount)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrSlotAlreadyBooked):
			writeJSON(w, http.StatusConflict, ErrorResponse{Error: "slot already booked"})
		case errors.Is(err, service.ErrInvalidPlayerCount),
			errors.Is(err, repository.ErrInvalidPlayerCount),
			errors.Is(err, repository.ErrExceedsMaxPlayers),
			errors.Is(err, service.ErrSlotInPast),
			errors.Is(err, repository.ErrNotFound):
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		default:
			log.Printf("Error creating booking for user %s on slot %s: %v", userID, req.SlotID, err)
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, booking)
}

func (h *BookingHandler) HandleGetMyBookings(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	bookings, err := h.bookingService.GetMyBookings(r.Context(), userID)
	if err != nil {
		log.Printf("Error fetching my bookings for user %s: %v", userID, err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, bookings)
}

func (h *BookingHandler) HandleCancelBooking(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	bookingID := r.PathValue("id")
	if bookingID == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "booking id required"})
		return
	}

	err := h.bookingService.CancelMyBooking(r.Context(), userID, bookingID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrBookingNotFound):
			writeJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case errors.Is(err, repository.ErrBookingNotOwned):
			writeJSON(w, http.StatusForbidden, ErrorResponse{Error: err.Error()})
		default:
			log.Printf("Error cancelling booking %s for user %s: %v", bookingID, userID, err)
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "booking cancelled successfully"})
}

func (h *BookingHandler) HandleJoinWaitlist(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	slotID := r.PathValue("id")
	if slotID == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "slot id required"})
		return
	}

	entry, err := h.bookingService.JoinWaitlist(r.Context(), userID, slotID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrAlreadyOnWaitlist):
			writeJSON(w, http.StatusConflict, ErrorResponse{Error: err.Error()})
		case errors.Is(err, repository.ErrSlotAvailable),
			errors.Is(err, service.ErrSlotInPast),
			errors.Is(err, repository.ErrNotFound):
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		default:
			log.Printf("Error joining waitlist for user %s on slot %s: %v", userID, slotID, err)
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, entry)
}

func (h *BookingHandler) HandleNotificationStream(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "streaming unsupported"})
		return
	}
	flusher.Flush()

	if h.sseHub == nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "notification hub uninitialized"})
		return
	}

	ch := make(chan service.NotificationPayload, 10)
	h.sseHub.Register(userID, ch)
	defer h.sseHub.Unregister(userID, ch)

	for {
		select {
		case <-r.Context().Done():
			return
		case payload, ok := <-ch:
			if !ok {
				return
			}
			data, err := json.Marshal(payload)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}
