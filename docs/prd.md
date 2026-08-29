# Huddle Up — Product Requirements Document

## 1. Overview

Huddle Up is a sports facility booking system for IIT Guwahati students, built for the Sports Board x Tech Board hackathon (SDE track). The product's central promise is **correctness under contention**: when multiple students want the same court slot, the system must guarantee exactly one of them gets it — with no double-bookings, no lost updates, and no silent failures.

## 2. Problem Statement

IITG has a limited number of sports facilities (courts, turfs) shared across the student body. Manual or naive booking systems (physical sign-up sheets, group chats, first-come-first-serve at the venue, or booking apps without proper concurrency handling) break down under real contention: two students can both believe they've secured the same slot, leading to disputes, wasted trips, and no reliable source of truth.

## 3. Goals

- Guarantee **at most one confirmed booking per slot**, even under simultaneous requests.
- Restrict access to verified IITG students (`@iitg.ac.in` email only).
- Give students visibility into slot availability in real time.
- When a slot they wanted opens up (via cancellation), notify interested students automatically instead of requiring them to keep refreshing.
- Keep the system's guarantees demonstrable — the correctness claims should be provable live, not just asserted.

## 4. Non-Goals (for this prototype)

- Payment processing.
- Admin-side analytics dashboard (identified as a roadmap item, not built).
- Check-in / photo verification flow (referenced in early UI concepts, not implemented in the backend).
- Multi-campus or multi-institution support.

## 5. User Roles

- **Student** — browses facilities/courts/slots, books, cancels, joins waitlists, receives notifications.
- **Admin** (`role` field exists on `users`, default `student`) — reserved for future admin capabilities; no admin-only endpoints exist yet.

## 6. Functional Requirements

### 6.1 Authentication
- Login restricted to `@iitg.ac.in` email addresses.
- OTP-based flow: request a 6-digit code, verify it, receive a JWT (24h expiry).
- OTP requests rate-limited to one per email per 60 seconds.

### 6.2 Facility & Slot Discovery
- List active facilities, courts per facility, and slots per court (filterable by date).
- Each slot exposes a computed `available` boolean.

### 6.3 Booking
- A student can book any available, future slot with a player count.
- Exactly one confirmed booking may exist per slot at any time — enforced at the database level, not just in application logic.
- Concurrent booking attempts on the same slot must resolve to exactly one success (`201`) and all others receiving a conflict response (`409`).

### 6.4 Cancellation
- A student can cancel only their own booking.
- Cancelling a booking immediately frees the slot for rebooking.

### 6.5 Waitlist & Notifications
- A student may only join a waitlist for a slot that is **currently booked** (not for open slots).
- Joining the same slot's waitlist twice is rejected — enforced at the database level via a uniqueness constraint, not in-memory tracking.
- When a booking on a waitlisted slot is cancelled, **all** waiting students for that slot are notified simultaneously via a live connection (not polling).
- The first student to act on the notification and successfully book the now-open slot wins — using the same booking-conflict mechanism as any other booking race.
- Waitlist entries not acted upon (i.e., belonging to students who didn't win the re-booking race) are cleared out once the slot is re-booked, so they don't linger or resurface with stale information.

### 6.6 My Bookings
- A student can view their own bookings, listed with facility/court/time details.

## 7. Non-Functional Requirements

- **Correctness over throughput**: for a hackathon-scale prototype, prioritizing demonstrable correctness (no double-booking under load) over performance optimization.
- **Statelessness where it matters**: any mechanism affecting booking or waitlist correctness must not depend on in-memory, single-process state, so the guarantees hold even if the system were later scaled to multiple instances.
- **Auditability**: the reasoning behind each concurrency-safety decision should be traceable (see `decision.md`).

## 8. Success Criteria (Hackathon Demo)

- Live demonstration: N simultaneous booking requests on the same slot resolve to exactly 1 success, N-1 conflicts, reproducibly.
- Live or recorded demonstration of the waitlist notify → race → resolve flow.
- No plaintext secrets or credentials present in the submitted repository.

## 9. Known Gaps / Roadmap

- Admin analytics dashboard — not built.
- Check-in / photo verification with booking status chips beyond confirmed/cancelled — not built.
- OTP delivery is currently returned directly in the API response for demo convenience rather than emailed — must be gated before any real deployment (see `decision.md` for rationale and risk).