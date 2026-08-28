import React, { useState, useEffect, useMemo } from 'react';
import { Trophy, CalendarDays, ArrowRight, RefreshCw, AlertCircle } from 'lucide-react';
import BookingCard, { computeBookingStatus } from '../components/BookingCard';
import BookingCardSkeleton from '../components/BookingCardSkeleton';
import ConfirmModal from '../components/ConfirmModal';
import Toast from '../components/Toast';
import { fetchMyBookings, cancelBookingApi } from '../api/client';

export default function MyBookingsPage({ onNavigateBrowse, onNavigateLogin }) {
  const [bookings, setBookings] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  // Cancellation Modal & Toast states
  const [selectedBookingForCancel, setSelectedBookingForCancel] = useState(null);
  const [isCancelling, setIsCancelling] = useState(false);
  const [toast, setToast] = useState({ type: 'error', message: '' });

  const loadBookings = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const data = await fetchMyBookings();
      setBookings(data || []);
    } catch (err) {
      if (err.message === 'UNAUTHORIZED') {
        if (onNavigateLogin) onNavigateLogin();
        return;
      }
      setError(err.message || 'Failed to load bookings');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadBookings();
  }, []);

  // Sort: Upcoming (soonest first), then Completed & Cancelled (most recent first)
  const sortedBookings = useMemo(() => {
    if (!bookings.length) return [];

    const upcoming = [];
    const pastOrCancelled = [];

    bookings.forEach((b) => {
      const st = computeBookingStatus(b);
      if (st === 'upcoming') {
        upcoming.push(b);
      } else {
        pastOrCancelled.push(b);
      }
    });

    // Upcoming: soonest first
    upcoming.sort((a, b) => new Date(a.start_time) - new Date(b.start_time));

    // Past or Cancelled: most recent start_time first
    pastOrCancelled.sort((a, b) => new Date(b.start_time) - new Date(a.start_time));

    return [...upcoming, ...pastOrCancelled];
  }, [bookings]);

  const handleConfirmCancel = async () => {
    if (!selectedBookingForCancel) return;

    setIsCancelling(true);
    try {
      await cancelBookingApi(selectedBookingForCancel.id);

      // In-place update status to cancelled
      setBookings((prev) =>
        prev.map((b) =>
          b.id === selectedBookingForCancel.id ? { ...b, status: 'cancelled' } : b
        )
      );

      setToast({ type: 'success', message: 'Booking cancelled successfully.' });
      setSelectedBookingForCancel(null);
    } catch (err) {
      if (err.message === 'UNAUTHORIZED') {
        if (onNavigateLogin) onNavigateLogin();
        return;
      }
      setToast({ type: 'error', message: err.message || 'Could not cancel booking.' });
    } finally {
      setIsCancelling(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto px-4 py-8 sm:px-6 space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between gap-4 pb-6 border-b border-[#1F1F2E]">
        <div>
          <div className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-[#F5793A] mb-1">
            <CalendarDays className="w-4 h-4" />
            <span>Dashboard</span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight text-[#F1F5F9]">
            My Bookings
          </h1>
          <p className="text-xs sm:text-sm text-[#94A3B8] mt-1">
            Manage your court reservations and upcoming sessions.
          </p>
        </div>

        <button
          onClick={loadBookings}
          disabled={isLoading}
          className="p-2.5 rounded-xl bg-[#14141E] hover:bg-[#1A1A26] border border-[#2A2A3C] text-[#94A3B8] hover:text-[#F1F5F9] transition-all disabled:opacity-50"
          title="Refresh Bookings"
        >
          <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
        </button>
      </div>

      {/* Main Content Area */}
      {isLoading ? (
        /* Skeleton Loaders */
        <div className="space-y-4">
          <BookingCardSkeleton />
          <BookingCardSkeleton />
          <BookingCardSkeleton />
        </div>
      ) : error ? (
        /* Error State */
        <div className="p-6 rounded-2xl bg-rose-950/20 border border-rose-900/40 text-center space-y-3">
          <AlertCircle className="w-8 h-8 text-rose-400 mx-auto" />
          <h3 className="text-sm font-semibold text-rose-200">{error}</h3>
          <button
            onClick={loadBookings}
            className="px-4 py-2 rounded-xl text-xs font-semibold bg-rose-900/40 hover:bg-rose-900/60 text-rose-200 transition-colors"
          >
            Try Again
          </button>
        </div>
      ) : sortedBookings.length === 0 ? (
        /* Empty State */
        <div className="py-16 text-center rounded-3xl bg-[#14141E]/40 border border-[#1F1F2E] p-8 space-y-4">
          <div className="w-16 h-16 rounded-full bg-[#1A1A26] border border-[#2A2A3C] flex items-center justify-center mx-auto text-[#F5793A]">
            <Trophy className="w-8 h-8" />
          </div>
          <div className="space-y-1">
            <h3 className="text-lg font-bold text-[#F1F5F9]">No bookings yet</h3>
            <p className="text-xs sm:text-sm text-[#94A3B8] max-w-sm mx-auto">
              You haven't reserved any sports courts. Browse active campus facilities and book your slot now!
            </p>
          </div>
          <button
            onClick={onNavigateBrowse}
            className="inline-flex items-center gap-2 px-5 py-2.5 rounded-xl text-xs font-bold text-white bg-[#F5793A] hover:bg-[#E06728] transition-all shadow-lg shadow-[#F5793A]/20"
          >
            <span>Browse Facilities</span>
            <ArrowRight className="w-4 h-4" />
          </button>
        </div>
      ) : (
        /* Sorted Booking Cards */
        <div className="space-y-4">
          {sortedBookings.map((booking) => (
            <BookingCard
              key={booking.id}
              booking={booking}
              onCancelClick={setSelectedBookingForCancel}
            />
          ))}
        </div>
      )}

      {/* Confirmation Modal */}
      <ConfirmModal
        isOpen={Boolean(selectedBookingForCancel)}
        title="Cancel this booking?"
        message={
          selectedBookingForCancel
            ? `Cancel reservation for ${selectedBookingForCancel.facility_name} (${selectedBookingForCancel.court_label})?`
            : ''
        }
        isLoading={isCancelling}
        onConfirm={handleConfirmCancel}
        onCancel={() => setSelectedBookingForCancel(null)}
      />

      {/* Inline Toast */}
      <Toast
        type={toast.type}
        message={toast.message}
        onClose={() => setToast({ type: 'error', message: '' })}
      />
    </div>
  );
}
