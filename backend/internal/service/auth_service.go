package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"playhack/internal/repository"
)

var (
	ErrInvalidEmailDomain = errors.New("email must end with @iitg.ac.in")
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
	repo      repository.Repository
	jwtSecret []byte
}

func NewAuthService(repo repository.Repository, jwtSecret string) *AuthService {
	return &AuthService{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *AuthService) RequestOTP(ctx context.Context, email string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if !strings.HasSuffix(email, "@iitg.ac.in") {
		return "", ErrInvalidEmailDomain
	}

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

func (s *AuthService) VerifyOTP(ctx context.Context, email, code string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	code = strings.TrimSpace(code)

	otpRow, err := s.repo.GetLatestOTPByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", ErrOTPNotFound
		}
		return "", fmt.Errorf("error fetching otp: %w", err)
	}

	if otpRow.Consumed {
		return "", ErrOTPAlreadyUsed
	}

	if time.Now().After(otpRow.ExpiresAt) {
		return "", ErrOTPExpired
	}

	if err := bcrypt.CompareHashAndPassword([]byte(otpRow.CodeHash), []byte(code)); err != nil {
		return "", ErrOTPInvalid
	}

	if err := s.repo.MarkOTPConsumed(ctx, otpRow.ID); err != nil {
		return "", fmt.Errorf("failed to consume otp: %w", err)
	}

	user, err := s.repo.UpsertUserByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("failed to upsert user: %w", err)
	}

	tokenStr, err := s.generateJWT(user.ID, user.Role)
	if err != nil {
		return "", fmt.Errorf("failed to generate jwt: %w", err)
	}

	return tokenStr, nil
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
