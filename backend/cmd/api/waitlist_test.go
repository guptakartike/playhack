package main

import (
	"bytes"
	"context"
	"encoding/json"
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

type waitlistTestMockRepo struct {
	users    map[string]*repository.User
	otps     map[string][]*repository.OTPCode
	slots    map[string]*repository.SlotWithAvailability
	bookings map[string]*repository.Booking
	waitlist map[string]map[string]*repository.WaitlistEntry // slotID -> userID -> entry
}

func newWaitlistTestMockRepo() *waitlistTestMockRepo {
	return &waitlistTestMockRepo{
		users:    make(map[string]*repository.User),
		otps:     make(map[string][]*repository.OTPCode),
		slots:    make(map[string]*repository.SlotWithAvailability),
		bookings: make(map[string]*repository.Booking),
		waitlist: make(map[string]map[string]*repository.WaitlistEntry),
	}
}

func (m *waitlistTestMockRepo) CreateOTP(ctx context.Context, email, codeHash string, expiresAt time.Time) (*repository.OTPCode, error) {
	otp := &repository.OTPCode{ID: "otp-1", Email: email, CodeHash: codeHash, ExpiresAt: expiresAt, Consumed: false, CreatedAt: time.Now()}
	m.otps[email] = append(m.otps[email], otp)
	return otp, nil
}

func (m *waitlistTestMockRepo) GetLatestOTPByEmail(ctx context.Context, email string) (*repository.OTPCode, error) {
	list := m.otps[email]
	if len(list) == 0 {
		return nil, repository.ErrNotFound
	}
	return list[len(list)-1], nil
}

func (m *waitlistTestMockRepo) MarkOTPConsumed(ctx context.Context, id string) error {
	for _, list := range m.otps {
		for _, o := range list {
			if o.ID == id {
				o.Consumed = true
			}
		}
	}
	return nil
}

func (m *waitlistTestMockRepo) UpsertUserByEmail(ctx context.Context, email string) (*repository.User, error) {
	for _, u := range m.users {
		if u.CollegeEmail == email {
			return u, nil
		}
	}
	u := &repository.User{ID: "user-" + email, CollegeEmail: email, Role: "student", CreatedAt: time.Now()}
	m.users[u.ID] = u
	return u, nil
}

func (m *waitlistTestMockRepo) GetUserByID(ctx context.Context, id string) (*repository.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return u, nil
}

func (m *waitlistTestMockRepo) ListFacilities(ctx context.Context) ([]repository.Facility, error) {
	return nil, nil
}

func (m *waitlistTestMockRepo) ListCourtsByFacility(ctx context.Context, facilityID string) ([]repository.Court, error) {
	return nil, nil
}

func (m *waitlistTestMockRepo) ListSlotsByCourtAndDate(ctx context.Context, courtID string, date string) ([]repository.SlotWithAvailability, error) {
	return nil, nil
}

func (m *waitlistTestMockRepo) GetBookingsByUser(ctx context.Context, userID string) ([]repository.BookingDetail, error) {
	return nil, nil
}

func (m *waitlistTestMockRepo) GetSlotByID(ctx context.Context, slotID string) (*repository.SlotWithAvailability, error) {
	s, ok := m.slots[slotID]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return s, nil
}

func (m *waitlistTestMockRepo) CreateBooking(ctx context.Context, slotID, userID string, playerCount int) (*repository.Booking, error) {
	if playerCount <= 0 {
		return nil, repository.ErrInvalidPlayerCount
	}
	if playerCount > 20 {
		return nil, repository.ErrExceedsMaxPlayers
	}
	for _, b := range m.bookings {
		if b.SlotID == slotID && b.Status == "confirmed" {
			return nil, repository.ErrSlotAlreadyBooked
		}
	}
	b := &repository.Booking{
		ID:          "booking-" + slotID,
		SlotID:      slotID,
		UserID:      userID,
		PlayerCount: playerCount,
		Status:      "confirmed",
		CreatedAt:   time.Now(),
	}
	m.bookings[b.ID] = b
	if s, ok := m.slots[slotID]; ok {
		s.Available = false
	}
	return b, nil
}

func (m *waitlistTestMockRepo) JoinWaitlist(ctx context.Context, slotID, userID string) (*repository.WaitlistEntry, error) {
	s, ok := m.slots[slotID]
	if !ok {
		return nil, repository.ErrNotFound
	}
	if s.Available {
		return nil, repository.ErrSlotAvailable
	}
	if m.waitlist[slotID] == nil {
		m.waitlist[slotID] = make(map[string]*repository.WaitlistEntry)
	}
	if _, exists := m.waitlist[slotID][userID]; exists {
		return nil, repository.ErrAlreadyOnWaitlist
	}
	entry := &repository.WaitlistEntry{
		ID:        "waitlist-entry-" + userID,
		SlotID:    slotID,
		UserID:    userID,
		Status:    "waiting",
		CreatedAt: time.Now(),
	}
	m.waitlist[slotID][userID] = entry
	return entry, nil
}

func (m *waitlistTestMockRepo) CancelBookingAndNotifyWaitlist(ctx context.Context, bookingID, userID string) (string, []string, *repository.SlotNotificationPayload, error) {
	b, ok := m.bookings[bookingID]
	if !ok {
		return "", nil, nil, repository.ErrBookingNotFound
	}
	if b.UserID != userID {
		return "", nil, nil, repository.ErrBookingNotOwned
	}
	b.Status = "cancelled"
	if s, ok := m.slots[b.SlotID]; ok {
		s.Available = true
	}

	notifiedUserIDs := []string{}
	if entries, ok := m.waitlist[b.SlotID]; ok {
		for uID, entry := range entries {
			if entry.Status == "waiting" {
				entry.Status = "notified"
				now := time.Now()
				entry.NotifiedAt = &now
				notifiedUserIDs = append(notifiedUserIDs, uID)
			}
		}
	}

	payload := &repository.SlotNotificationPayload{
		SlotID:       b.SlotID,
		CourtLabel:   "Court 1",
		FacilityName: "Badminton Complex",
		StartTime:    time.Now().Add(2 * time.Hour),
	}

	return b.SlotID, notifiedUserIDs, payload, nil
}

func (m *waitlistTestMockRepo) CancelBooking(ctx context.Context, bookingID, userID string) error {
	_, _, _, err := m.CancelBookingAndNotifyWaitlist(ctx, bookingID, userID)
	return err
}

func TestWaitlistAndNotificationEndpoints(t *testing.T) {
	_ = godotenv.Load("../.env")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/playhack_db?sslmode=disable"
	}
	jwtSecret := "test-waitlist-secret"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
		cleanup = func() { pool.Close() }
	} else {
		t.Log("PostgreSQL dev DB unreachable, running waitlist tests against mock repo")
		mock := newWaitlistTestMockRepo()
		repo = mock
		cleanup = func() {}
	}

	sseHub := service.NewSSEHub()
	authService := service.NewAuthService(repo, jwtSecret)
	bookingService := service.NewBookingService(repo, sseHub)
	bookingHandler := handler.NewBookingHandler(bookingService, sseHub)
	authMiddleware := middleware.NewAuthMiddleware(authService)

	mux := http.NewServeMux()
	mux.Handle("POST /bookings", authMiddleware.Authenticate(http.HandlerFunc(bookingHandler.HandleCreateBooking)))
	mux.Handle("DELETE /bookings/{id}", authMiddleware.Authenticate(http.HandlerFunc(bookingHandler.HandleCancelBooking)))
	mux.Handle("POST /slots/{id}/waitlist", authMiddleware.Authenticate(http.HandlerFunc(bookingHandler.HandleJoinWaitlist)))
	mux.Handle("GET /notifications/stream", authMiddleware.Authenticate(http.HandlerFunc(bookingHandler.HandleNotificationStream)))

	router := middleware.CORS(mux)

	// Prepare users
	code1, _ := authService.RequestOTP(ctx, "user1@iitg.ac.in")
	user1Token, _, _ := authService.VerifyOTP(ctx, "user1@iitg.ac.in", code1)

	code2, _ := authService.RequestOTP(ctx, "user2@iitg.ac.in")
	user2Token, _, _ := authService.VerifyOTP(ctx, "user2@iitg.ac.in", code2)

	// Setup slots
	var slotID1, slotID2 string
	if mock, ok := repo.(*waitlistTestMockRepo); ok {
		slotID1 = "slot-avail-1"
		mock.slots[slotID1] = &repository.SlotWithAvailability{
			ID: slotID1, StartTime: time.Now().Add(2 * time.Hour), EndTime: time.Now().Add(3 * time.Hour), Available: true,
		}

		slotID2 = "slot-booked-1"
		mock.slots[slotID2] = &repository.SlotWithAvailability{
			ID: slotID2, StartTime: time.Now().Add(4 * time.Hour), EndTime: time.Now().Add(5 * time.Hour), Available: false,
		}
		mock.bookings["booking-user1"] = &repository.Booking{
			ID: "booking-user1", SlotID: slotID2, UserID: "user-user1@iitg.ac.in", PlayerCount: 2, Status: "confirmed", CreatedAt: time.Now(),
		}
	}

	t.Run("1. POST /slots/{id}/waitlist on an available slot -> 400 Bad Request", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/slots/"+slotID1+"/waitlist", nil)
		req.Header.Set("Authorization", "Bearer "+user2Token)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d. Body: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("2. POST /slots/{id}/waitlist on a booked slot -> 201 Created", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/slots/"+slotID2+"/waitlist", nil)
		req.Header.Set("Authorization", "Bearer "+user2Token)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected status 201 Created, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		var entry repository.WaitlistEntry
		if err := json.Unmarshal(rr.Body.Bytes(), &entry); err != nil {
			t.Fatalf("failed to unmarshal waitlist entry: %v", err)
		}
		if entry.Status != "waiting" {
			t.Errorf("expected status 'waiting', got '%s'", entry.Status)
		}
	})

	t.Run("3. POST /slots/{id}/waitlist duplicate join -> 409 Conflict", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/slots/"+slotID2+"/waitlist", nil)
		req.Header.Set("Authorization", "Bearer "+user2Token)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusConflict {
			t.Errorf("expected status 409 Conflict, got %d. Body: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("4. POST /bookings with player_count > max_players -> 400 Bad Request", func(t *testing.T) {
		body := map[string]interface{}{"slot_id": slotID1, "player_count": 50}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/bookings", bytes.NewBuffer(jsonBody))
		req.Header.Set("Authorization", "Bearer "+user1Token)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status 400 Bad Request, got %d. Body: %s", rr.Code, rr.Body.String())
		}
	})

	_ = cleanup
}
