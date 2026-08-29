import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import Header from '../components/Header';
import BottomNav from '../components/BottomNav';
import BookingCard from '../components/BookingCard';
import { bookingApi } from '../api/client';
import { useAuth } from '../context/AuthContext';
import type { BookingDetail } from '../api/types';

function formatBookingTime(startStr: string, endStr: string): string {
  try {
    const start = new Date(startStr);
    const end = new Date(endStr);

    const isToday = new Date().toDateString() === start.toDateString();
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);
    const isTomorrow = tomorrow.toDateString() === start.toDateString();

    let dayLabel = start.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
    if (isToday) dayLabel = 'Today';
    if (isTomorrow) dayLabel = 'Tomorrow';

    const sTime = `${String(start.getHours()).padStart(2, '0')}:${String(start.getMinutes()).padStart(2, '0')}`;
    const eTime = `${String(end.getHours()).padStart(2, '0')}:${String(end.getMinutes()).padStart(2, '0')}`;

    return `${dayLabel}, ${sTime} - ${eTime}`;
  } catch {
    return `${startStr} - ${endStr}`;
  }
}

export default function MyBookingsPage() {
  const navigate = useNavigate();
  const { user, token, isAuthenticated, logout } = useAuth();

  const [bookings, setBookings] = useState<BookingDetail[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [cancellingId, setCancellingId] = useState<string | null>(null);
  const [bookingToCancel, setBookingToCancel] = useState<BookingDetail | null>(null);
  const [toastMessage, setToastMessage] = useState<string | null>(null);

  useEffect(() => {
    async function loadBookings() {
      if (!isAuthenticated || !user) {
        setBookings([]);
        setLoading(false);
        return;
      }

      setLoading(true);
      setBookings([]); // Clear old user's bookings immediately

      try {
        const data = await bookingApi.getMyBookings();
        setBookings(data || []);
      } catch (err) {
        console.warn('Could not load user bookings:', err);
      } finally {
        setLoading(false);
      }
    }

    loadBookings();
  }, [user?.id, token]);

  const handleConfirmCancel = async () => {
    if (!bookingToCancel) return;

    setCancellingId(bookingToCancel.id);
    try {
      await bookingApi.cancelBooking(bookingToCancel.id);

      // Optimistically update status to cancelled
      setBookings((prev) =>
        prev.map((b) => (b.id === bookingToCancel.id ? { ...b, status: 'cancelled' } : b))
      );

      setToastMessage('Booking cancelled. Slot has been released and waitlisted players were notified.');
      setTimeout(() => setToastMessage(null), 5000);
    } catch (err: any) {
      alert(err?.message || 'Failed to cancel booking.');
    } finally {
      setCancellingId(null);
      setBookingToCancel(null);
    }
  };

  const now = new Date();
  const upcomingBookings = bookings.filter(
    (b) => b.status === 'confirmed' && new Date(b.end_time) > now
  );
  const archivedBookings = bookings.filter(
    (b) => b.status === 'cancelled' || new Date(b.end_time) <= now
  );

  return (
    <div className="min-h-dvh bg-blush pb-28">
      <Header title="My Bookings" showLogo showAvatar />

      {/* User info & Sign out strip */}
      {user && (
        <div className="mx-5 mb-3 bg-white/70 backdrop-blur-sm rounded-xl px-4 py-2 flex items-center justify-between border border-blush-dark/30">
          <div className="flex items-center gap-2 overflow-hidden">
            <span className="w-2 h-2 rounded-full bg-success" />
            <span className="text-xs font-semibold text-plum truncate">{user.college_email}</span>
            <span className="text-[10px] font-bold tracking-wider uppercase text-plum-muted bg-blush px-1.5 py-0.5 rounded">
              {user.role}
            </span>
          </div>
          <button
            onClick={() => {
              logout();
              navigate('/login');
            }}
            className="text-xs font-bold text-danger hover:underline ml-2"
          >
            Logout
          </button>
        </div>
      )}

      <div className="px-5 pt-1 pb-2">
        <h2 className="text-2xl font-extrabold text-plum leading-tight">Mission Log</h2>
        <p className="text-sm text-plum-muted mt-1">Your upcoming campus sports bookings.</p>
      </div>

      {toastMessage && (
        <div className="mx-5 my-3 bg-plum text-white rounded-xl p-3 text-xs font-medium shadow-lg flex items-center gap-2">
          <span>✅</span>
          <span>{toastMessage}</span>
        </div>
      )}

      {!isAuthenticated ? (
        <div className="mx-5 mt-10 bg-white rounded-3xl p-8 text-center shadow-xs border border-blush-dark/20 flex flex-col items-center gap-4">
          <div className="w-16 h-16 rounded-full bg-blush flex items-center justify-center text-2xl">
            🔒
          </div>
          <div>
            <h3 className="text-lg font-bold text-plum">Authentication Required</h3>
            <p className="text-xs text-plum-muted mt-1 max-w-xs">
              Please sign in with your college email to manage your court reservations and waitlists.
            </p>
          </div>
          <button
            onClick={() => navigate('/login')}
            className="w-full max-w-xs bg-plum text-white py-3.5 rounded-xl text-sm font-semibold hover:bg-plum-light transition-colors"
          >
            Sign In Now
          </button>
        </div>
      ) : loading ? (
        <div className="py-16 flex flex-col items-center justify-center gap-3">
          <div className="w-8 h-8 border-3 border-plum/20 border-t-plum rounded-full animate-spin" />
          <p className="text-xs text-plum-muted">Loading your bookings...</p>
        </div>
      ) : (
        <>
          {/* Upcoming Missions */}
          <div className="px-5 mt-4">
            <h3 className="text-xs font-bold tracking-widest uppercase text-plum-muted mb-3">
              Upcoming Missions ({upcomingBookings.length})
            </h3>

            {upcomingBookings.length === 0 ? (
              <div className="bg-white rounded-2xl p-6 text-center shadow-xs border border-blush-dark/20">
                <p className="text-sm font-semibold text-plum">No upcoming bookings</p>
                <p className="text-xs text-plum-muted mt-1">Explore campus courts and book a session!</p>
                <button
                  onClick={() => navigate('/sports')}
                  className="mt-4 bg-plum text-white px-5 py-2.5 rounded-xl text-xs font-bold hover:bg-plum-light transition-colors"
                >
                  Browse Sports →
                </button>
              </div>
            ) : (
              <div className="flex flex-col gap-3">
                {upcomingBookings.map((b) => (
                  <BookingCard
                    key={b.id}
                    venue={`${b.facility_name} • ${b.court_label}`}
                    title={`${b.sport_type} Session`}
                    dateTime={formatBookingTime(b.start_time, b.end_time)}
                    status="confirmed"
                    playerCount={b.player_count}
                    showCancel={true}
                    isCancelling={cancellingId === b.id}
                    onCancel={() => setBookingToCancel(b)}
                  />
                ))}
              </div>
            )}
          </div>

          {/* Archive Log */}
          {archivedBookings.length > 0 && (
            <div className="px-5 mt-8">
              <h3 className="text-lg font-bold text-plum mb-3">Archive Log</h3>
              <div className="flex flex-col gap-2">
                {archivedBookings.map((b) => (
                  <div
                    key={b.id}
                    className="bg-white rounded-2xl px-5 py-4 flex items-center justify-between shadow-xs border border-blush-dark/20"
                  >
                    <div className="flex items-center gap-3">
                      <div className="w-10 h-10 rounded-full bg-blush flex items-center justify-center text-sm">
                        {b.status === 'cancelled' ? '❌' : '🎾'}
                      </div>
                      <div>
                        <p className="text-sm font-semibold text-plum">
                          {b.facility_name} • {b.court_label}
                        </p>
                        <p className="text-xs text-plum-muted">
                          {formatBookingTime(b.start_time, b.end_time)} •{' '}
                          <span
                            className={
                              b.status === 'cancelled' ? 'text-danger font-medium' : 'text-plum-muted'
                            }
                          >
                            {b.status === 'cancelled' ? 'Cancelled' : 'Completed'}
                          </span>
                        </p>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </>
      )}

      {/* Cancel Confirmation Modal */}
      {bookingToCancel && (
        <div className="fixed inset-0 z-50 flex items-center justify-center px-6">
          <div className="absolute inset-0 bg-plum/40 backdrop-blur-sm" onClick={() => setBookingToCancel(null)} />
          <div className="relative bg-white rounded-3xl p-6 w-full max-w-sm shadow-2xl animate-[scaleIn_0.2s_ease-out]">
            <div className="w-12 h-12 rounded-full bg-danger/10 text-danger flex items-center justify-center mx-auto mb-3 text-xl">
              ⚠️
            </div>
            <h3 className="text-lg font-bold text-plum text-center">Cancel Booking?</h3>
            <p className="text-xs text-plum-muted text-center mt-1">
              {bookingToCancel.facility_name} ({bookingToCancel.court_label})
            </p>
            <p className="text-xs text-plum-muted/80 text-center mt-3 bg-blush p-3 rounded-xl">
              Canceling will immediately free this slot and alert waitlisted players. This action cannot be undone.
            </p>

            <div className="flex gap-2 mt-5">
              <button
                onClick={() => setBookingToCancel(null)}
                className="flex-1 bg-blush text-plum py-3 rounded-xl font-semibold text-xs hover:bg-blush-dark transition-colors"
              >
                Keep Booking
              </button>
              <button
                onClick={handleConfirmCancel}
                disabled={cancellingId === bookingToCancel.id}
                className="flex-1 bg-danger text-white py-3 rounded-xl font-semibold text-xs hover:bg-danger/90 transition-colors disabled:opacity-50 flex items-center justify-center gap-1.5"
              >
                {cancellingId === bookingToCancel.id ? (
                  <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                ) : (
                  'Yes, Cancel'
                )}
              </button>
            </div>
          </div>
        </div>
      )}

      <BottomNav />
    </div>
  );
}
