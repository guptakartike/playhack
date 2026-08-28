package service

import (
	"context"
	"errors"
	"time"

	"playhack/internal/repository"
)

var (
	ErrInvalidPlayerCount = errors.New("player count must be greater than 0")
	ErrSlotInPast         = errors.New("cannot book a slot in the past")
)

type BookingService struct {
	repo repository.Repository
}

func NewBookingService(repo repository.Repository) *BookingService {
	return &BookingService{repo: repo}
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
	return s.repo.CancelBooking(ctx, bookingID, userID)
}
