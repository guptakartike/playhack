#!/bin/bash
# ============================================================
#  PlayHack — Manual Waitlist + Concurrency + SSE Test Script
# ============================================================
#  Prerequisites:
#    1. Backend running: cd backend && go run cmd/api/main.go
#    2. Database seeded: cd backend && go run cmd/seed/main.go
#
#  Usage:  chmod +x test_waitlist_flow.sh && ./test_waitlist_flow.sh
# ============================================================

set -e
BASE="http://127.0.0.1:5532"
BOLD="\033[1m"
GREEN="\033[32m"
YELLOW="\033[33m"
CYAN="\033[36m"
RED="\033[31m"
RESET="\033[0m"

sep() { echo -e "\n${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"; }

TS=$(date +%s%N | cut -b1-13)
USER_A="alice_${TS}@iitg.ac.in"
USER_B="bob_${TS}@iitg.ac.in"

# ── Step 1: Authenticate User A ──
sep
echo -e "${BOLD}Step 1: Authenticate User A ($USER_A)${RESET}"
OTP_RESP_A=$(curl -s -X POST "$BASE/auth/request-otp" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$USER_A\"}")
echo "  OTP Response: $OTP_RESP_A"
CODE_A=$(echo "$OTP_RESP_A" | python3 -c "import sys,json; print(json.load(sys.stdin).get('code',''))" 2>/dev/null || echo "")

VERIFY_A=$(curl -s -X POST "$BASE/auth/verify-otp" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$USER_A\",\"code\":\"$CODE_A\"}")
TOKEN_A=$(echo "$VERIFY_A" | python3 -c "import sys,json; print(json.load(sys.stdin).get('token',''))" 2>/dev/null || echo "")
echo -e "  ${GREEN}✓ User A Token: ${TOKEN_A:0:30}...${RESET}"

# ── Step 2: Authenticate User B ──
sep
echo -e "${BOLD}Step 2: Authenticate User B ($USER_B)${RESET}"
OTP_RESP_B=$(curl -s -X POST "$BASE/auth/request-otp" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$USER_B\"}")
CODE_B=$(echo "$OTP_RESP_B" | python3 -c "import sys,json; print(json.load(sys.stdin).get('code',''))" 2>/dev/null || echo "")

VERIFY_B=$(curl -s -X POST "$BASE/auth/verify-otp" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$USER_B\",\"code\":\"$CODE_B\"}")
TOKEN_B=$(echo "$VERIFY_B" | python3 -c "import sys,json; print(json.load(sys.stdin).get('token',''))" 2>/dev/null || echo "")
echo -e "  ${GREEN}✓ User B Token: ${TOKEN_B:0:30}...${RESET}"

# ── Step 3: List facilities and pick a court + slot ──
sep
echo -e "${BOLD}Step 3: Find a bookable slot${RESET}"
FACILITIES=$(curl -s "$BASE/facilities")
FACILITY_ID=$(echo "$FACILITIES" | python3 -c "import sys,json; print(json.load(sys.stdin)[0]['id'])" 2>/dev/null || echo "")
FACILITY_NAME=$(echo "$FACILITIES" | python3 -c "import sys,json; print(json.load(sys.stdin)[0]['name'])" 2>/dev/null || echo "")
echo "  Facility: $FACILITY_NAME ($FACILITY_ID)"

COURTS=$(curl -s "$BASE/facilities/$FACILITY_ID/courts")
COURT_ID=$(echo "$COURTS" | python3 -c "import sys,json; print(json.load(sys.stdin)[0]['id'])" 2>/dev/null || echo "")
COURT_LABEL=$(echo "$COURTS" | python3 -c "import sys,json; print(json.load(sys.stdin)[0]['label'])" 2>/dev/null || echo "")
echo "  Court: $COURT_LABEL ($COURT_ID)"

TARGET_DATE=$(python3 -c "import datetime; print((datetime.datetime.now() + datetime.timedelta(days=1)).strftime('%Y-%m-%d'))")
echo "  Date: $TARGET_DATE (Tomorrow)"
SLOTS=$(curl -s "$BASE/courts/$COURT_ID/slots?date=$TARGET_DATE")
SLOT_ID=$(echo "$SLOTS" | python3 -c "
import sys, json
slots = json.load(sys.stdin)
avail = [s for s in slots if s['available']]
print(avail[0]['id'] if avail else '')
" 2>/dev/null || echo "")

if [ -z "$SLOT_ID" ]; then
  echo -e "  ${RED}✗ No available slots for today. Try seeding: go run cmd/seed/main.go${RESET}"
  exit 1
fi
echo -e "  ${GREEN}✓ Available Slot: $SLOT_ID${RESET}"

# ── Step 4: User A books the slot ──
sep
echo -e "${BOLD}Step 4: User A books the slot${RESET}"
BOOK_RESP=$(curl -s -w "\n%{http_code}" -X POST "$BASE/bookings" \
  -H "Authorization: Bearer $TOKEN_A" \
  -H "Content-Type: application/json" \
  -d "{\"slot_id\":\"$SLOT_ID\",\"player_count\":2}")
BOOK_HTTP=$(echo "$BOOK_RESP" | tail -1)
BOOK_BODY=$(echo "$BOOK_RESP" | head -1)
BOOKING_ID=$(echo "$BOOK_BODY" | python3 -c "import sys,json; print(json.load(sys.stdin).get('id',''))" 2>/dev/null || echo "")
echo -e "  HTTP $BOOK_HTTP — Booking ID: $BOOKING_ID"
if [ "$BOOK_HTTP" = "201" ]; then
  echo -e "  ${GREEN}✓ Slot booked by User A${RESET}"
else
  echo -e "  ${RED}✗ Booking failed: $BOOK_BODY${RESET}"
  exit 1
fi

# ── Step 5: User B tries to book the same slot → 409 ──
sep
echo -e "${BOLD}Step 5: User B tries to book the same slot (expect 409 Conflict)${RESET}"
CONFLICT_RESP=$(curl -s -w "\n%{http_code}" -X POST "$BASE/bookings" \
  -H "Authorization: Bearer $TOKEN_B" \
  -H "Content-Type: application/json" \
  -d "{\"slot_id\":\"$SLOT_ID\",\"player_count\":2}")
CONFLICT_HTTP=$(echo "$CONFLICT_RESP" | tail -1)
echo -e "  HTTP $CONFLICT_HTTP"
if [ "$CONFLICT_HTTP" = "409" ]; then
  echo -e "  ${GREEN}✓ Correctly rejected — slot already booked${RESET}"
else
  echo -e "  ${RED}✗ Unexpected response: $(echo "$CONFLICT_RESP" | head -1)${RESET}"
fi

# ── Step 6: User B joins waitlist ──
sep
echo -e "${BOLD}Step 6: User B joins the waitlist for the booked slot${RESET}"
WAITLIST_RESP=$(curl -s -w "\n%{http_code}" -X POST "$BASE/slots/$SLOT_ID/waitlist" \
  -H "Authorization: Bearer $TOKEN_B")
WAITLIST_HTTP=$(echo "$WAITLIST_RESP" | tail -1)
WAITLIST_BODY=$(echo "$WAITLIST_RESP" | head -1)
echo -e "  HTTP $WAITLIST_HTTP — $WAITLIST_BODY"
if [ "$WAITLIST_HTTP" = "201" ]; then
  echo -e "  ${GREEN}✓ User B is now on the waitlist!${RESET}"
else
  echo -e "  ${RED}✗ Waitlist join failed${RESET}"
fi

# ── Step 7: User B tries to join again → 409 ──
sep
echo -e "${BOLD}Step 7: User B joins waitlist again (expect 409 duplicate)${RESET}"
DUP_RESP=$(curl -s -w "\n%{http_code}" -X POST "$BASE/slots/$SLOT_ID/waitlist" \
  -H "Authorization: Bearer $TOKEN_B")
DUP_HTTP=$(echo "$DUP_RESP" | tail -1)
echo -e "  HTTP $DUP_HTTP"
if [ "$DUP_HTTP" = "409" ]; then
  echo -e "  ${GREEN}✓ Correctly rejected — already on waitlist${RESET}"
else
  echo -e "  ${RED}✗ Unexpected: $(echo "$DUP_RESP" | head -1)${RESET}"
fi

# ── Step 8: Start SSE listener for User B (background) ──
sep
echo -e "${BOLD}Step 8: Start SSE notification listener for User B (background)${RESET}"
SSE_LOG="/tmp/playhack_sse_test.log"
> "$SSE_LOG"
curl -s -N -H "Authorization: Bearer $TOKEN_B" "$BASE/notifications/stream" > "$SSE_LOG" 2>&1 &
SSE_PID=$!
echo -e "  ${YELLOW}SSE listener PID: $SSE_PID (logging to $SSE_LOG)${RESET}"
sleep 1

# ── Step 9: User A cancels booking → triggers SSE notification to User B ──
sep
echo -e "${BOLD}Step 9: User A cancels booking → should trigger SSE to User B${RESET}"
CANCEL_RESP=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE/bookings/$BOOKING_ID" \
  -H "Authorization: Bearer $TOKEN_A")
CANCEL_HTTP=$(echo "$CANCEL_RESP" | tail -1)
echo -e "  HTTP $CANCEL_HTTP"
if [ "$CANCEL_HTTP" = "200" ]; then
  echo -e "  ${GREEN}✓ Booking cancelled by User A${RESET}"
else
  echo -e "  ${RED}✗ Cancel failed: $(echo "$CANCEL_RESP" | head -1)${RESET}"
fi

sleep 2  # Give SSE time to deliver

# ── Step 10: Check SSE log for notification ──
sep
echo -e "${BOLD}Step 10: Check if User B received the SSE notification${RESET}"
SSE_CONTENT=$(cat "$SSE_LOG")
if echo "$SSE_CONTENT" | grep -q "slot_id"; then
  echo -e "  ${GREEN}✓ SSE NOTIFICATION RECEIVED!${RESET}"
  echo -e "  Payload: $SSE_CONTENT"
else
  echo -e "  ${YELLOW}⚠ No SSE data captured yet (may need more time or check backend logs)${RESET}"
  echo -e "  Log contents: $SSE_CONTENT"
fi

# Cleanup SSE listener
kill $SSE_PID 2>/dev/null || true
wait $SSE_PID 2>/dev/null || true

# ── Step 11: Concurrency Race Test ──
sep
echo -e "${BOLD}Step 11: Concurrency Race — 10 users booking the SAME slot simultaneously${RESET}"

# Pick a fresh available slot
SLOTS2=$(curl -s "$BASE/courts/$COURT_ID/slots?date=$TARGET_DATE")
RACE_SLOT_ID=$(echo "$SLOTS2" | python3 -c "
import sys, json
slots = json.load(sys.stdin)
avail = [s for s in slots if s['available']]
print(avail[0]['id'] if avail else '')
" 2>/dev/null || echo "")

if [ -z "$RACE_SLOT_ID" ]; then
  echo -e "  ${YELLOW}⚠ No more available slots for concurrency test. Skipping.${RESET}"
else
  echo "  Target Slot: $RACE_SLOT_ID"
  echo "  Creating 10 race users..."

  RACE_TOKENS=()
  for i in $(seq 1 10); do
    EMAIL="racer_${TS}_${i}@iitg.ac.in"
    CODE=$(curl -s -X POST "$BASE/auth/request-otp" \
      -H "Content-Type: application/json" \
      -d "{\"email\":\"$EMAIL\"}" | python3 -c "import sys,json; print(json.load(sys.stdin).get('code',''))" 2>/dev/null)
    TOKEN=$(curl -s -X POST "$BASE/auth/verify-otp" \
      -H "Content-Type: application/json" \
      -d "{\"email\":\"$EMAIL\",\"code\":\"$CODE\"}" | python3 -c "import sys,json; print(json.load(sys.stdin).get('token',''))" 2>/dev/null)
    RACE_TOKENS+=("$TOKEN")
  done
  echo -e "  ${GREEN}✓ 10 users authenticated${RESET}"

  echo "  Launching all 10 booking requests simultaneously..."
  RESULT_FILE="/tmp/playhack_race_results.txt"
  > "$RESULT_FILE"

  for TOKEN in "${RACE_TOKENS[@]}"; do
    curl -s -o /dev/null -w "%{http_code}\n" -X POST "$BASE/bookings" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "{\"slot_id\":\"$RACE_SLOT_ID\",\"player_count\":1}" >> "$RESULT_FILE" &
  done
  wait

  WINNERS=$(grep -c "201" "$RESULT_FILE" || true)
  CONFLICTS=$(grep -c "409" "$RESULT_FILE" || true)
  echo ""
  echo -e "  ┌─────────────────────────────────┐"
  echo -e "  │  ${BOLD}CONCURRENCY RACE RESULTS${RESET}       │"
  echo -e "  ├─────────────────────────────────┤"
  echo -e "  │  Winners (201):   $WINNERS                 │"
  echo -e "  │  Conflicts (409): $CONFLICTS                 │"
  echo -e "  └─────────────────────────────────┘"

  if [ "$WINNERS" = "1" ]; then
    echo -e "  ${GREEN}🎉 RACE PASSED: Exactly 1 winner!${RESET}"
  else
    echo -e "  ${RED}✗ RACE FAILED: Expected 1 winner, got $WINNERS${RESET}"
  fi
fi

# ── Done ──
sep
echo -e "${GREEN}${BOLD}All manual tests complete!${RESET}"
echo ""
