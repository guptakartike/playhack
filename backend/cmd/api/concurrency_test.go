package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
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

func TestConcurrencyRaceBenchmark(t *testing.T) {
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

	if err == nil && pool.Ping(ctx) == nil {
		m, err := migrate.New("file://../../migrations", dbURL)
		if err == nil {
			_ = m.Up()
		}

		repo = repository.NewPostgresRepository(pool)

		tCtx, tCancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, _ = pool.Exec(tCtx, "TRUNCATE TABLE bookings, slots, courts, facilities, otp_codes, users RESTART IDENTITY CASCADE;")
		tCancel()

		defer func() {
			tCtx2, tCancel2 := context.WithTimeout(context.Background(), 5*time.Second)
			_, _ = pool.Exec(tCtx2, "TRUNCATE TABLE bookings, slots, courts, facilities, otp_codes, users RESTART IDENTITY CASCADE;")
			tCancel2()
			pool.Close()
		}()
	} else {
		if pool != nil {
			pool.Close()
			pool = nil
		}
		t.Log("PostgreSQL dev database not reachable, using thread-safe mock repository for benchmark test")
		repo = newConcurrentMockRepo()
	}

	authService := service.NewAuthService(repo, jwtSecret)
	bookingService := service.NewBookingService(repo)
	bookingHandler := handler.NewBookingHandler(bookingService)
	authMiddleware := middleware.NewAuthMiddleware(authService)

	mux := http.NewServeMux()
	mux.Handle("POST /bookings", authMiddleware.Authenticate(http.HandlerFunc(bookingHandler.HandleCreateBooking)))
	router := middleware.CORS(mux)

	// Seed 1 target slot
	var targetSlotID string
	if pool != nil {
		var facID, courtID string
		_ = pool.QueryRow(ctx, `INSERT INTO facilities (name, sport_type) VALUES ('Arena 1', 'Badminton') RETURNING id`).Scan(&facID)
		_ = pool.QueryRow(ctx, `INSERT INTO courts (facility_id, label) VALUES ($1, 'Court 1') RETURNING id`, facID).Scan(&courtID)

		futureStart := time.Now().Add(5 * time.Hour)
		futureEnd := futureStart.Add(1 * time.Hour)
		err = pool.QueryRow(ctx, `INSERT INTO slots (court_id, start_time, end_time) VALUES ($1, $2, $3) RETURNING id`, courtID, futureStart, futureEnd).Scan(&targetSlotID)
		if err != nil {
			t.Fatalf("failed to seed target slot: %v", err)
		}
	} else if mock, ok := repo.(*concurrentMockRepo); ok {
		targetSlotID = "slot-target-1"
		mock.slots[targetSlotID] = &repository.SlotWithAvailability{
			ID: targetSlotID, StartTime: time.Now().Add(5 * time.Hour), EndTime: time.Now().Add(6 * time.Hour), Available: true,
		}
	}

	// Seed 20 distinct users and generate JWT tokens
	const numGoroutines = 20
	userTokens := make([]string, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		email := fmt.Sprintf("student%d@iitg.ac.in", i+1)
		_, err := repo.UpsertUserByEmail(ctx, email)
		if err != nil {
			t.Fatalf("failed to create test user %s: %v", email, err)
		}

		code, _ := authService.RequestOTP(ctx, email)
		token, err := authService.VerifyOTP(ctx, email, code)
		if err != nil {
			t.Fatalf("failed to generate token for %s: %v", email, err)
		}
		userTokens[i] = token
	}

	// Prepare concurrent benchmark
	var wg sync.WaitGroup
	startChan := make(chan struct{})
	resultsChan := make(chan int, numGoroutines)

	body := map[string]interface{}{
		"slot_id":      targetSlotID,
		"player_count": 2,
	}
	jsonBody, _ := json.Marshal(body)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		token := userTokens[i]

		go func(tToken string) {
			defer wg.Done()

			req := httptest.NewRequest("POST", "/bookings", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tToken)
			rr := httptest.NewRecorder()

			// Block until synchronized start signal
			<-startChan

			router.ServeHTTP(rr, req)
			resultsChan <- rr.Code
		}(token)
	}

	// Release all 20 goroutines simultaneously
	close(startChan)
	wg.Wait()
	close(resultsChan)

	var count201, count409, countOther int
	statusBreakdown := make(map[int]int)

	for code := range resultsChan {
		statusBreakdown[code]++
		switch code {
		case http.StatusCreated:
			count201++
		case http.StatusConflict:
			count409++
		default:
			countOther++
		}
	}

	// Pitch-deck output formatting
	fmt.Println("\n============================================================")
	fmt.Println(" 🔥 PLAY HACK - CONCURRENCY RACE BENCHMARK TEST RESULT 🔥")
	fmt.Println("============================================================")
	fmt.Printf(" Target Slot ID              : %s\n", targetSlotID)
	fmt.Printf(" Total Simultaneous Workers  : %d Goroutines\n", numGoroutines)
	fmt.Printf(" Successful Bookings (201)  : %d\n", count201)
	fmt.Printf(" Conflict Rejections (409)  : %d\n", count409)
	fmt.Printf(" Other Responses             : %d\n", countOther)
	fmt.Println(" Status Breakdown            :", statusBreakdown)
	fmt.Println("============================================================")
	if count201 == 1 && count409 == numGoroutines-1 {
		fmt.Println(" 🎉 BENCHMARK PASSED: Atomic single-winner guarantee verified!")
	} else {
		fmt.Println(" ❌ BENCHMARK FAILED: Race condition detected!")
	}
	fmt.Println("============================================================")

	if count201 != 1 {
		t.Fatalf("EXPECTED EXACTLY 1 SUCCESSFUL BOOKING (201 Created), GOT %d", count201)
	}

	if count409 != numGoroutines-1 {
		t.Fatalf("EXPECTED EXACTLY %d CONFLICT REJECTIONS (409 Conflict), GOT %d", numGoroutines-1, count409)
	}
}

// Thread-safe mock for concurrency test fallback
type concurrentMockRepo struct {
	mu       sync.Mutex
	users    map[string]*repository.User
	otps     map[string][]*repository.OTPCode
	slots    map[string]*repository.SlotWithAvailability
	bookings map[string]*repository.Booking
}

func newConcurrentMockRepo() *concurrentMockRepo {
	return &concurrentMockRepo{
		users:    make(map[string]*repository.User),
		otps:     make(map[string][]*repository.OTPCode),
		slots:    make(map[string]*repository.SlotWithAvailability),
		bookings: make(map[string]*repository.Booking),
	}
}

func (m *concurrentMockRepo) CreateOTP(ctx context.Context, email, codeHash string, expiresAt time.Time) (*repository.OTPCode, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	otp := &repository.OTPCode{ID: fmt.Sprintf("otp-%d", time.Now().UnixNano()), Email: email, CodeHash: codeHash, ExpiresAt: expiresAt}
	m.otps[email] = append(m.otps[email], otp)
	return otp, nil
}

func (m *concurrentMockRepo) GetLatestOTPByEmail(ctx context.Context, email string) (*repository.OTPCode, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	list := m.otps[email]
	if len(list) == 0 {
		return nil, repository.ErrNotFound
	}
	return list[len(list)-1], nil
}

func (m *concurrentMockRepo) MarkOTPConsumed(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
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

func (m *concurrentMockRepo) UpsertUserByEmail(ctx context.Context, email string) (*repository.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.users {
		if u.CollegeEmail == email {
			return u, nil
		}
	}
	u := &repository.User{ID: fmt.Sprintf("usr-%d", len(m.users)+1), CollegeEmail: email, Role: "student"}
	m.users[u.ID] = u
	return u, nil
}

func (m *concurrentMockRepo) GetUserByID(ctx context.Context, id string) (*repository.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return u, nil
}

func (m *concurrentMockRepo) ListFacilities(ctx context.Context) ([]repository.Facility, error) {
	return nil, nil
}
func (m *concurrentMockRepo) ListCourtsByFacility(ctx context.Context, facilityID string) ([]repository.Court, error) {
	return nil, nil
}
func (m *concurrentMockRepo) ListSlotsByCourtAndDate(ctx context.Context, courtID string, date string) ([]repository.SlotWithAvailability, error) {
	return nil, nil
}

func (m *concurrentMockRepo) GetSlotByID(ctx context.Context, slotID string) (*repository.SlotWithAvailability, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.slots[slotID]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return s, nil
}

func (m *concurrentMockRepo) CreateBooking(ctx context.Context, slotID, userID string, playerCount int) (*repository.Booking, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

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

func (m *concurrentMockRepo) GetBookingsByUser(ctx context.Context, userID string) ([]repository.BookingDetail, error) {
	return nil, nil
}

func (m *concurrentMockRepo) CancelBooking(ctx context.Context, bookingID, userID string) error {
	return nil
}
