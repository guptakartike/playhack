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
	"golang.org/x/crypto/bcrypt"

	"playhack/internal/handler"
	"playhack/internal/middleware"
	"playhack/internal/repository"
	"playhack/internal/service"
)

func setupTestApp(t *testing.T) (http.Handler, *service.AuthService, repository.Repository, func()) {
	_ = godotenv.Load("../.env")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/playhack_db?sslmode=disable"
	}

	jwtSecret := "test-secret-key"

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

		truncateDB := func() {
			tCtx, tCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer tCancel()
			_, _ = pool.Exec(tCtx, "TRUNCATE TABLE users, otp_codes RESTART IDENTITY CASCADE;")
		}

		truncateDB()

		cleanup = func() {
			truncateDB()
			pool.Close()
		}
	} else {
		t.Log("PostgreSQL dev database not reachable, running test against in-memory repository fallback")
		mock := newMockRepo()
		repo = mock
		cleanup = func() {}
	}

	authService := service.NewAuthService(repo, jwtSecret)
	authHandler := handler.NewAuthHandler(authService)
	authMiddleware := middleware.NewAuthMiddleware(authService)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/request-otp", authHandler.HandleRequestOTP)
	mux.HandleFunc("POST /auth/verify-otp", authHandler.HandleVerifyOTP)
	mux.Handle("GET /me", authMiddleware.Authenticate(http.HandlerFunc(authHandler.HandleGetMe)))

	handlerWithCORS := middleware.CORS(mux)

	return handlerWithCORS, authService, repo, cleanup
}

func TestAuthEndpoints(t *testing.T) {
	router, authService, repo, cleanup := setupTestApp(t)
	defer cleanup()

	var validOTPCode string
	var validJWTToken string

	t.Run("1. request-otp with a non-college email -> 400", func(t *testing.T) {
		body := map[string]string{"email": "student@gmail.com"}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/auth/request-otp", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d. Body: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("2. request-otp with a valid college email -> 200", func(t *testing.T) {
		body := map[string]string{"email": "test@iitg.ac.in"}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/auth/request-otp", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		var resp map[string]string
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		code, ok := resp["code"]
		if !ok || code == "" {
			t.Fatalf("expected plaintext code in response, got %v", resp)
		}
		validOTPCode = code
	})

	t.Run("3. request-otp rate limiting within 60s -> 429", func(t *testing.T) {
		body := map[string]string{"email": "test@iitg.ac.in"}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/auth/request-otp", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusTooManyRequests {
			t.Errorf("expected status 429 Too Many Requests, got %d. Body: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("4. verify-otp with wrong code -> 401", func(t *testing.T) {
		body := map[string]string{
			"email": "test@iitg.ac.in",
			"code":  "000000",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/auth/verify-otp", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		var resp map[string]string
		json.Unmarshal(rr.Body.Bytes(), &resp)
		if resp["error"] != "invalid code" {
			t.Errorf("expected error message 'invalid code', got '%s'", resp["error"])
		}
	})

	t.Run("5. verify-otp with correct code -> 200 + valid JWT returned", func(t *testing.T) {
		body := map[string]string{
			"email": "test@iitg.ac.in",
			"code":  validOTPCode,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/auth/verify-otp", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		tokenRaw, ok := resp["token"]
		token, _ := tokenRaw.(string)
		if !ok || token == "" {
			t.Fatalf("expected JWT token in response, got %v", resp)
		}

		claims, err := authService.ValidateJWT(token)
		if err != nil {
			t.Fatalf("returned JWT token is invalid: %v", err)
		}
		if claims.Role != "student" {
			t.Errorf("expected role 'student', got '%s'", claims.Role)
		}

		validJWTToken = token
	})

	t.Run("6. verify-otp reusing an already-consumed code -> 401", func(t *testing.T) {
		body := map[string]string{
			"email": "test@iitg.ac.in",
			"code":  validOTPCode,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/auth/verify-otp", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		var resp map[string]string
		json.Unmarshal(rr.Body.Bytes(), &resp)
		if resp["error"] != "already used" {
			t.Errorf("expected error message 'already used', got '%s'", resp["error"])
		}
	})

	t.Run("7. verify-otp with expired code -> 401", func(t *testing.T) {
		expiredEmail := "expired@iitg.ac.in"
		hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
		_, _ = repo.CreateOTP(context.Background(), expiredEmail, string(hash), time.Now().Add(-10*time.Minute))

		body := map[string]string{
			"email": expiredEmail,
			"code":  "123456",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/auth/verify-otp", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		var resp map[string]string
		json.Unmarshal(rr.Body.Bytes(), &resp)
		if resp["error"] != "expired" {
			t.Errorf("expected error message 'expired', got '%s'", resp["error"])
		}
	})

	t.Run("8. GET /me with no token -> 401", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/me", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d. Body: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("9. GET /me with a valid token -> 200, correct user data", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/me", nil)
		req.Header.Set("Authorization", "Bearer "+validJWTToken)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		var user repository.User
		if err := json.Unmarshal(rr.Body.Bytes(), &user); err != nil {
			t.Fatalf("failed to decode user response: %v", err)
		}

		if user.CollegeEmail != "test@iitg.ac.in" {
			t.Errorf("expected email 'test@iitg.ac.in', got '%s'", user.CollegeEmail)
		}
		if user.Role != "student" {
			t.Errorf("expected role 'student', got '%s'", user.Role)
		}
	})

	t.Run("10. OPTIONS preflight request -> 204 No Content with CORS headers", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/me", nil)
		req.Header.Set("Origin", "http://localhost:5173")
		req.Header.Set("Access-Control-Request-Method", "GET")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Errorf("expected status 204 No Content, got %d", rr.Code)
		}

		if rr.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
			t.Errorf("expected CORS origin header, got '%s'", rr.Header().Get("Access-Control-Allow-Origin"))
		}
	})
}

// In-memory fallback mock for offline execution
type mockRepo struct {
	users map[string]*repository.User
	otps  map[string][]*repository.OTPCode
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		users: make(map[string]*repository.User),
		otps:  make(map[string][]*repository.OTPCode),
	}
}

func (m *mockRepo) CreateOTP(ctx context.Context, email, codeHash string, expiresAt time.Time) (*repository.OTPCode, error) {
	otp := &repository.OTPCode{
		ID:        "otp-123",
		Email:     email,
		CodeHash:  codeHash,
		ExpiresAt: expiresAt,
		Consumed:  false,
		CreatedAt: time.Now(),
	}
	m.otps[email] = append(m.otps[email], otp)
	return otp, nil
}

func (m *mockRepo) GetLatestOTPByEmail(ctx context.Context, email string) (*repository.OTPCode, error) {
	list, ok := m.otps[email]
	if !ok || len(list) == 0 {
		return nil, repository.ErrNotFound
	}
	return list[len(list)-1], nil
}

func (m *mockRepo) MarkOTPConsumed(ctx context.Context, id string) error {
	for _, list := range m.otps {
		for _, otp := range list {
			if otp.ID == id {
				otp.Consumed = true
				return nil
			}
		}
	}
	return repository.ErrNotFound
}

func (m *mockRepo) UpsertUserByEmail(ctx context.Context, email string) (*repository.User, error) {
	for _, u := range m.users {
		if u.CollegeEmail == email {
			return u, nil
		}
	}

	user := &repository.User{
		ID:           "user-123",
		CollegeEmail: email,
		Role:         "student",
		CreatedAt:    time.Now(),
	}
	m.users[user.ID] = user
	return user, nil
}

func (m *mockRepo) GetUserByID(ctx context.Context, id string) (*repository.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return user, nil
}

func (m *mockRepo) ListFacilities(ctx context.Context) ([]repository.Facility, error) {
	return nil, nil
}
func (m *mockRepo) ListCourtsByFacility(ctx context.Context, facilityID string) ([]repository.Court, error) {
	return nil, nil
}
func (m *mockRepo) ListSlotsByCourtAndDate(ctx context.Context, courtID string, date string) ([]repository.SlotWithAvailability, error) {
	return nil, nil
}
func (m *mockRepo) GetSlotByID(ctx context.Context, slotID string) (*repository.SlotWithAvailability, error) {
	return nil, nil
}
func (m *mockRepo) CreateBooking(ctx context.Context, slotID, userID string, playerCount int) (*repository.Booking, error) {
	return nil, nil
}
func (m *mockRepo) GetBookingsByUser(ctx context.Context, userID string) ([]repository.BookingDetail, error) {
	return nil, nil
}
func (m *mockRepo) CancelBooking(ctx context.Context, bookingID, userID string) error {
	return nil
}
