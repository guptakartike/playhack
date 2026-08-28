package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"playhack/internal/handler"
	"playhack/internal/middleware"
	"playhack/internal/repository"
	"playhack/internal/service"
)

func setupFacilityTestApp(t *testing.T) (http.Handler, repository.Repository, *pgxpool.Pool, func()) {
	_ = godotenv.Load("../.env")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/playhack_db?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	var repo repository.Repository
	var cleanup func()

	if err == nil && pool.Ping(ctx) == nil {
		m, err := migrate.New("file://../../migrations", dbURL)
		if err == nil {
			_ = m.Up()
		}

		repo = repository.NewPostgresRepository(pool)

		truncateDB := func() {
			tCtx, tCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer tCancel()
			_, _ = pool.Exec(tCtx, "TRUNCATE TABLE bookings, slots, courts, facilities, otp_codes, users RESTART IDENTITY CASCADE;")
		}

		truncateDB()

		cleanup = func() {
			truncateDB()
			pool.Close()
		}
	} else {
		if pool != nil {
			pool.Close()
			pool = nil
		}
		t.Log("PostgreSQL dev database not reachable, using in-memory mock repository")
		mock := newFacilityMockRepo()
		repo = mock
		cleanup = func() {}
	}

	facilityService := service.NewFacilityService(repo)
	facilityHandler := handler.NewFacilityHandler(facilityService)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /facilities", facilityHandler.HandleListFacilities)
	mux.HandleFunc("GET /facilities/{id}/courts", facilityHandler.HandleListCourts)
	mux.HandleFunc("GET /courts/{id}/slots", facilityHandler.HandleListSlots)

	handlerWithCORS := middleware.CORS(mux)

	return handlerWithCORS, repo, pool, cleanup
}

func TestFacilityEndpoints(t *testing.T) {
	router, repo, pool, cleanup := setupFacilityTestApp(t)
	defer cleanup()

	ctx := context.Background()
	todayStr := time.Now().Format("2006-01-02")

	var testFacilityID string
	var testCourtID string
	var testSlotID string

	if pool != nil {
		// Insert test facility
		err := pool.QueryRow(ctx, `INSERT INTO facilities (name, sport_type) VALUES ('Test Badminton Complex', 'Badminton') RETURNING id`).Scan(&testFacilityID)
		if err != nil {
			t.Fatalf("failed to seed test facility: %v", err)
		}

		// Insert test court
		err = pool.QueryRow(ctx, `INSERT INTO courts (facility_id, label) VALUES ($1, 'Court 1') RETURNING id`, testFacilityID).Scan(&testCourtID)
		if err != nil {
			t.Fatalf("failed to seed test court: %v", err)
		}

		// Insert 2 test slots for today
		startTime1 := time.Now().Truncate(24 * time.Hour).Add(10 * time.Hour)
		endTime1 := startTime1.Add(1 * time.Hour)
		err = pool.QueryRow(ctx, `INSERT INTO slots (court_id, start_time, end_time) VALUES ($1, $2, $3) RETURNING id`, testCourtID, startTime1, endTime1).Scan(&testSlotID)
		if err != nil {
			t.Fatalf("failed to seed test slot 1: %v", err)
		}

		startTime2 := startTime1.Add(1 * time.Hour)
		endTime2 := startTime2.Add(1 * time.Hour)
		_, err = pool.Exec(ctx, `INSERT INTO slots (court_id, start_time, end_time) VALUES ($1, $2, $3)`, testCourtID, startTime2, endTime2)
		if err != nil {
			t.Fatalf("failed to seed test slot 2: %v", err)
		}
	} else if mock, ok := repo.(*facilityMockRepo); ok {
		testFacilityID = "fac-1"
		testCourtID = "court-1"
		testSlotID = "slot-1"
		mock.facilities = []repository.Facility{{ID: testFacilityID, Name: "Test Badminton Complex", SportType: "Badminton"}}
		mock.courts[testFacilityID] = []repository.Court{{ID: testCourtID, Label: "Court 1"}}
		mock.slots[testCourtID] = []repository.SlotWithAvailability{
			{ID: testSlotID, StartTime: time.Now().Truncate(24 * time.Hour).Add(10 * time.Hour), EndTime: time.Now().Truncate(24 * time.Hour).Add(11 * time.Hour), Available: true},
			{ID: "slot-2", StartTime: time.Now().Truncate(24 * time.Hour).Add(11 * time.Hour), EndTime: time.Now().Truncate(24 * time.Hour).Add(12 * time.Hour), Available: true},
		}
	}

	t.Run("1. GET /facilities returns the seeded list", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/facilities", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		var facilities []repository.Facility
		if err := json.Unmarshal(rr.Body.Bytes(), &facilities); err != nil {
			t.Fatalf("failed to parse facilities response: %v", err)
		}

		if len(facilities) == 0 {
			t.Errorf("expected at least 1 facility in response")
		}
	})

	t.Run("2. GET /facilities/{bad-id}/courts returns an empty array (not 404)", func(t *testing.T) {
		badUUID := "00000000-0000-0000-0000-000000000000"
		req := httptest.NewRequest("GET", "/facilities/"+badUUID+"/courts", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		var courts []repository.Court
		if err := json.Unmarshal(rr.Body.Bytes(), &courts); err != nil {
			t.Fatalf("failed to parse courts response: %v", err)
		}

		if len(courts) != 0 {
			t.Errorf("expected 0 courts for non-existent facility, got %d", len(courts))
		}
	})

	t.Run("3. GET /courts/{id}/slots with no bookings -> all slots have available: true", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/courts/%s/slots?date=%s", testCourtID, todayStr), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		var slots []repository.SlotWithAvailability
		if err := json.Unmarshal(rr.Body.Bytes(), &slots); err != nil {
			t.Fatalf("failed to parse slots response: %v", err)
		}

		if len(slots) < 2 {
			t.Fatalf("expected at least 2 slots, got %d", len(slots))
		}

		for _, s := range slots {
			if !s.Available {
				t.Errorf("expected slot %s to be available: true", s.ID)
			}
		}
	})

	t.Run("4. After manually inserting a confirmed booking, that slot shows available: false", func(t *testing.T) {
		if pool != nil {
			var testUserID string
			err := pool.QueryRow(ctx, `INSERT INTO users (college_email) VALUES ('player@iitg.ac.in') RETURNING id`).Scan(&testUserID)
			if err != nil {
				t.Fatalf("failed to insert user: %v", err)
			}

			_, err = pool.Exec(ctx, `INSERT INTO bookings (slot_id, user_id, status) VALUES ($1, $2, 'confirmed')`, testSlotID, testUserID)
			if err != nil {
				t.Fatalf("failed to insert booking: %v", err)
			}
		} else if mock, ok := repo.(*facilityMockRepo); ok {
			mock.bookedSlots[testSlotID] = true
		}

		req := httptest.NewRequest("GET", fmt.Sprintf("/courts/%s/slots?date=%s", testCourtID, todayStr), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		var slots []repository.SlotWithAvailability
		if err := json.Unmarshal(rr.Body.Bytes(), &slots); err != nil {
			t.Fatalf("failed to parse slots response: %v", err)
		}

		var bookedSlotFound bool
		var unbookedSlotFound bool

		for _, s := range slots {
			if s.ID == testSlotID {
				if s.Available {
					t.Errorf("expected booked slot %s to have available: false", s.ID)
				}
				bookedSlotFound = true
			} else {
				if !s.Available {
					t.Errorf("expected unbooked slot %s to remain available: true", s.ID)
				}
				unbookedSlotFound = true
			}
		}

		if !bookedSlotFound {
			t.Errorf("booked slot %s was not returned in slots list", testSlotID)
		}
		if !unbookedSlotFound {
			t.Errorf("unbooked slots were not returned in slots list")
		}
	})

	t.Run("5. GET /courts/{id}/slots?date=not-a-date -> 400", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/courts/%s/slots?date=not-a-date", testCourtID), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status 400 Bad Request, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		var resp map[string]string
		json.Unmarshal(rr.Body.Bytes(), &resp)
		if resp["error"] != "invalid date format, expected YYYY-MM-DD" {
			t.Errorf("expected error message 'invalid date format, expected YYYY-MM-DD', got '%s'", resp["error"])
		}
	})
}

// In-memory fallback mock for facility repository
type facilityMockRepo struct {
	facilities  []repository.Facility
	courts      map[string][]repository.Court
	slots       map[string][]repository.SlotWithAvailability
	bookedSlots map[string]bool
}

func newFacilityMockRepo() *facilityMockRepo {
	return &facilityMockRepo{
		facilities:  []repository.Facility{},
		courts:      make(map[string][]repository.Court),
		slots:       make(map[string][]repository.SlotWithAvailability),
		bookedSlots: make(map[string]bool),
	}
}

func (m *facilityMockRepo) CreateOTP(ctx context.Context, email, codeHash string, expiresAt time.Time) (*repository.OTPCode, error) {
	return nil, nil
}
func (m *facilityMockRepo) GetLatestOTPByEmail(ctx context.Context, email string) (*repository.OTPCode, error) {
	return nil, nil
}
func (m *facilityMockRepo) MarkOTPConsumed(ctx context.Context, id string) error { return nil }
func (m *facilityMockRepo) UpsertUserByEmail(ctx context.Context, email string) (*repository.User, error) {
	return nil, nil
}
func (m *facilityMockRepo) GetUserByID(ctx context.Context, id string) (*repository.User, error) {
	return nil, nil
}

func (m *facilityMockRepo) ListFacilities(ctx context.Context) ([]repository.Facility, error) {
	return m.facilities, nil
}

func (m *facilityMockRepo) ListCourtsByFacility(ctx context.Context, facilityID string) ([]repository.Court, error) {
	courts, ok := m.courts[facilityID]
	if !ok {
		return []repository.Court{}, nil
	}
	return courts, nil
}

func (m *facilityMockRepo) ListSlotsByCourtAndDate(ctx context.Context, courtID string, date string) ([]repository.SlotWithAvailability, error) {
	slots, ok := m.slots[courtID]
	if !ok {
		return []repository.SlotWithAvailability{}, nil
	}

	result := make([]repository.SlotWithAvailability, len(slots))
	for i, s := range slots {
		avail := !m.bookedSlots[s.ID]
		result[i] = repository.SlotWithAvailability{
			ID:        s.ID,
			StartTime: s.StartTime,
			EndTime:   s.EndTime,
			Available: avail,
		}
	}
	return result, nil
}

func (m *facilityMockRepo) GetSlotByID(ctx context.Context, slotID string) (*repository.SlotWithAvailability, error) {
	return nil, nil
}
func (m *facilityMockRepo) CreateBooking(ctx context.Context, slotID, userID string, playerCount int) (*repository.Booking, error) {
	return nil, nil
}
func (m *facilityMockRepo) GetBookingsByUser(ctx context.Context, userID string) ([]repository.BookingDetail, error) {
	return nil, nil
}
func (m *facilityMockRepo) CancelBooking(ctx context.Context, bookingID, userID string) error {
	return nil
}
func (m *facilityMockRepo) CancelBookingAndNotifyWaitlist(ctx context.Context, bookingID, userID string) (string, []string, *repository.SlotNotificationPayload, error) {
	return "", nil, nil, nil
}
func (m *facilityMockRepo) JoinWaitlist(ctx context.Context, slotID, userID string) (*repository.WaitlistEntry, error) {
	return nil, nil
}
