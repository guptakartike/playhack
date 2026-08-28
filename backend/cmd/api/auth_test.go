package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"playhack/internal/handler"
	"playhack/internal/middleware"
	"playhack/internal/repository"
	"playhack/internal/service"
)

// mockRepo is an in-memory implementation of repository.Repository for unit tests
type mockRepo struct {
	mu    sync.Mutex
	users map[string]*repository.User    // by ID
	otps  map[string][]*repository.OTPCode // by email
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		users: make(map[string]*repository.User),
		otps:  make(map[string][]*repository.OTPCode),
	}
}

func (m *mockRepo) CreateOTP(ctx context.Context, email, codeHash string, expiresAt time.Time) (*repository.OTPCode, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

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

func (m *mockRepo) GetLatestOTPByEmail(ctx context.Context, email string) (*repository.OTPCode, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	list, ok := m.otps[email]
	if !ok || len(list) == 0 {
		return nil, repository.ErrNotFound
	}
	return list[len(list)-1], nil
}

func (m *mockRepo) MarkOTPConsumed(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

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
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, u := range m.users {
		if u.CollegeEmail == email {
			return u, nil
		}
	}

	user := &repository.User{
		ID:           fmt.Sprintf("usr-%d", time.Now().UnixNano()),
		CollegeEmail: email,
		Role:         "student",
		CreatedAt:    time.Now(),
	}
	m.users[user.ID] = user
	return user, nil
}

func (m *mockRepo) GetUserByID(ctx context.Context, id string) (*repository.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, ok := m.users[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return user, nil
}

func setupTestServer() (http.Handler, *service.AuthService, *mockRepo) {
	repo := newMockRepo()
	jwtSecret := "test-secret-key"
	authService := service.NewAuthService(repo, jwtSecret)
	authHandler := handler.NewAuthHandler(authService)
	authMiddleware := middleware.NewAuthMiddleware(authService)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/request-otp", authHandler.HandleRequestOTP)
	mux.HandleFunc("POST /auth/verify-otp", authHandler.HandleVerifyOTP)
	mux.Handle("GET /me", authMiddleware.Authenticate(http.HandlerFunc(authHandler.HandleGetMe)))

	return mux, authService, repo
}

func TestAuthEndpoints(t *testing.T) {
	router, authService, mockDB := setupTestServer()

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

	t.Run("3. verify-otp with wrong code -> 401", func(t *testing.T) {
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

	t.Run("4. verify-otp with correct code -> 200 + valid JWT returned", func(t *testing.T) {
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

		var resp map[string]string
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		token, ok := resp["token"]
		if !ok || token == "" {
			t.Fatalf("expected JWT token in response, got %v", resp)
		}

		// Validate JWT claims
		claims, err := authService.ValidateJWT(token)
		if err != nil {
			t.Fatalf("returned JWT token is invalid: %v", err)
		}
		if claims.Role != "student" {
			t.Errorf("expected role 'student', got '%s'", claims.Role)
		}

		validJWTToken = token
	})

	t.Run("5. verify-otp reusing an already-consumed code -> 401", func(t *testing.T) {
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

	t.Run("6. verify-otp with expired code -> 401", func(t *testing.T) {
		expiredEmail := "expired@iitg.ac.in"
		// Insert an expired OTP directly into mock DB
		mockDB.CreateOTP(context.Background(), expiredEmail, "$2a$10$abcdefghijklmnopqrstuu", time.Now().Add(-10*time.Minute))

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

	t.Run("7. GET /me with no token -> 401", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/me", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d. Body: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("8. GET /me with a valid token -> 200, correct user data", func(t *testing.T) {
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
}
