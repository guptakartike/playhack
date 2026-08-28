package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"playhack/internal/repository"
)

var (
	ErrInvalidPlayerCount = errors.New("player count must be greater than 0")
	ErrSlotInPast         = errors.New("cannot book a slot in the past")
)

type BookingService struct {
	repo   repository.Repository
	sseHub *SSEHub
}

func NewBookingService(repo repository.Repository, sseHub *SSEHub) *BookingService {
	return &BookingService{
		repo:   repo,
		sseHub: sseHub,
	}
}

func (s *BookingService) BookSlot(ctx context.Context, userID, slotID string, playerCount int) (*repository.Booking, error) {
	if playerCount <= 0 {
		return nil, ErrInvalidPlayerCount
	}

	slot, err := s.repo.GetSlotByID(ctx, slotID)
	if err != nil {
		return nil, err
	}

	if time.Now().After(slot.StartTime) {
		return nil, ErrSlotInPast
	}

	return s.repo.CreateBooking(ctx, slotID, userID, playerCount)
}

func (s *BookingService) GetMyBookings(ctx context.Context, userID string) ([]repository.BookingDetail, error) {
	return s.repo.GetBookingsByUser(ctx, userID)
}

func (s *BookingService) CancelMyBooking(ctx context.Context, userID, bookingID string) error {
	_, notifiedUserIDs, payload, err := s.repo.CancelBookingAndNotifyWaitlist(ctx, bookingID, userID)
	if err != nil {
		return err
	}

	if s.sseHub != nil && payload != nil && len(notifiedUserIDs) > 0 {
		notification := NotificationPayload{
			SlotID:       payload.SlotID,
			CourtLabel:   payload.CourtLabel,
			FacilityName: payload.FacilityName,
			StartTime:    payload.StartTime,
			Message:      fmt.Sprintf("Slot you're waiting on (%s - %s) is now open!", payload.FacilityName, payload.CourtLabel),
		}
		for _, uID := range notifiedUserIDs {
			s.sseHub.BroadcastToUser(uID, notification)
		}
	}

	return nil
}

func (s *BookingService) JoinWaitlist(ctx context.Context, userID, slotID string) (*repository.WaitlistEntry, error) {
	slot, err := s.repo.GetSlotByID(ctx, slotID)
	if err != nil {
		return nil, err
	}

	if time.Now().After(slot.StartTime) {
		return nil, ErrSlotInPast
	}

	return s.repo.JoinWaitlist(ctx, slotID, userID)
}
