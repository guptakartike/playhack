# Huddle Up — Architecture

## 1. System Overview

Huddle Up is a two-tier application: a Go backend exposing a REST + SSE API over PostgreSQL, and a React (TypeScript) single-page frontend. There is no separate queue, cache, or third-party notification service — all correctness guarantees are enforced by PostgreSQL itself, and real-time delivery is handled by a lightweight in-process SSE hub.

```
                  ┌─────────────────────┐
                  │   React (Vite/TS)   │
                  │   frontend/src/     │
                  └──────────┬──────────┘
                             │ REST (fetch) + SSE (EventSource)
                             ▼
                  ┌─────────────────────┐
                  │   Go net/http mux    │
                  │   cmd/api/main.go    │
                  └──────────┬──────────┘
             ┌───────────────┼────────────────┐
             ▼                ▼                ▼
      ┌─────────────┐  ┌─────────────┐  ┌──────────────┐
      │  handler/   │  │  middleware/│  │  service/     │
      │  (HTTP I/O) │  │  (auth,CORS)│  │  (business    │
      └──────┬──────┘  └─────────────┘  │   logic +     │
             │                          │   SSE hub)     │
             ▼                          └──────┬────────┘
      ┌─────────────┐                          │
      │ repository/ │◄─────────────────────────┘
      │ (pgx queries)│
      └──────┬──────┘
             ▼
      ┌─────────────┐
      │ PostgreSQL   │
      │ (pgxpool)    │
      └─────────────┘
```

## 2. Backend Layering

Strict four-layer separation, consistent across every feature added (auth, facilities, bookings, waitlist):

- **`handler/`** — thin HTTP layer. Decodes requests, calls into `service/`, maps domain errors to HTTP status codes, writes JSON responses via a shared `writeJSON` helper. No business logic or SQL here.
- **`service/`** — business logic and validation (e.g., email domain checks, OTP expiry, player-count limits, slot-availability checks before a waitlist join). Owns sentinel errors (e.g., `ErrInvalidPlayerCount`, `ErrSlotInPast`) and holds the SSE hub for pushing notifications.
- **`repository/`** — all database access via `pgx`/`pgxpool`. Owns the `Repository` interface, all SQL, and the translation of Postgres-specific errors (e.g., `23505` unique violations) into domain sentinel errors (`ErrSlotAlreadyBooked`, `ErrAlreadyOnWaitlist`).
- **`middleware/`** — cross-cutting concerns: JWT authentication (`auth_middleware.go`) and CORS (`cors.go`), wrapping the mux.

Routing is wired centrally in `cmd/api/main.go` using Go 1.22+'s method-aware `http.ServeMux` (`mux.HandleFunc("POST /bookings", ...)`), with protected routes wrapped individually in `authMiddleware.Authenticate(...)`.

## 3. Data Model

| Table | Purpose | Key constraint |
|---|---|---|
| `users` | Student/admin accounts, keyed by `college_email` | `UNIQUE(college_email)` |
| `otp_codes` | Login OTPs, hashed, time-limited | indexed by `email` |
| `facilities` | Sports facilities (e.g., Badminton Complex) | — |
| `courts` | Individual courts within a facility | `max_players` (added in migration 03) |
| `slots` | Bookable time windows per court | `UNIQUE(court_id, start_time)` |
| `bookings` | A student's booking of a slot | **partial unique index**: `UNIQUE(slot_id) WHERE status='confirmed'` — the core correctness guarantee |
| `waitlist_entries` | Students waiting on a currently-booked slot | `UNIQUE(slot_id, user_id)` — prevents duplicate joins |

Full schema lives in `backend/migrations/`.

## 4. Concurrency Model (the core architectural decision)

Both booking races and waitlist-join races are resolved the **same way**: by letting PostgreSQL's unique constraints be the single arbiter, rather than any application-level locking.

- **Booking:** `CreateBooking` performs a direct `INSERT` with no pre-check `SELECT`. The partial unique index on `bookings` guarantees only one `confirmed` row can exist per `slot_id`. A losing concurrent insert fails with Postgres error `23505`, which the repository translates to `ErrSlotAlreadyBooked` → HTTP `409`.
- **Waitlist join:** Same shape — direct `INSERT` into `waitlist_entries`, `UNIQUE(slot_id, user_id)` prevents duplicates, `23505` → `ErrAlreadyOnWaitlist` → HTTP `409`.
- **Cancellation → notify → re-race:** `CancelBooking` runs inside a single `pgx.Tx`: verify ownership → mark booking `cancelled` → mark all `waiting` waitlist rows for that slot as `notified`, committed atomically. Notified users are then pushed an SSE event and race for the now-open slot using the exact same booking-insert mechanism above.
- **Cleanup:** When a slot with `notified` waitlist rows is successfully re-booked, remaining `notified` rows are transitioned to `expired` in the same transaction as the new booking, so stale entries don't resurface.

No mutexes, goroutine-based locks, or other in-process synchronization are used for any of the above — deliberately, so the guarantees hold even if the API were later run as multiple instances behind a load balancer.

## 5. Real-Time Notification Path

```
CancelBooking (tx commits)
        │
        ▼
 fetch notified user_ids
        │
        ▼
SSEHub.BroadcastToUser(userID, payload)  ── per user_id
        │
        ▼
 in-memory map[user_id]map[chan]struct{}
        │
        ▼
GET /notifications/stream (open EventSource per logged-in client)
        │
        ▼
 Frontend NotificationContext.tsx
```

The SSE hub (`internal/service/sse_hub.go`) is the one piece of intentionally in-memory, non-durable state in the system — see `decisions.md` (D3) for why that's a safe trade-off here: it routes live pushes to currently-connected clients, but `waitlist_entries.status` in Postgres remains the actual source of truth, so nothing is lost if a client isn't connected when a notification fires.

## 6. Authentication

- Email/OTP flow scoped to `@iitg.ac.in` addresses.
- OTP hashed with bcrypt, 5-minute expiry, single-use (`consumed` flag).
- On successful verification, a 24-hour JWT (HS256) is issued and required as a `Bearer` token on all protected routes.
- `AuthMiddleware` extracts and validates the token, attaching `user_id`/`role` to the request context for downstream handlers.

## 7. Frontend

React + TypeScript, built with Vite, styled with Tailwind CSS v4. Structure:

- `pages/` — route-level screens: `LoginPage`, `BookCourtPage`, `CourtDetailsPage`, `SelectTimePage`, `CheckoutPage`, `MyBookingsPage`.
- `components/` — reusable UI: `SportCard`, `CourtCard`, `TimeSlot`, `BookingCard`, `StatusBadge`, `Header`, `BottomNav`.
- `context/` — `AuthContext` (token/session state) and `NotificationContext` (SSE connection + incoming waitlist notifications).
- `api/` — typed API client (`client.ts`, `types.ts`) wrapping REST calls to the Go backend.

Routing via `react-router-dom`. No server-side rendering — pure client-side SPA, served as a static build (`frontend/dist/`) in production.

## 8. Demo Tooling

- `cmd/demo/race_demo.go` — a standalone CLI (separate from the test suite) that auto-registers demo users, discovers a real open slot via the live API, and fires N simultaneous booking requests with visual pass/fail output and latency timers — built specifically to make the concurrency guarantee demonstrable live in front of judges, not just provable in unit tests.
- `cmd/seed/main.go` — seeds representative facilities/courts/slots for local development and demos.

## 9. Testing

- `cmd/api/auth_test.go`, `facility_test.go`, `auth_booking_test.go` — table-driven `httptest`-based unit/integration tests per module.
- `cmd/api/concurrency_test.go` — 20-goroutine synchronized-start race test against real Postgres, asserting exactly one `201` and the rest `409`.
- `cmd/api/waitlist_test.go` — tests for waitlist join, duplicate rejection, and notify-on-cancel behavior.