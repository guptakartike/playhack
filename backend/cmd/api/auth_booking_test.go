package main

import (
	"bytes"
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

func setupBookingTestApp(t *testing.T) (http.Handler, *service.AuthService, repository.Repository, *pgxpool.Pool, func()) {
	_ = godotenv.Load("../.env")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/playhack_db?sslmode=disable"
	}

	jwtSecret := "test-secret-key"

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
		mock := newBookingMockRepo()
		repo = mock
		cleanup = func() {}
	}

	authService := service.NewAuthService(repo, jwtSecret)
	bookingService := service.NewBookingService(repo, nil)
	facilityService := service.NewFacilityService(repo)

	bookingHandler := handler.NewBookingHandler(bookingService, nil)
	facilityHandler := handler.NewFacilityHandler(facilityService)
	authMiddleware := middleware.NewAuthMiddleware(authService)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /courts/{id}/slots", facilityHandler.HandleListSlots)

	mux.Handle("POST /bookings", authMiddleware.Authenticate(http.HandlerFunc(bookingHandler.HandleCreateBooking)))
	mux.Handle("GET /bookings/mine", authMiddleware.Authenticate(http.HandlerFunc(bookingHandler.HandleGetMyBookings)))
	mux.Handle("DELETE /bookings/{id}", authMiddleware.Authenticate(http.HandlerFunc(bookingHandler.HandleCancelBooking)))

	handlerWithCORS := middleware.CORS(mux)

	return handlerWithCORS, authService, repo, pool, cleanup
}

func TestBookingNormalPath(t *testing.T) {
	router, authService, repo, pool, cleanup := setupBookingTestApp(t)
	defer cleanup()

	ctx := context.Background()
	todayStr := time.Now().Format("2006-01-02")

	var user1ID, user1Token string
	var user2ID, user2Token string
	var testCourtID, testSlot1ID, testSlot2ID string
	var createdBookingID string

	if pool != nil {
		// Insert User 1
		err := pool.QueryRow(ctx, `INSERT INTO users (college_email) VALUES ('user1@iitg.ac.in') RETURNING id`).Scan(&user1ID)
		if err != nil {
			t.Fatalf("failed to insert user 1: %v", err)
		}
		code1, _ := authService.RequestOTP(ctx, "user1@iitg.ac.in")
		user1Token, _, _ = authService.VerifyOTP(ctx, "user1@iitg.ac.in", code1)

		// Insert User 2
		err = pool.QueryRow(ctx, `INSERT INTO users (college_email) VALUES ('user2@iitg.ac.in') RETURNING id`).Scan(&user2ID)
		if err != nil {
			t.Fatalf("failed to insert user 2: %v", err)
		}
		code2, _ := authService.RequestOTP(ctx, "user2@iitg.ac.in")
		user2Token, _, _ = authService.VerifyOTP(ctx, "user2@iitg.ac.in", code2)

		// Insert Facility, Court, and 2 Slots in future
		var facID string
		_ = pool.QueryRow(ctx, `INSERT INTO facilities (name, sport_type) VALUES ('Complex A', 'Badminton') RETURNING id`).Scan(&facID)
		_ = pool.QueryRow(ctx, `INSERT INTO courts (facility_id, label) VALUES ($1, 'Court 1') RETURNING id`, facID).Scan(&testCourtID)

		futureStart1 := time.Now().Add(2 * time.Hour)
		futureEnd1 := futureStart1.Add(1 * time.Hour)
		_ = pool.QueryRow(ctx, `INSERT INTO slots (court_id, start_time, end_time) VALUES ($1, $2, $3) RETURNING id`, testCourtID, futureStart1, futureEnd1).Scan(&testSlot1ID)

		futureStart2 := futureStart1.Add(2 * time.Hour)
		futureEnd2 := futureStart2.Add(1 * time.Hour)
		_ = pool.QueryRow(ctx, `INSERT INTO slots (court_id, start_time, end_time) VALUES ($1, $2, $3) RETURNING id`, testCourtID, futureStart2, futureEnd2).Scan(&testSlot2ID)
	} else if mock, ok := repo.(*bookingMockRepo); ok {
		testCourtID = "court-1"
		testSlot1ID = "slot-1"
		testSlot2ID = "slot-2"

		u1, _ := mock.UpsertUserByEmail(ctx, "user1@iitg.ac.in")
		user1ID = u1.ID
		code1, _ := authService.RequestOTP(ctx, "user1@iitg.ac.in")
		user1Token, _, _ = authService.VerifyOTP(ctx, "user1@iitg.ac.in", code1)

		u2, _ := mock.UpsertUserByEmail(ctx, "user2@iitg.ac.in")
		user2ID = u2.ID
		code2, _ := authService.RequestOTP(ctx, "user2@iitg.ac.in")
		user2Token, _, _ = authService.VerifyOTP(ctx, "user2@iitg.ac.in", code2)

		mock.slots[testSlot1ID] = &repository.SlotWithAvailability{
			ID: testSlot1ID, StartTime: time.Now().Add(2 * time.Hour), EndTime: time.Now().Add(3 * time.Hour), Available: true,
		}
		mock.slots[testSlot2ID] = &repository.SlotWithAvailability{
			ID: testSlot2ID, StartTime: time.Now().Add(4 * time.Hour), EndTime: time.Now().Add(5 * time.Hour), Available: true,
		}
	}

	t.Run("1. POST /bookings without a valid JWT -> 401", func(t *testing.T) {
		body := map[string]interface{}{"slot_id": testSlot1ID, "player_count": 2}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/bookings", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d. Body: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("2. POST /bookings with player_count 0 or negative -> 400", func(t *testing.T) {
		body := map[string]interface{}{"slot_id": testSlot1ID, "player_count": 0}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/bookings", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+user1Token)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d. Body: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("3. POST /bookings with valid slot -> 201", func(t *testing.T) {
		body := map[string]interface{}{"slot_id": testSlot1ID, "player_count": 2}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/bookings", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+user1Token)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected status 201 Created, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		var booking repository.Booking
		if err := json.Unmarshal(rr.Body.Bytes(), &booking); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if booking.ID == "" || booking.SlotID != testSlot1ID {
			t.Fatalf("invalid booking object returned: %v", booking)
		}
		createdBookingID = booking.ID
	})

	t.Run("4. POST /bookings on an already-booked slot -> 409", func(t *testing.T) {
		body := map[string]interface{}{"slot_id": testSlot1ID, "player_count": 4}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/bookings", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+user2Token)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusConflict {
			t.Errorf("expected status 409 Conflict, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		var resp map[string]string
		json.Unmarshal(rr.Body.Bytes(), &resp)
		if resp["error"] != "slot already booked" {
			t.Errorf("expected error 'slot already booked', got '%s'", resp["error"])
		}
	})

	t.Run("5. GET /bookings/mine returns only that user's bookings with joined info", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/bookings/mine", nil)
		req.Header.Set("Authorization", "Bearer "+user1Token)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		var bookings []repository.BookingDetail
		if err := json.Unmarshal(rr.Body.Bytes(), &bookings); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if len(bookings) != 1 {
			t.Fatalf("expected 1 booking for user 1, got %d", len(bookings))
		}

		if bookings[0].ID != createdBookingID {
			t.Errorf("expected booking ID %s, got %s", createdBookingID, bookings[0].ID)
		}
	})

	t.Run("6. DELETE /bookings/{id} by a different user -> 403", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/bookings/"+createdBookingID, nil)
		req.Header.Set("Authorization", "Bearer "+user2Token)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("expected status 403 Forbidden, got %d. Body: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("7. DELETE /bookings/{id} by the owner -> 200, slot becomes available again", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/bookings/"+createdBookingID, nil)
		req.Header.Set("Authorization", "Bearer "+user1Token)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		reqSlots := httptest.NewRequest("GET", fmt.Sprintf("/courts/%s/slots?date=%s", testCourtID, todayStr), nil)
		rrSlots := httptest.NewRecorder()

		router.ServeHTTP(rrSlots, reqSlots)

		if rrSlots.Code != http.StatusOK {
			t.Fatalf("expected status 200 for slots, got %d. Body: %s", rrSlots.Code, rrSlots.Body.String())
		}

		var slots []repository.SlotWithAvailability
		json.Unmarshal(rrSlots.Body.Bytes(), &slots)

		for _, s := range slots {
			if s.ID == testSlot1ID {
				if !s.Available {
					t.Errorf("expected cancelled slot %s to become available: true again", testSlot1ID)
				}
			}
		}
	})
}

// In-memory mock for booking repository tests
type bookingMockRepo struct {
	users    map[string]*repository.User
	otps     map[string][]*repository.OTPCode
	slots    map[string]*repository.SlotWithAvailability
	bookings map[string]*repository.Booking
}

func newBookingMockRepo() *bookingMockRepo {
	return &bookingMockRepo{
		users:    make(map[string]*repository.User),
		otps:     make(map[string][]*repository.OTPCode),
		slots:    make(map[string]*repository.SlotWithAvailability),
		bookings: make(map[string]*repository.Booking),
	}
}

func (m *bookingMockRepo) CreateOTP(ctx context.Context, email, codeHash string, expiresAt time.Time) (*repository.OTPCode, error) {
	otp := &repository.OTPCode{
		ID:        fmt.Sprintf("otp-%d", time.Now().UnixNano()),
		Email:     email,
		CodeHash:  codeHash,
		ExpiresAt: expiresAt,
		Consumed:  false,
		CreatedAt: time.Now(),
	}
	m.otps[email] = append(m.otps[email], otp)
	return otp, nil
}

func (m *bookingMockRepo) GetLatestOTPByEmail(ctx context.Context, email string) (*repository.OTPCode, error) {
	list, ok := m.otps[email]
	if !ok || len(list) == 0 {
		return nil, repository.ErrNotFound
	}
	return list[len(list)-1], nil
}

func (m *bookingMockRepo) MarkOTPConsumed(ctx context.Context, id string) error {
	for _, list := range m.otps {
		for _, o := range list {
			if o.ID == id {
				o.Consumed = true
				return nil
			}
		}
	}
	return nil
}

func (m *bookingMockRepo) UpsertUserByEmail(ctx context.Context, email string) (*repository.User, error) {
	for _, u := range m.users {
		if u.CollegeEmail == email {
			return u, nil
		}
	}
	u := &repository.User{ID: fmt.Sprintf("usr-%d", len(m.users)+1), CollegeEmail: email, Role: "student"}
	m.users[u.ID] = u
	return u, nil
}

func (m *bookingMockRepo) GetUserByID(ctx context.Context, id string) (*repository.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return u, nil
}

func (m *bookingMockRepo) ListFacilities(ctx context.Context) ([]repository.Facility, error) {
	return nil, nil
}
func (m *bookingMockRepo) ListCourtsByFacility(ctx context.Context, facilityID string) ([]repository.Court, error) {
	return nil, nil
}
func (m *bookingMockRepo) ListSlotsByCourtAndDate(ctx context.Context, courtID string, date string) ([]repository.SlotWithAvailability, error) {
	res := []repository.SlotWithAvailability{}
	for _, s := range m.slots {
		avail := true
		for _, b := range m.bookings {
			if b.SlotID == s.ID && b.Status == "confirmed" {
				avail = false
				break
			}
		}
		res = append(res, repository.SlotWithAvailability{
			ID:        s.ID,
			StartTime: s.StartTime,
			EndTime:   s.EndTime,
			Available: avail,
		})
	}
	return res, nil
}

func (m *bookingMockRepo) GetSlotByID(ctx context.Context, slotID string) (*repository.SlotWithAvailability, error) {
	s, ok := m.slots[slotID]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return s, nil
}

func (m *bookingMockRepo) CreateBooking(ctx context.Context, slotID, userID string, playerCount int) (*repository.Booking, error) {
	for _, b := range m.bookings {
		if b.SlotID == slotID && b.Status == "confirmed" {
			return nil, repository.ErrSlotAlreadyBooked
		}
	}

	b := &repository.Booking{
		ID:          fmt.Sprintf("book-%d", len(m.bookings)+1),
		SlotID:      slotID,
		UserID:      userID,
		PlayerCount: playerCount,
		Status:      "confirmed",
		CreatedAt:   time.Now(),
	}
	m.bookings[b.ID] = b
	return b, nil
}

func (m *bookingMockRepo) GetBookingsByUser(ctx context.Context, userID string) ([]repository.BookingDetail, error) {
	res := []repository.BookingDetail{}
	for _, b := range m.bookings {
		if b.UserID == userID && b.Status == "confirmed" {
			res = append(res, repository.BookingDetail{
				ID:           b.ID,
				SlotID:       b.SlotID,
				UserID:       b.UserID,
				FacilityName: "Complex A",
				SportType:    "Badminton",
				CourtLabel:   "Court 1",
				StartTime:    time.Now().Add(2 * time.Hour),
				EndTime:      time.Now().Add(3 * time.Hour),
				PlayerCount:  b.PlayerCount,
				Status:       b.Status,
				CreatedAt:    b.CreatedAt,
			})
		}
	}
	return res, nil
}

func (m *bookingMockRepo) CancelBooking(ctx context.Context, bookingID, userID string) error {
	b, ok := m.bookings[bookingID]
	if !ok {
		return repository.ErrBookingNotFound
	}
	if b.UserID != userID {
		return repository.ErrBookingNotOwned
	}
	b.Status = "cancelled"
	return nil
}

func (m *bookingMockRepo) CancelBookingAndNotifyWaitlist(ctx context.Context, bookingID, userID string) (string, []string, *repository.SlotNotificationPayload, error) {
	b, ok := m.bookings[bookingID]
	if !ok {
		return "", nil, nil, repository.ErrBookingNotFound
	}
	if b.UserID != userID {
		return "", nil, nil, repository.ErrBookingNotOwned
	}
	b.Status = "cancelled"
	payload := &repository.SlotNotificationPayload{
		SlotID:       b.SlotID,
		CourtLabel:   "Court 1",
		FacilityName: "Complex A",
		StartTime:    time.Now().Add(2 * time.Hour),
	}
	return b.SlotID, nil, payload, nil
}

func (m *bookingMockRepo) JoinWaitlist(ctx context.Context, slotID, userID string) (*repository.WaitlistEntry, error) {
	slot, ok := m.slots[slotID]
	if !ok {
		return nil, repository.ErrNotFound
	}
	if slot.Available {
		return nil, repository.ErrSlotAvailable
	}
	entry := &repository.WaitlistEntry{
		ID:        "waitlist-1",
		SlotID:    slotID,
		UserID:    userID,
		Status:    "waiting",
		CreatedAt: time.Now(),
	}
	return entry, nil
}
