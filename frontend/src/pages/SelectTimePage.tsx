import { useState, useEffect } from 'react';
import { useParams, useSearchParams, useNavigate } from 'react-router-dom';
import Header from '../components/Header';
import TimeSlot from '../components/TimeSlot';
import { facilityApi, bookingApi, waitlistApi } from '../api/client';
import { useAuth } from '../context/AuthContext';
import type { SlotWithAvailability } from '../api/types';

interface DayItem {
  day: string;
  date: number;
  monthName: string;
  dateStr: string; // YYYY-MM-DD
  key: string;
}

interface DisplaySlot {
  id: string;
  time: string;
  type: 'standard' | 'peak';
  state: 'available' | 'selected' | 'locked';
  rawSlot?: SlotWithAvailability;
}

function generateUpcomingDays(count: number = 5): DayItem[] {
  const days: DayItem[] = [];
  const now = new Date();

  for (let i = 0; i < count; i++) {
    const d = new Date(now);
    d.setDate(now.getDate() + i);

    const year = d.getFullYear();
    const month = String(d.getMonth() + 1).padStart(2, '0');
    const dayOfMonth = String(d.getDate()).padStart(2, '0');
    const dateStr = `${year}-${month}-${dayOfMonth}`;

    const dayName = i === 0 ? 'TODAY' : d.toLocaleDateString('en-US', { weekday: 'short' }).toUpperCase();
    const monthName = d.toLocaleDateString('en-US', { month: 'long' });

    days.push({
      day: dayName,
      date: d.getDate(),
      monthName,
      dateStr,
      key: dateStr,
    });
  }

  return days;
}

function formatTimeRange(startStr: string, endStr: string): string {
  try {
    const start = new Date(startStr);
    const end = new Date(endStr);
    const sHours = String(start.getHours()).padStart(2, '0');
    const sMins = String(start.getMinutes()).padStart(2, '0');
    const eHours = String(end.getHours()).padStart(2, '0');
    const eMins = String(end.getMinutes()).padStart(2, '0');
    return `${sHours}:${sMins} - ${eHours}:${eMins}`;
  } catch {
    return '08:00 - 09:00';
  }
}

export default function SelectTimePage() {
  const { courtId } = useParams<{ sport: string; courtId: string }>();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();

  const courtName = searchParams.get('courtName') || `Court ${courtId}`;
  const facilityName = searchParams.get('facilityName') || 'Campus Complex';
  const maxPlayers = parseInt(searchParams.get('maxPlayers') || '4', 10);

  const days = generateUpcomingDays(5);
  const [selectedDayKey, setSelectedDayKey] = useState<string>(days[0].dateStr);

  const [slots, setSlots] = useState<DisplaySlot[]>([]);
  const [selectedSlotId, setSelectedSlotId] = useState<string | null>(null);
  const [playerCount, setPlayerCount] = useState<number>(2);

  const [loading, setLoading] = useState<boolean>(true);
  const [bookingLoading, setBookingLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  // Modals
  const [showConfirmModal, setShowConfirmModal] = useState<boolean>(false);
  const [confirmedBookingDetails, setConfirmedBookingDetails] = useState<{
    court: string;
    date: string;
    time: string;
    players: number;
  } | null>(null);

  // Waitlist Modal
  const [waitlistTargetSlot, setWaitlistTargetSlot] = useState<DisplaySlot | null>(null);
  const [waitlistLoading, setWaitlistLoading] = useState<boolean>(false);
  const [waitlistSuccess, setWaitlistSuccess] = useState<string | null>(null);

  const selectedDay = days.find((d) => d.dateStr === selectedDayKey) || days[0];

  const isUUID = (str?: string) =>
    !!str && /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(str);

  useEffect(() => {
    async function loadSlots() {
      if (!courtId) {
        setLoading(false);
        return;
      }

      setLoading(true);
      setError(null);
      setSelectedSlotId(null);

      try {
        if (isUUID(courtId)) {
          const backendSlots = await facilityApi.listSlots(courtId, selectedDay.dateStr);

          if (backendSlots && backendSlots.length > 0) {
            const mapped: DisplaySlot[] = backendSlots.map((s) => {
              const timeRange = formatTimeRange(s.start_time, s.end_time);
              const startHour = new Date(s.start_time).getHours();
              // Peak hours are 7-9am and 17-21pm
              const isPeak = (startHour >= 7 && startHour <= 9) || (startHour >= 17 && startHour <= 21);

              return {
                id: s.id,
                time: timeRange,
                type: isPeak ? 'peak' : 'standard',
                state: s.available ? 'available' : 'locked',
                rawSlot: s,
              };
            });

            setSlots(mapped);

            // Auto-select first available slot
            const firstAvailable = mapped.find((m) => m.state === 'available');
            if (firstAvailable) {
              setSelectedSlotId(firstAvailable.id);
            }
            setLoading(false);
            return;
          }
        }

        // Fallback default slots if courtId is mock or not yet seeded
        const defaultSlotTimes = [
          { time: '06:00 - 07:00', isPeak: false },
          { time: '07:15 - 08:15', isPeak: true },
          { time: '08:30 - 09:30', isPeak: true },
          { time: '09:45 - 10:45', isPeak: true },
          { time: '11:00 - 12:00', isPeak: false },
          { time: '12:15 - 13:15', isPeak: false },
          { time: '13:30 - 14:30', isPeak: false },
          { time: '16:00 - 17:00', isPeak: false },
          { time: '17:15 - 18:15', isPeak: true },
          { time: '18:30 - 19:30', isPeak: true },
        ];

        const mockSlots: DisplaySlot[] = defaultSlotTimes.map((item, idx) => ({
          id: `mock-${idx}`,
          time: item.time,
          type: item.isPeak ? 'peak' : 'standard',
          state: idx === 1 ? 'locked' : 'available',
        }));

        setSlots(mockSlots);
        setSelectedSlotId(mockSlots[2].id);
      } catch (err) {
        console.warn('Could not load slots from backend, using fallback:', err);
      } finally {
        setLoading(false);
      }
    }

    loadSlots();
  }, [courtId, selectedDayKey]);

  const handleSlotClick = (slot: DisplaySlot) => {
    if (slot.state === 'locked') {
      // Open Waitlist prompt for booked slot
      setWaitlistTargetSlot(slot);
      setWaitlistSuccess(null);
    } else {
      setSelectedSlotId(slot.id);
    }
  };

  const handleJoinWaitlist = async () => {
    if (!waitlistTargetSlot) return;

    if (!isAuthenticated) {
      navigate('/login');
      return;
    }

    setWaitlistLoading(true);
    setError(null);

    try {
      if (isUUID(waitlistTargetSlot.id)) {
        await waitlistApi.joinWaitlist(waitlistTargetSlot.id);
      }
      setWaitlistSuccess(`You're on the waitlist for ${waitlistTargetSlot.time}! We'll notify you instantly if it opens.`);
    } catch (err: any) {
      if (err?.status === 409) {
        setWaitlistSuccess("You're already on the waitlist for this slot! We'll alert you if it frees up.");
      } else {
        setError(err?.message || 'Failed to join waitlist. Please try again.');
      }
    } finally {
      setWaitlistLoading(false);
    }
  };

  const handleBookSlot = async () => {
    if (!selectedSlotId) return;

    if (!isAuthenticated) {
      navigate('/login');
      return;
    }

    const currentSlot = slots.find((s) => s.id === selectedSlotId);
    if (!currentSlot) return;

    setBookingLoading(true);
    setError(null);

    try {
      if (isUUID(selectedSlotId)) {
        await bookingApi.createBooking(selectedSlotId, playerCount);
      }

      setConfirmedBookingDetails({
        court: `${courtName} (${facilityName})`,
        date: `${selectedDay.day}, ${selectedDay.monthName} ${selectedDay.date}`,
        time: currentSlot.time,
        players: playerCount,
      });

      setShowConfirmModal(true);
    } catch (err: any) {
      if (err?.status === 409) {
        // Slot conflict: prompt to join waitlist!
        setWaitlistTargetSlot(currentSlot);
        setError('This slot was just booked by another player. Would you like to join the waitlist?');
      } else {
        setError(err?.message || 'Failed to book slot. Please try again.');
      }
    } finally {
      setBookingLoading(false);
    }
  };

  const selectedSlot = slots.find((s) => s.id === selectedSlotId);

  // Group slots into morning & afternoon
  const morningSlots = slots.slice(0, Math.ceil(slots.length / 2));
  const afternoonSlots = slots.slice(Math.ceil(slots.length / 2));

  return (
    <div className="min-h-dvh bg-blush pb-36">
      <Header title="Book Court" showLogo showAvatar />

      <div className="px-5 pt-2">
        <h2 className="text-2xl font-extrabold text-plum leading-tight">Select a Time</h2>
        <p className="text-sm text-plum-muted mt-1">{courtName} • {facilityName}</p>
      </div>

      {/* Month Header */}
      <div className="px-5 mt-5 flex items-center justify-between">
        <h3 className="text-lg font-bold text-plum">{selectedDay.monthName}</h3>
      </div>

      {/* Day Selector */}
      <div className="px-5 mt-3 flex gap-3 overflow-x-auto hide-scrollbar">
        {days.map((d) => (
          <button
            key={d.key}
            onClick={() => setSelectedDayKey(d.dateStr)}
            className={`flex flex-col items-center min-w-[64px] py-3 px-3 rounded-2xl transition-all duration-300 ${
              selectedDayKey === d.dateStr
                ? 'bg-plum text-white shadow-md shadow-plum/20 scale-[1.03]'
                : 'bg-white text-plum hover:bg-blush-dark/30'
            }`}
          >
            <span className={`text-[10px] font-bold tracking-wider ${selectedDayKey === d.dateStr ? 'text-white/70' : 'text-plum-muted'}`}>
              {d.day}
            </span>
            <span className="text-lg font-extrabold mt-0.5">{d.date}</span>
          </button>
        ))}
      </div>

      {/* Legend */}
      <div className="px-5 mt-5 flex items-center justify-between">
        <div className="flex items-center gap-5">
          <div className="flex items-center gap-2">
            <span className="w-3 h-3 rounded-full bg-white border border-blush-dark" />
            <span className="text-xs font-medium text-plum-muted">Standard</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="w-3.5 h-3.5 rounded-full bg-[#E05A2B]/15 border-2 border-[#E05A2B]" />
            <span className="text-xs font-medium text-plum-muted">Peak</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="w-3 h-3 rounded-full bg-blush-dark/40" />
            <span className="text-xs font-medium text-plum-muted">Booked</span>
          </div>
        </div>
      </div>

      {/* Player Count Selector */}
      <div className="mx-5 mt-4 bg-white rounded-2xl p-4 shadow-xs flex items-center justify-between border border-blush-dark/20">
        <div>
          <p className="text-xs font-bold text-plum uppercase tracking-wider">Players</p>
          <p className="text-[11px] text-plum-muted">Max capacity: {maxPlayers} players</p>
        </div>
        <div className="flex items-center gap-2 bg-blush rounded-xl p-1">
          <button
            type="button"
            onClick={() => setPlayerCount((p) => Math.max(1, p - 1))}
            className="w-8 h-8 rounded-lg bg-white text-plum font-bold shadow-xs flex items-center justify-center hover:bg-blush-light active:scale-95 transition-transform"
          >
            -
          </button>
          <span className="w-6 text-center font-bold text-sm text-plum">{playerCount}</span>
          <button
            type="button"
            onClick={() => setPlayerCount((p) => Math.min(maxPlayers, p + 1))}
            className="w-8 h-8 rounded-lg bg-white text-plum font-bold shadow-xs flex items-center justify-center hover:bg-blush-light active:scale-95 transition-transform"
          >
            +
          </button>
        </div>
      </div>

      {error && (
        <div className="mx-5 mt-4 bg-danger/10 border border-danger/20 rounded-xl p-3 flex items-center justify-between">
          <p className="text-xs font-medium text-danger">{error}</p>
          {error.includes('waitlist') && (
            <button
              onClick={() => {
                if (selectedSlot) setWaitlistTargetSlot(selectedSlot);
              }}
              className="text-xs font-bold text-danger underline ml-2"
            >
              Join Waitlist
            </button>
          )}
        </div>
      )}

      {loading ? (
        <div className="py-12 flex flex-col items-center justify-center gap-3">
          <div className="w-8 h-8 border-3 border-plum/20 border-t-plum rounded-full animate-spin" />
          <p className="text-xs text-plum-muted">Loading court slot availability...</p>
        </div>
      ) : (
        <>
          {/* Morning Slots */}
          {morningSlots.length > 0 && (
            <div className="px-5 mt-5">
              <p className="text-xs font-semibold tracking-widest uppercase text-plum-muted mb-3">Morning</p>
              <div className="grid grid-cols-2 gap-2.5">
                {morningSlots.map((slot) => (
                  <TimeSlot
                    key={slot.id}
                    time={slot.time}
                    type={slot.type}
                    state={slot.id === selectedSlotId ? 'selected' : slot.state}
                    onClick={() => handleSlotClick(slot)}
                  />
                ))}
              </div>
            </div>
          )}

          {/* Afternoon Slots */}
          {afternoonSlots.length > 0 && (
            <div className="px-5 mt-5">
              <p className="text-xs font-semibold tracking-widest uppercase text-plum-muted mb-3">Afternoon & Evening</p>
              <div className="grid grid-cols-2 gap-2.5">
                {afternoonSlots.map((slot) => (
                  <TimeSlot
                    key={slot.id}
                    time={slot.time}
                    type={slot.type}
                    state={slot.id === selectedSlotId ? 'selected' : slot.state}
                    onClick={() => handleSlotClick(slot)}
                  />
                ))}
              </div>
            </div>
          )}
        </>
      )}

      {/* Sticky Bottom Button */}
      <div className="fixed bottom-0 left-1/2 -translate-x-1/2 w-full max-w-[430px] px-5 pb-[max(1.25rem,env(safe-area-inset-bottom))] pt-3 bg-gradient-to-t from-blush via-blush to-blush/0 z-40">
        <button
          onClick={handleBookSlot}
          disabled={!selectedSlotId || bookingLoading}
          className="w-full bg-plum text-white py-4 rounded-2xl font-semibold text-base flex items-center justify-center gap-2 hover:bg-plum-light transition-colors active:scale-[0.98] shadow-lg shadow-plum/20 disabled:opacity-50"
        >
          {bookingLoading ? (
            <span className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
          ) : (
            <>
              Book Slot ({playerCount} {playerCount === 1 ? 'player' : 'players'})
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <line x1="5" y1="12" x2="19" y2="12" />
                <polyline points="12 5 19 12 12 19" />
              </svg>
            </>
          )}
        </button>
      </div>

      {/* Waitlist Modal */}
      {waitlistTargetSlot && (
        <div className="fixed inset-0 z-50 flex items-center justify-center px-6">
          <div className="absolute inset-0 bg-plum/40 backdrop-blur-sm" onClick={() => setWaitlistTargetSlot(null)} />
          <div className="relative bg-white rounded-3xl p-6 w-full max-w-sm shadow-2xl animate-[scaleIn_0.2s_ease-out]">
            <div className="w-12 h-12 rounded-full bg-warning/15 text-warning flex items-center justify-center mx-auto mb-3 text-xl">
              ⏳
            </div>
            <h3 className="text-lg font-bold text-plum text-center">Slot Booked</h3>
            <p className="text-xs text-plum-muted text-center mt-1">
              {courtName} • {selectedDay.day}, {selectedDay.date} ({waitlistTargetSlot.time})
            </p>

            {waitlistSuccess ? (
              <div className="mt-4 bg-success/10 border border-success/20 rounded-2xl p-4 text-center">
                <p className="text-xs font-semibold text-success">{waitlistSuccess}</p>
                <button
                  onClick={() => setWaitlistTargetSlot(null)}
                  className="mt-3 w-full bg-plum text-white py-2.5 rounded-xl font-semibold text-xs"
                >
                  Got It
                </button>
              </div>
            ) : (
              <>
                <p className="text-xs text-plum-muted/80 text-center mt-3 bg-blush p-3 rounded-xl">
                  This slot is currently occupied. Join the waitlist to receive an instant real-time alert the second it gets freed up!
                </p>

                <div className="flex gap-2 mt-5">
                  <button
                    onClick={() => setWaitlistTargetSlot(null)}
                    className="flex-1 bg-blush text-plum py-3 rounded-xl font-semibold text-xs hover:bg-blush-dark transition-colors"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={handleJoinWaitlist}
                    disabled={waitlistLoading}
                    className="flex-1 bg-[#E05A2B] text-white py-3 rounded-xl font-semibold text-xs hover:bg-[#C84E23] transition-colors disabled:opacity-50 flex items-center justify-center gap-1.5"
                  >
                    {waitlistLoading ? (
                      <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                    ) : (
                      'Join Waitlist'
                    )}
                  </button>
                </div>
              </>
            )}
          </div>
        </div>
      )}

      {/* Booking Confirmation Popup */}
      {showConfirmModal && confirmedBookingDetails && (
        <div className="fixed inset-0 z-50 flex items-center justify-center px-6">
          <div className="absolute inset-0 bg-plum/40 backdrop-blur-sm" onClick={() => setShowConfirmModal(false)} />
          <div className="relative bg-white rounded-3xl p-6 w-full max-w-sm shadow-2xl animate-[scaleIn_0.2s_ease-out]">
            {/* Success Icon */}
            <div className="w-14 h-14 rounded-full bg-success/10 flex items-center justify-center mx-auto mb-4">
              <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#22C55E" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <polyline points="20 6 9 17 4 12" />
              </svg>
            </div>

            <h3 className="text-lg font-bold text-plum text-center">Booking Confirmed!</h3>
            <p className="text-sm text-plum-muted text-center mt-1">Your slot has been recorded in the system.</p>

            {/* Booking Details */}
            <div className="bg-blush rounded-2xl p-4 mt-5 flex flex-col gap-2.5">
              <div className="flex items-center justify-between">
                <span className="text-xs font-medium text-plum-muted">Court</span>
                <span className="text-sm font-semibold text-plum">{confirmedBookingDetails.court}</span>
              </div>
              <div className="border-t border-blush-dark/30" />
              <div className="flex items-center justify-between">
                <span className="text-xs font-medium text-plum-muted">Date</span>
                <span className="text-sm font-semibold text-plum">{confirmedBookingDetails.date}</span>
              </div>
              <div className="border-t border-blush-dark/30" />
              <div className="flex items-center justify-between">
                <span className="text-xs font-medium text-plum-muted">Time</span>
                <span className="text-sm font-semibold text-plum">{confirmedBookingDetails.time}</span>
              </div>
              <div className="border-t border-blush-dark/30" />
              <div className="flex items-center justify-between">
                <span className="text-xs font-medium text-plum-muted">Players</span>
                <span className="text-sm font-semibold text-plum">{confirmedBookingDetails.players}</span>
              </div>
            </div>

            <div className="flex gap-2 mt-5">
              <button
                onClick={() => {
                  setShowConfirmModal(false);
                  navigate('/bookings');
                }}
                className="flex-1 bg-plum text-white py-3.5 rounded-xl font-semibold text-sm hover:bg-plum-light transition-colors active:scale-[0.98]"
              >
                View in My Bookings
              </button>
              <button
                onClick={() => setShowConfirmModal(false)}
                className="bg-blush text-plum px-4 py-3.5 rounded-xl font-semibold text-sm hover:bg-blush-dark transition-colors"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
