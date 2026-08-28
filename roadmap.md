# Play Hack - Build Order & System Roadmap

## Phase 1: Authentication & User Management (Completed)
**Goal:** Clean, domain-restricted college email login with OTP verification and JWT authentication.

- [x] **Database Schema & Migrations**
  - `users` table (`id` UUID, `college_email` UNIQUE, `role`)
  - `otp_codes` table (`id`, `email`, `code_hash`, `expires_at`, `consumed`)
- [x] **`POST /auth/request-otp`**
  - Validate `@iitg.ac.in` email domain
  - Generate secure 6-digit OTP code using `crypto/rand`
  - Hash with bcrypt and store with ~5 min expiry
  - In-memory rate limiting (60s cooldown per email)
- [x] **`POST /auth/verify-otp`**
  - Look up latest unconsumed code for email
  - Validate bcrypt hash and expiration
  - Mark code as consumed, upsert user record
  - Issue 24h JWT token
- [x] **JWT Middleware**
  - `Authorization: Bearer <token>` header parsing
  - Validate token & attach `user_id` and `role` to request context
  - Reject invalid/expired requests with HTTP `401 Unauthorized`
- [x] **Protected Verification Route (`GET /me`)**
  - Confirm middleware blocks unauthenticated requests and returns current user context

---

## Phase 2: Facilities, Courts & Slot Discovery (Completed)
**Goal:** Seed real campus sports data and expose discovery endpoints with computed slot availability.

- [x] **Database Schema & Migrations**
  - `facilities` table (`id`, `name`, `sport_type`, `is_active`)
  - `courts` table (`id`, `facility_id`, `label`, `is_active`)
  - `slots` table (`id`, `court_id`, `start_time`, `end_time`, unique on `(court_id, start_time)`)
  - `bookings` table (`id`, `slot_id` UNIQUE, `user_id`, `player_count`, `status`)
- [x] **Seeding Command (`cmd/seed/main.go`)**
  - Seed 3 facilities (Badminton Complex, Tennis Arena, Football Turf), 2 courts per facility, and hourly slots (8am to 9pm) for today and tomorrow
- [x] **`GET /facilities`** — List active sports facilities
- [x] **`GET /facilities/{id}/courts`** — List courts for a facility (returns `[]` for unknown UUIDs)
- [x] **`GET /courts/{id}/slots`** — List slots with date filter (`?date=YYYY-MM-DD`, defaults to today) and computed `available: bool` via `LEFT JOIN bookings`

---

## Phase 3: Core Booking Module & Concurrency Engine (Completed)
**Goal:** Guarantee single valid winner under heavy concurrent booking attempts.

- [x] **Partial Unique Index**
  - PostgreSQL partial unique index `idx_unique_confirmed_booking ON bookings (slot_id) WHERE status = 'confirmed'`
- [x] **`POST /bookings` (Critical Endpoint)**
  - Direct atomic `INSERT INTO bookings` relying on partial unique index (no SELECT pre-check)
  - Map PostgreSQL constraint error `23505` to `409 Conflict` ("slot already booked")
  - Validate player count > 0 and reject past slots with `400 Bad Request`
- [x] **`GET /bookings/mine`** — List authenticated user's active bookings joined with court/facility metadata
- [x] **`DELETE /bookings/{id}`** — Cancel authenticated user's booking (200 OK), releasing slot availability immediately (403 for unauthorized users)
- [x] **20-Goroutine Race Benchmark Test (`cmd/api/concurrency_test.go`)**
  - Spin up 20 parallel goroutines with distinct user JWT tokens firing simultaneously via a synchronized start channel
  - Assert exactly 1 request returns `201 Created` and 19 return `409 Conflict` with 100% pass consistency
