# Huddle Up — Engineering Decisions

This document records the significant technical decisions made while building Huddle Up, along with the reasoning and trade-offs, so the choices are explainable in a demo or judge Q&A rather than looking arbitrary.

---

## D1: Booking correctness is enforced at the database level, not in application code

**Decision:** A PostgreSQL **partial unique index** —
```sql
CREATE UNIQUE INDEX idx_unique_confirmed_booking ON bookings (slot_id) WHERE status = 'confirmed';
```
— is the sole mechanism preventing two confirmed bookings on the same slot. The application does **not** pre-check availability with a `SELECT` before inserting; it inserts directly and lets Postgres reject the second writer.

**Why:** Any application-level check-then-act pattern (`SELECT ... ; if available: INSERT`) has a race window between the check and the act. Two requests can both pass the check before either commits the insert. A database constraint closes that window entirely — Postgres itself guarantees uniqueness regardless of how many requests arrive at the same instant or how many server instances are running.

**Trade-off accepted:** The repository layer must translate Postgres error code `23505` (unique violation) into a domain-level `ErrSlotAlreadyBooked` rather than getting a clean "not available" answer up front. This is a small ergonomic cost for a much stronger correctness guarantee.

**Validated by:** `cmd/api/concurrency_test.go` — 20 simultaneous goroutines attempt the same booking; exactly 1 succeeds (`201`), 19 receive `409`, on every run.

---

## D2: Waitlist duplicate-prevention reuses the same pattern — a DB constraint, not a goroutine/channel lock

**Decision:** Duplicate waitlist joins are prevented by a `UNIQUE (slot_id, user_id)` constraint on `waitlist_entries`, following the exact same pattern as D1.

**Why this over an in-memory lock (e.g., a goroutine + channel, or a mutex-guarded map):** In-memory state lives inside a single running process. It:
- resets on every server restart,
- doesn't work correctly across multiple instances of the API (e.g., if the service is later scaled horizontally) — two different instances would each think they're the sole gatekeeper and could both accept the same duplicate.

A database constraint has neither problem: it's the single source of truth regardless of how many processes are talking to it. This was an explicit correction made mid-build after initially considering goroutine-based deduplication — the DB-constraint approach was chosen instead specifically because it doesn't introduce a second, weaker correctness mechanism alongside the one already trusted for bookings.

---

## D3: Waitlist notifications are delivered via Server-Sent Events (SSE), not polling or WebSockets

**Decision:** Real-time notification delivery uses SSE (`GET /notifications/stream`) with an in-memory per-user connection registry (`internal/service/sse_hub.go`).

**Why SSE over polling:** Polling (`GET /notifications` every N seconds) is simpler but gives a weak "real-time" story and adds unnecessary request volume for a mostly-idle channel.

**Why SSE over WebSockets:** This use case is one-directional — the server pushes notification events to the client; the client never needs to push data back through this channel. WebSockets would add full-duplex complexity (handshake upgrade, message framing both ways) for a capability that's never used. SSE is the narrower, more appropriate tool.

**Why the connection registry being in-memory is acceptable here (unlike D1/D2):** This state is purely about routing a *live* push to a *currently connected* client — it is not the source of truth for correctness. The durable state is `waitlist_entries.status` in Postgres. If a user isn't connected when a notification fires, they simply don't get the live push (they'd see the slot's true state next time they load the app) — no data is lost, and no double-booking risk is introduced by this being in-memory. This is the deliberate distinction from D1/D2, where in-memory state would have been a correctness bug; here it's a delivery-layer optimization.

---

## D4: Waitlist notification is "notify-all," not "notify-first-then-next"

**Decision:** When a waitlisted slot becomes available, **every** waiting user is notified simultaneously. Whoever successfully books first wins the same race described in D1; the rest see the slot as taken again on their next attempt.

**Why over notify-first-only-with-timeout:** A "notify one person, wait for a timeout, then notify the next" model is fairer in principle but requires an expiry/timeout mechanism and materially more state management for a hackathon-scope feature. Notify-all is simpler, reuses the existing booking-race mechanism with zero new machinery, and is easy to reason about and demo. The trade-off (a popular slot's waitlist might see several near-simultaneous rejections) was judged acceptable for this prototype's scope.

**Consequence:** Waitlist entries that don't win the re-booking race must be actively cleaned up (see D5) so they don't linger in a stale "notified" state.

---

## D5: Cancellation and re-booking correctness for waitlists is transaction-scoped

**Decision:**
- `CancelBooking` runs inside a single `pgx.Tx`: it verifies ownership, updates the booking to `cancelled`, and marks all `waiting` waitlist entries for that slot as `notified` — all in one transaction.
- When a slot with `notified` waitlist entries is successfully re-booked, those remaining entries are updated to `expired` in the same transaction as the new booking insert.

**Why:** Without transaction scoping, a partial failure (e.g., the cancellation succeeds but the waitlist-notify update fails) could leave the system in an inconsistent state — a cancelled booking with a waitlist that never got notified, or (worse) stale `notified` entries left behind after the slot is already re-booked, which would let cleared-out students receive misleading "still waiting" state on their next check.

---

## D6: Player count validation added after initial build

**Decision:** `player_count` is validated against a per-court `max_players` column (added via migration, default 20) rather than accepted unbounded.

**Why:** The original implementation only checked `player_count > 0`, allowing an unbounded value (e.g., booking a badminton court for 500 players) — flagged as a correctness/sanity gap during review and fixed by tying validation to actual court capacity.

---

## D7: OTP code is currently returned in the API response, not emailed — flagged as a demo-only shortcut

**Decision (current state, explicitly flagged as needing to change before any real deployment):** `POST /auth/request-otp` returns the generated OTP directly in the JSON response body, rather than sending it via email.

**Why it exists:** This removes the dependency on an email-sending service for a hackathon prototype, letting the login flow be demoed end-to-end without external infrastructure.

**Risk:** As implemented, this means anyone who knows a valid `@iitg.ac.in`-shaped email string can authenticate as that user without ever accessing the inbox — the domain check provides no real security by itself. This is acceptable for a local/demo environment only. Before any deployment beyond the hackathon demo, this must be gated behind an environment check (e.g., only include the code in the response when running in a `dev`/`demo` mode) or replaced with actual email delivery.

---

## D8: Secrets management — `.env` committed, flagged as a submission risk

**Observation (not yet fully resolved):** An early version of the repository included a `.env` file with a real database password and JWT signing secret, with no `.gitignore` entry for it. This was corrected — the repo now includes a `.gitignore` (ignoring `.env`, `bin/`, `vendor/`, OS files, per `docs/checkpoint.md` Checkpoint 2) and a `.env.example` with placeholder values — but the original secrets should be treated as compromised and rotated if they were ever pushed to a public remote before the fix.