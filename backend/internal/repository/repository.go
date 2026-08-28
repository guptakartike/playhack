package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound          = errors.New("record not found")
	ErrSlotAlreadyBooked = errors.New("slot already booked")
	ErrBookingNotFound   = errors.New("booking not found")
	ErrBookingNotOwned    = errors.New("booking belongs to another user")
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

type Facility struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SportType string `json:"sport_type"`
}

type Court struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type SlotWithAvailability struct {
	ID        string    `json:"id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Available bool      `json:"available"`
}

type Booking struct {
	ID          string    `json:"id"`
	SlotID      string    `json:"slot_id"`
	UserID      string    `json:"user_id"`
	PlayerCount int       `json:"player_count"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type BookingDetail struct {
	ID           string    `json:"id"`
	SlotID       string    `json:"slot_id"`
	UserID       string    `json:"user_id"`
	FacilityName string    `json:"facility_name"`
	SportType    string    `json:"sport_type"`
	CourtLabel   string    `json:"court_label"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	PlayerCount  int       `json:"player_count"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type Repository interface {
	CreateOTP(ctx context.Context, email, codeHash string, expiresAt time.Time) (*OTPCode, error)
	GetLatestOTPByEmail(ctx context.Context, email string) (*OTPCode, error)
	MarkOTPConsumed(ctx context.Context, id string) error

	UpsertUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)

	ListFacilities(ctx context.Context) ([]Facility, error)
	ListCourtsByFacility(ctx context.Context, facilityID string) ([]Court, error)
	ListSlotsByCourtAndDate(ctx context.Context, courtID string, date string) ([]SlotWithAvailability, error)
	GetSlotByID(ctx context.Context, slotID string) (*SlotWithAvailability, error)

	CreateBooking(ctx context.Context, slotID, userID string, playerCount int) (*Booking, error)
	GetBookingsByUser(ctx context.Context, userID string) ([]BookingDetail, error)
	CancelBooking(ctx context.Context, bookingID, userID string) error
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

func (r *PostgresRepository) ListFacilities(ctx context.Context) ([]Facility, error) {
	query := `
		SELECT id, name, sport_type
		FROM facilities
		WHERE is_active = true
		ORDER BY name ASC;
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list facilities: %w", err)
	}
	defer rows.Close()

	facilities := []Facility{}
	for rows.Next() {
		var f Facility
		if err := rows.Scan(&f.ID, &f.Name, &f.SportType); err != nil {
			return nil, fmt.Errorf("failed to scan facility: %w", err)
		}
		facilities = append(facilities, f)
	}
	return facilities, nil
}

func (r *PostgresRepository) ListCourtsByFacility(ctx context.Context, facilityID string) ([]Court, error) {
	query := `
		SELECT id, label
		FROM courts
		WHERE facility_id = $1 AND is_active = true
		ORDER BY label ASC;
	`
	rows, err := r.pool.Query(ctx, query, facilityID)
	if err != nil {
		return nil, fmt.Errorf("failed to list courts: %w", err)
	}
	defer rows.Close()

	courts := []Court{}
	for rows.Next() {
		var c Court
		if err := rows.Scan(&c.ID, &c.Label); err != nil {
			return nil, fmt.Errorf("failed to scan court: %w", err)
		}
		courts = append(courts, c)
	}
	return courts, nil
}

func (r *PostgresRepository) ListSlotsByCourtAndDate(ctx context.Context, courtID string, date string) ([]SlotWithAvailability, error) {
	query := `
		SELECT 
			s.id, 
			s.start_time, 
			s.end_time, 
			(b.id IS NULL) AS available
		FROM slots s
		LEFT JOIN bookings b ON s.id = b.slot_id AND b.status = 'confirmed'
		WHERE s.court_id = $1 AND s.start_time::date = $2::date
		ORDER BY s.start_time ASC;
	`
	rows, err := r.pool.Query(ctx, query, courtID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to list slots: %w", err)
	}
	defer rows.Close()

	slots := []SlotWithAvailability{}
	for rows.Next() {
		var s SlotWithAvailability
		if err := rows.Scan(&s.ID, &s.StartTime, &s.EndTime, &s.Available); err != nil {
			return nil, fmt.Errorf("failed to scan slot: %w", err)
		}
		slots = append(slots, s)
	}
	return slots, nil
}

func (r *PostgresRepository) GetSlotByID(ctx context.Context, slotID string) (*SlotWithAvailability, error) {
	query := `
		SELECT s.id, s.start_time, s.end_time, (b.id IS NULL) AS available
		FROM slots s
		LEFT JOIN bookings b ON s.id = b.slot_id AND b.status = 'confirmed'
		WHERE s.id = $1;
	`
	var slot SlotWithAvailability
	err := r.pool.QueryRow(ctx, query, slotID).Scan(&slot.ID, &slot.StartTime, &slot.EndTime, &slot.Available)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to fetch slot: %w", err)
	}
	return &slot, nil
}

func (r *PostgresRepository) CreateBooking(ctx context.Context, slotID, userID string, playerCount int) (*Booking, error) {
	query := `
		INSERT INTO bookings (slot_id, user_id, player_count)
		VALUES ($1, $2, $3)
		RETURNING id, slot_id, user_id, player_count, status, created_at;
	`
	var b Booking
	err := r.pool.QueryRow(ctx, query, slotID, userID, playerCount).Scan(
		&b.ID, &b.SlotID, &b.UserID, &b.PlayerCount, &b.Status, &b.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrSlotAlreadyBooked
		}
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}
	return &b, nil
}

func (r *PostgresRepository) GetBookingsByUser(ctx context.Context, userID string) ([]BookingDetail, error) {
	query := `
		SELECT 
			b.id, b.slot_id, b.user_id, 
			f.name AS facility_name, f.sport_type, c.label AS court_label,
			s.start_time, s.end_time, b.player_count, b.status, b.created_at
		FROM bookings b
		JOIN slots s ON b.slot_id = s.id
		JOIN courts c ON s.court_id = c.id
		JOIN facilities f ON c.facility_id = f.id
		WHERE b.user_id = $1 AND b.status = 'confirmed'
		ORDER BY s.start_time DESC;
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user bookings: %w", err)
	}
	defer rows.Close()

	details := []BookingDetail{}
	for rows.Next() {
		var d BookingDetail
		if err := rows.Scan(
			&d.ID, &d.SlotID, &d.UserID,
			&d.FacilityName, &d.SportType, &d.CourtLabel,
			&d.StartTime, &d.EndTime, &d.PlayerCount, &d.Status, &d.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan booking detail: %w", err)
		}
		details = append(details, d)
	}
	return details, nil
}

func (r *PostgresRepository) CancelBooking(ctx context.Context, bookingID, userID string) error {
	var ownerID string
	var status string
	err := r.pool.QueryRow(ctx, `SELECT user_id, status FROM bookings WHERE id = $1`, bookingID).Scan(&ownerID, &status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrBookingNotFound
		}
		return fmt.Errorf("failed to fetch booking: %w", err)
	}

	if ownerID != userID {
		return ErrBookingNotOwned
	}

	_, err = r.pool.Exec(ctx, `UPDATE bookings SET status = 'cancelled' WHERE id = $1 AND user_id = $2`, bookingID, userID)
	if err != nil {
		return fmt.Errorf("failed to cancel booking: %w", err)
	}

	return nil
}
