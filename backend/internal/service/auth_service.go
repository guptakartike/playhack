package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"playhack/internal/repository"
)

var (
	ErrInvalidEmailDomain = errors.New("email must end with @iitg.ac.in")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded, please wait 60 seconds")
	ErrOTPNotFound        = errors.New("invalid code")
	ErrOTPExpired         = errors.New("expired")
	ErrOTPAlreadyUsed     = errors.New("already used")
	ErrOTPInvalid         = errors.New("invalid code")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type AuthService struct {
	repo         repository.Repository
	jwtSecret    []byte
	rateLimitMu  sync.Mutex
	rateLimitMap map[string]time.Time
}

func NewAuthService(repo repository.Repository, jwtSecret string) *AuthService {
	return &AuthService{
		repo:         repo,
		jwtSecret:    []byte(jwtSecret),
		rateLimitMap: make(map[string]time.Time),
	}
}

func (s *AuthService) RequestOTP(ctx context.Context, email string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if !strings.HasSuffix(email, "@iitg.ac.in") {
		return "", ErrInvalidEmailDomain
	}

	// 60-second in-memory rate limiting check
	s.rateLimitMu.Lock()
	lastRequest, exists := s.rateLimitMap[email]
	if exists && time.Since(lastRequest) < 60*time.Second {
		s.rateLimitMu.Unlock()
		return "", ErrRateLimitExceeded
	}
	s.rateLimitMap[email] = time.Now()
	s.rateLimitMu.Unlock()

	code, err := generateNumericOTP(6)
	if err != nil {
		return "", fmt.Errorf("failed to generate otp: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash otp: %w", err)
	}

	expiresAt := time.Now().Add(5 * time.Minute)

	_, err = s.repo.CreateOTP(ctx, email, string(hash), expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to store otp: %w", err)
	}

	return code, nil
}

func (s *AuthService) VerifyOTP(ctx context.Context, email, code string) (string, *repository.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	code = strings.TrimSpace(code)

	otpRow, err := s.repo.GetLatestOTPByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", nil, ErrOTPNotFound
		}
		return "", nil, fmt.Errorf("error fetching otp: %w", err)
	}

	if otpRow.Consumed {
		return "", nil, ErrOTPAlreadyUsed
	}

	if time.Now().After(otpRow.ExpiresAt) {
		return "", nil, ErrOTPExpired
	}

	if err := bcrypt.CompareHashAndPassword([]byte(otpRow.CodeHash), []byte(code)); err != nil {
		return "", nil, ErrOTPInvalid
	}

	if err := s.repo.MarkOTPConsumed(ctx, otpRow.ID); err != nil {
		return "", nil, fmt.Errorf("failed to consume otp: %w", err)
	}

	user, err := s.repo.UpsertUserByEmail(ctx, email)
	if err != nil {
		return "", nil, fmt.Errorf("failed to upsert user: %w", err)
	}

	tokenStr, err := s.generateJWT(user.ID, user.Role)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate jwt: %w", err)
	}

	return tokenStr, user, nil
}

func (s *AuthService) ValidateJWT(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*repository.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

func (s *AuthService) generateJWT(userID, role string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func generateNumericOTP(length int) (string, error) {
	var code strings.Builder
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code.WriteString(num.String())
	}
	return code.String(), nil
}
