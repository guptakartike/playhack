# Huddle Up — Flows

## 1. Authentication Flow

```
Student enters @iitg.ac.in email
        │
        ▼
POST /auth/request-otp
        │  ├─ reject if domain ≠ @iitg.ac.in       → 400
        │  ├─ reject if requested < 60s ago         → 429
        │  └─ generate 6-digit OTP, bcrypt-hash it,
        │     store with 5-min expiry
        ▼
Student enters the OTP
        │
        ▼
POST /auth/verify-otp
        │  ├─ reject if no matching / already-used / expired OTP → 401
        │  ├─ compare bcrypt hash                                → 401 on mismatch
        │  ├─ mark OTP consumed
        │  ├─ upsert user record by email
        │  └─ issue 24h JWT
        ▼
Client stores JWT, sends as `Authorization: Bearer <token>`
on all subsequent protected requests
```

## 2. Discovery → Booking Flow (happy path)

```
GET /facilities                       → list active facilities
        │
        ▼
GET /facilities/{id}/courts           → list courts for chosen facility
        │
        ▼
GET /courts/{id}/slots?date=YYYY-MM-DD → list slots with computed `available`
        │
        ▼
Student selects an open slot, sets player count
        │
        ▼
POST /bookings  { slot_id, player_count }
        │  ├─ reject player_count ≤ 0 or > court's max_players → 400
        │  ├─ reject slot in the past                          → 400
        │  ├─ INSERT booking (status='confirmed')
        │  │     └─ partial unique index guarantees only one
        │  │        confirmed row can exist per slot_id
        │  └─ on unique violation (23505)                      → 409 "slot already booked"
        ▼
201 Created → booking confirmed
```

## 3. Concurrent Booking Race (the core correctness scenario)

```
User A ──┐
User B ──┼── POST /bookings (same slot_id, ~same instant)
User C ──┘
              │
              ▼
    All three INSERTs hit Postgres
              │
              ▼
  Partial unique index allows exactly ONE
  confirmed row for this slot_id
              │
    ┌─────────┼─────────┐
    ▼         ▼         ▼
  201        409       409
 (winner)  (loser)   (loser)
```

No pre-check, no locking — the database itself is the only arbiter. This is demoed live via `cmd/demo/race_demo.go`, which fires N simultaneous requests against a real discovered slot and prints color-coded pass/fail results with latency timings.

## 4. Cancellation Flow

```
Student cancels their booking
        │
        ▼
DELETE /bookings/{id}
        │  ├─ reject if booking doesn't exist        → 404
        │  ├─ reject if not owned by requester        → 403
        │  ▼
        BEGIN transaction
          ├─ UPDATE bookings SET status='cancelled'
          ├─ SELECT waitlist_entries WHERE slot_id=? AND status='waiting'
          └─ UPDATE those rows → status='notified', notified_at=now()
        COMMIT
        │
        ▼
For each notified user_id → SSEHub.BroadcastToUser(...)
        │
        ▼
200 OK → slot is now open again
```

## 5. Waitlist Join Flow

```
Student taps an already-booked slot
        │
        ▼
UI shows: "Notify me when this slot opens up?"
        │
        ▼ (student taps Yes)
POST /slots/{id}/waitlist
        │  ├─ check slot is currently unavailable (booked) → else 400
        │  ├─ INSERT waitlist_entries (slot_id, user_id)
        │  │     └─ UNIQUE(slot_id, user_id) prevents duplicate joins
        │  └─ on unique violation (23505)                  → 409 "already on waitlist"
        ▼
201 Created → student is on the waitlist
```

The waitlist option is only ever shown on slots that are currently booked — it does not appear on open slots, matching the product decision that this is a "notify me" mechanism, not a general subscribe/follow feature.

## 6. Notify → Re-race → Cleanup Flow (full waitlist lifecycle)

```
Slot X is booked; students P, Q, R are on its waitlist ("waiting")
        │
        ▼
The booking on Slot X is cancelled (see Flow 4)
        │
        ▼
P, Q, R are all marked "notified" and pushed a live SSE event:
  { slot_id, court_label, facility_name, start_time }
        │
        ▼
P, Q, and R race to book Slot X — same mechanism as Flow 3
        │
    ┌───┼───┐
    ▼   ▼   ▼
   201 409 409     (say P wins)
        │
        ▼
Because Slot X had "notified" waitlist rows, in the SAME
transaction as P's successful booking insert:
  UPDATE waitlist_entries SET status='expired'
  WHERE slot_id = X AND status = 'notified'
        │
        ▼
Q and R's waitlist entries are now "expired" — they will not
receive a stale "still waiting" notification on this slot again.
Q and R next see Slot X as booked (by P) if they check.
```

## 7. Real-Time Connection Lifecycle (SSE)

```
Client logs in → opens EventSource to GET /notifications/stream
        │
        ▼
Server: authenticate via JWT (same middleware as other routes)
        │
        ▼
SSEHub.Register(user_id, channel)
        │
        ▼
Connection held open; server streams events as they're
broadcast to this user_id via SSEHub.BroadcastToUser(...)
        │
        ▼ (client disconnects / closes tab)
r.Context().Done() fires
        │
        ▼
SSEHub.Unregister(user_id, channel)
```

If a user isn't connected when a notification fires, they simply don't get the live push — the underlying state (`waitlist_entries.status = 'notified'`) is still durable in Postgres, so nothing is lost; it's a delivery convenience, not the source of truth.

## 8. Demo Script Flow (`cmd/demo/race_demo.go`)

```
Run `go run cmd/demo/race_demo.go [--reset] [-n <count>]`
        │
        ▼
Auto-register/authenticate N timestamped demo judge accounts
via POST /auth/request-otp + /auth/verify-otp (bypassing rate limits)
        │
        ▼
Auto-discover a real open future slot via
GET /facilities → GET /facilities/{id}/courts → GET /courts/{id}/slots
        │
        ▼
Arm a synchronized start barrier (channel-based)
        │
        ▼
Fire all N booking requests simultaneously
        │
        ▼
Print color-coded results per request (✅ 201 / ❌ 409) with
microsecond/millisecond latency, then a PASS/FAIL verdict banner
        │
        ▼ (if --reset passed)
Reset state so the demo can be re-run instantly for the next
judge walkthrough, without a full DB reset
```