package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound = errors.New("record not found")
)

type User struct {
	ID           string    `json:"id"`
	CollegeEmail string    `json:"college_email"`
	Name         *string   `json:"name,omitempty"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type OTPCode struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CodeHash  string    `json:"code_hash"`
	ExpiresAt time.Time `json:"expires_at"`
	Consumed  bool      `json:"consumed"`
	CreatedAt time.Time `json:"created_at"`
}

type Repository interface {
	CreateOTP(ctx context.Context, email, codeHash string, expiresAt time.Time) (*OTPCode, error)
	GetLatestOTPByEmail(ctx context.Context, email string) (*OTPCode, error)
	MarkOTPConsumed(ctx context.Context, id string) error

	UpsertUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) CreateOTP(ctx context.Context, email, codeHash string, expiresAt time.Time) (*OTPCode, error) {
	query := `
		INSERT INTO otp_codes (email, code_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, email, code_hash, expires_at, consumed, created_at;
	`
	var otp OTPCode
	err := r.pool.QueryRow(ctx, query, email, codeHash, expiresAt).Scan(
		&otp.ID,
		&otp.Email,
		&otp.CodeHash,
		&otp.ExpiresAt,
		&otp.Consumed,
		&otp.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create otp: %w", err)
	}
	return &otp, nil
}

func (r *PostgresRepository) GetLatestOTPByEmail(ctx context.Context, email string) (*OTPCode, error) {
	query := `
		SELECT id, email, code_hash, expires_at, consumed, created_at
		FROM otp_codes
		WHERE email = $1
		ORDER BY created_at DESC
		LIMIT 1;
	`
	var otp OTPCode
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&otp.ID,
		&otp.Email,
		&otp.CodeHash,
		&otp.ExpiresAt,
		&otp.Consumed,
		&otp.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to fetch latest otp: %w", err)
	}
	return &otp, nil
}

func (r *PostgresRepository) MarkOTPConsumed(ctx context.Context, id string) error {
	query := `UPDATE otp_codes SET consumed = true WHERE id = $1;`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark otp consumed: %w", err)
	}
	return nil
}

func (r *PostgresRepository) UpsertUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		INSERT INTO users (college_email)
		VALUES ($1)
		ON CONFLICT (college_email) DO UPDATE SET college_email = EXCLUDED.college_email
		RETURNING id, college_email, name, role, created_at;
	`
	var user User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.CollegeEmail,
		&user.Name,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert user: %w", err)
	}
	return &user, nil
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT id, college_email, name, role, created_at
		FROM users
		WHERE id = $1;
	`
	var user User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.CollegeEmail,
		&user.Name,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to fetch user by id: %w", err)
	}
	return &user, nil
}
