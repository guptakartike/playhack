package service

import (
	"context"
	"errors"
	"time"

	"playhack/internal/repository"
)

var (
	ErrInvalidDateFormat = errors.New("invalid date format, expected YYYY-MM-DD")
)

type FacilityService struct {
	repo repository.Repository
}

func NewFacilityService(repo repository.Repository) *FacilityService {
	return &FacilityService{repo: repo}
}

func (s *FacilityService) ListFacilities(ctx context.Context) ([]repository.Facility, error) {
	return s.repo.ListFacilities(ctx)
}

func (s *FacilityService) ListCourtsByFacility(ctx context.Context, facilityID string) ([]repository.Court, error) {
	return s.repo.ListCourtsByFacility(ctx, facilityID)
}

func (s *FacilityService) ListSlotsByCourtAndDate(ctx context.Context, courtID string, dateStr string) ([]repository.SlotWithAvailability, error) {
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	} else {
		if _, err := time.Parse("2006-01-02", dateStr); err != nil {
			return nil, ErrInvalidDateFormat
		}
	}

	return s.repo.ListSlotsByCourtAndDate(ctx, courtID, dateStr)
}
