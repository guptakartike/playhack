# Huddle Up

Concurrency-safe sports facility booking for IIT Guwahati students.

Built for the IITG Sports Board x Tech Board hackathon (SDE track). Huddle Up's central engineering focus is **correctness under contention**: when multiple students try to book the same court slot at the same instant, the system guarantees exactly one of them succeeds — enforced by PostgreSQL itself, not application-level locking.

## Docs

- [`docs/prd.md`](./docs/prd.md) — product requirements and scope
- [`docs/architecture.md`](./docs/architecture.md) — system design, layering, data model
- [`docs/decisions.md`](./docs/decisions.md) — key engineering decisions and why they were made
- [`docs/stack.md`](./docs/stack.md) — full technology stack
- [`docs/flow.md`](./docs/flow.md) — user and system flow diagrams
- [`docs/checkpoint.md`](./docs/checkpoint.md) — build history by checkpoint
- [`docs/roadmap.md`](./docs/roadmap.md) — phased build plan

## What's built

- **Auth** — OTP-based login restricted to `@iitg.ac.in` emails, JWT sessions.
- **Facility/court/slot discovery** with computed real-time availability.
- **Concurrency-safe booking** — a PostgreSQL partial unique index guarantees only one confirmed booking per slot, no matter how many requests race for it. Verified by a 20-goroutine automated race test and a live demo CLI.
- **Cancellation** — frees the slot immediately.
- **Waitlist + real-time notifications** — join a waitlist on an already-booked slot; when it's cancelled, everyone waiting is notified live via Server-Sent Events and races for the reopened slot using the same booking-safety mechanism.
- **React + TypeScript frontend** — dark, sporty-minimal UI covering the full booking flow and a "My Bookings" dashboard.

See `docs/prd.md` for what's explicitly **not** built yet (admin analytics, check-in/photo verification).

## Project structure

```
playhack/
├── backend/
│   ├── cmd/
│   │   ├── api/       # HTTP server entrypoint + tests
│   │   ├── demo/       # standalone live concurrency demo CLI
│   │   └── seed/       # DB seed script
│   ├── internal/
│   │   ├── handler/    # HTTP request/response layer
│   │   ├── service/    # business logic + SSE hub
│   │   ├── repository/ # all DB access (pgx)
│   │   └── middleware/ # JWT auth, CORS
│   ├── migrations/     # SQL migrations (golang-migrate)
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── pages/
│   │   ├── components/
│   │   ├── context/     # Auth + Notification (SSE) context
│   │   └── api/         # typed API client
│   └── package.json
└── docs/
```

## Prerequisites

- Go 1.25+
- Node.js (for the frontend, matching the `vite`/React 19 toolchain in `frontend/package.json`)
- PostgreSQL running locally (or reachable via `DATABASE_URL`)

## Setup

### 1. Environment

Copy the example env file and fill in real values:

```bash
cp .env.example backend/.env
```

Edit `backend/.env`:

```
PORT=5532
DATABASE_URL=postgres://user:password@localhost:5432/playhack_db?sslmode=disable
JWT_SECRET=<generate a strong random secret — do not reuse the example>
FRONTEND_ORIGIN=http://localhost:5533
```

**Never commit `backend/.env`** — it's already listed in `.gitignore`. Only `.env.example` (with placeholder values) should be committed.

### 2. Backend

```bash
cd backend
go mod download
go run cmd/api/main.go
```

This will connect to Postgres, run all pending migrations automatically, and start listening on `PORT` (default `5532`).

Optionally seed sample data (facilities, courts, slots):

```bash
go run cmd/seed/main.go
```

### 3. Frontend

```bash
cd frontend
npm install
npm run dev
```

The dev server runs on port `5533` and proxies `/api` requests to the backend on `5532` (see `vite.config.ts`).

## Running tests

```bash
cd backend
go test ./cmd/api/...
```

This includes the automated concurrency race test (`concurrency_test.go`) and waitlist tests (`waitlist_test.go`), run against a real Postgres instance — not mocked.

## Live concurrency demo

For a fast, visual demonstration of the core correctness guarantee (built specifically for judge walkthroughs):

```bash
cd backend
go run cmd/demo/race_demo.go
```

This auto-registers demo users, discovers a real open slot, fires simultaneous booking requests against it, and prints a color-coded pass/fail result with latency timings. Pass `--reset` to re-run instantly without a full database reset.
