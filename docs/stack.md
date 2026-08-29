# Huddle Up — Technology Stack

## Backend

| Layer | Choice | Notes |
|---|---|---|
| Language | Go 1.25.5 | |
| HTTP routing | `net/http` (stdlib), Go 1.22+ method-aware `ServeMux` | No third-party router — `mux.HandleFunc("POST /bookings", ...)` pattern matching is native to Go 1.22+ |
| Database | PostgreSQL | Source of truth for all correctness guarantees (unique constraints, partial indexes) |
| DB driver | `github.com/jackc/pgx/v5` + `pgxpool` | Connection pooling, native Postgres error code access (used to detect `23505` unique violations) |
| Migrations | `github.com/golang-migrate/migrate/v4` | Run automatically on server start against `migrations/` |
| Auth tokens | `github.com/golang-jwt/jwt/v5` | HS256, 24h expiry |
| Password/OTP hashing | `golang.org/x/crypto/bcrypt` | Used to hash OTP codes at rest, not passwords (system is passwordless/OTP-based) |
| Env config | `github.com/joho/godotenv` | Loads `.env` in local dev; falls back to real environment variables |
| Real-time delivery | Server-Sent Events (SSE), hand-rolled over `net/http` | `internal/service/sse_hub.go` — no external pub/sub or message broker |

## Frontend

| Layer | Choice | Notes |
|---|---|---|
| Framework | React 19 | |
| Language | TypeScript | |
| Build tool | Vite (rolldown-vite variant) | |
| Routing | `react-router-dom` v7 | |
| Styling | Tailwind CSS v4 (`@tailwindcss/vite`) | Dark theme, sporty-minimal aesthetic per design brief |
| Real-time client | Native `EventSource` API | Consumes the backend's `GET /notifications/stream` SSE endpoint |

## Infrastructure / Tooling

| Concern | Choice |
|---|---|
| Local dev orchestration | Manual (`go run`, `npm run dev`) — no Docker Compose currently checked in |
| Testing | Go stdlib `testing` + `net/http/httptest`, run against a real Postgres instance (not mocked) |
| Demo tooling | Standalone Go CLI (`cmd/demo/race_demo.go`) for live concurrency demonstrations |
| Version control | Git |

## Explicitly Not Used (and why)

- **No ORM** — raw SQL via `pgx` was chosen so that Postgres-specific behavior (partial unique indexes, specific error codes) is visible and controllable at the query level, which is central to the project's core correctness story.
- **No message queue / pub-sub system** (e.g., Redis, RabbitMQ) for notifications — SSE with an in-memory connection registry was judged sufficient for this scope; see `decisions.md` (D3) for the reasoning.
- **No WebSockets** — the notification channel is one-directional (server → client), so SSE was chosen as the narrower, simpler-to-reason-about tool.
- **No in-memory/mutex-based locking for correctness-critical paths** (booking, waitlist joins) — deliberately avoided in favor of database constraints; see `decisions.md` (D1, D2).