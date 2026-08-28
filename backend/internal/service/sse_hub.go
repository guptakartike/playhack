package service

import (
	"sync"
	"time"
)

type NotificationPayload struct {
	SlotID       string    `json:"slot_id"`
	CourtLabel   string    `json:"court_label"`
	FacilityName string    `json:"facility_name"`
	StartTime    time.Time `json:"start_time"`
	Message      string    `json:"message"`
}

type SSEHub struct {
	mu      sync.RWMutex
	clients map[string]map[chan NotificationPayload]struct{}
}

func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients: make(map[string]map[chan NotificationPayload]struct{}),
	}
}

func (h *SSEHub) Register(userID string, ch chan NotificationPayload) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[userID]; !ok {
		h.clients[userID] = make(map[chan NotificationPayload]struct{})
	}
	h.clients[userID][ch] = struct{}{}
}

func (h *SSEHub) Unregister(userID string, ch chan NotificationPayload) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if userClients, ok := h.clients[userID]; ok {
		delete(userClients, ch)
		if len(userClients) == 0 {
			delete(h.clients, userID)
		}
	}
}

func (h *SSEHub) BroadcastToUser(userID string, payload NotificationPayload) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if userClients, ok := h.clients[userID]; ok {
		for ch := range userClients {
			select {
			case ch <- payload:
			default:
				// If channel buffer is full, skip to avoid blocking broadcast
			}
		}
	}
}
