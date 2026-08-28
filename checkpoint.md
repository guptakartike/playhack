# Project Checkpoints

## Checkpoint 1: Auth Module Implementation
- **Schema & Migrations**: Created PostgreSQL migrations (`migrations/`) for `users` (UUID PK, college_email, role) and `otp_codes` (email, code_hash, expires_at, consumed).
- **Repository Layer**: Implemented DB persistence layer (`internal/repository`) using `pgxpool` for query execution and upserts.
- **Service Layer**: Implemented business logic (`internal/service`) for `@iitg.ac.in` email domain validation, secure 6-digit OTP generation (`crypto/rand`), bcrypt hashing with 5-min expiry, and 24h JWT issuance (`golang-jwt/jwt/v5`).
- **Middleware**: Built JWT authentication middleware (`internal/middleware`) extracting bearer tokens and attaching `user_id` and `role` to context via unexported `contextKey`.
- **Handlers & Routing**: Created thin HTTP handlers (`internal/handler`) and registered Go 1.22+ `http.ServeMux` method routes (`cmd/api/main.go`) for `POST /auth/request-otp`, `POST /auth/verify-otp`, and `GET /me`.
- **Testing**: Added table-driven unit test suite (`cmd/api/auth_test.go`) using `httptest` covering all 8 specified scenarios with 100% pass rate.

## Checkpoint 2: Auth Enhancements, CORS & Rate Limiting
- **Rate Limiting**: Added a 60-second in-memory rate limiter (`sync.Mutex` + `map[string]time.Time`) on `POST /auth/request-otp` returning 429 status code for repeated requests.
- **pgcrypto Extension**: Updated `000001_create_auth_tables.up.sql` to include `CREATE EXTENSION IF NOT EXISTS pgcrypto;` for seamless PostgreSQL `gen_random_uuid()` execution.
- **CORS Middleware**: Created `internal/middleware/cors.go` handling preflight `OPTIONS` requests (204 No Content) with configurable `FRONTEND_ORIGIN` env variable.
- **Layer Cleanliness**: Ensured `HandleGetMe` delegates strictly through service (`authService.GetUserByID`) to repository (`repo.GetUserByID`).
- **PostgreSQL Test Harness**: Configured test suite (`cmd/api/auth_test.go`) to execute against the Postgres dev DB with automatic table truncation (`TRUNCATE TABLE users, otp_codes RESTART IDENTITY CASCADE`).
- **Git Hygiene**: Added `.gitignore` ignoring `.env`, `bin/`, `vendor/`, and OS-specific files.

## Checkpoint 3: Facilities, Courts & Slot Availability Module
- **Schema & Migrations**: Created PostgreSQL migrations (`migrations/000002_create_facility_slots_tables.up.sql`) for `facilities`, `courts`, `slots` (with unique `(court_id, start_time)` constraint), and `bookings` tables.
- **Seeding Command**: Built database seed script (`cmd/seed/main.go`) populating 3 facilities, 2 courts per facility, and hourly slots (8am-9pm) for today and tomorrow.
- **Repository Layer**: Added `ListFacilities`, `ListCourtsByFacility`, and `ListSlotsByCourtAndDate` queries using `LEFT JOIN bookings` to compute slot availability (`available: bool`).
- **Service Layer**: Implemented `FacilityService` (`internal/service/facility_service.go`) validating date parameters (`YYYY-MM-DD`) and defaulting to today's date when unspecified.
- **Handlers & Routing**: Exposed public GET routes (`GET /facilities`, `GET /facilities/{id}/courts`, `GET /courts/{id}/slots`) using Go 1.22+ `http.ServeMux` method routing.
- **Testing**: Added table-driven test suite (`cmd/api/facility_test.go`) covering all 5 facility/slot test cases (seeded lists, non-existent facility UUID returning `[]`, slot availability computation with confirmed booking, malformed date returning 400).

## Checkpoint 4: Core Booking Module & Atomic Concurrency Engine
- **Partial Unique Index**: Added `CREATE UNIQUE INDEX idx_unique_confirmed_booking ON bookings (slot_id) WHERE status = 'confirmed';` allowing cancelled slots to be re-booked while enforcing single confirmed booking per slot.
- **Repository Layer**: Built `CreateBooking`, `GetBookingsByUser`, and `CancelBooking`. Directly mapped PostgreSQL `23505` constraint violations to sentinel error `ErrSlotAlreadyBooked`.
- **Service & Handler Layers**: Implemented `BookingService` and `BookingHandler` with JWT protection for `POST /bookings` (201 Created / 409 Conflict), `GET /bookings/mine`, and `DELETE /bookings/{id}` (200 OK / 403 Forbidden).
- **Normal-Path Tests**: Created `cmd/api/auth_booking_test.go` verifying valid bookings, 409 conflicts, invalid player counts, unauthorized requests, user booking isolation, and cancelled slot re-availability.
- **20-Goroutine Race Benchmark**: Created `cmd/api/concurrency_test.go` executing 20 simultaneous booking requests using a synchronized start channel. Verified 100% consistent atomic guarantees (1x HTTP 201, 19x HTTP 409).
