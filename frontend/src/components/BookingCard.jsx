import React from 'react';
import { Calendar, Clock, Users, XCircle, Trophy, CheckCircle2 } from 'lucide-react';

export function formatBookingDateTime(startTimeStr, endTimeStr) {
  if (!startTimeStr || !endTimeStr) return '';
  
  const start = new Date(startTimeStr);
  const end = new Date(endTimeStr);
  const now = new Date();

  const isToday = start.toDateString() === now.toDateString();

  const tomorrow = new Date(now);
  tomorrow.setDate(tomorrow.getDate() + 1);
  const isTomorrow = start.toDateString() === tomorrow.toDateString();

  const timeFormatOptions = { hour: 'numeric', minute: '2-digit', hour12: true };
  const startTimeFormatted = start.toLocaleTimeString('en-US', timeFormatOptions);
  const endTimeFormatted = end.toLocaleTimeString('en-US', timeFormatOptions);

  let datePrefix = '';
  if (isToday) {
    datePrefix = 'Today';
  } else if (isTomorrow) {
    datePrefix = 'Tomorrow';
  } else {
    datePrefix = start.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  }

  return `${datePrefix}, ${startTimeFormatted} – ${endTimeFormatted}`;
}

export function computeBookingStatus(booking) {
  if (booking.status === 'cancelled') {
    return 'cancelled';
  }
  
  const start = new Date(booking.start_time);
  const now = new Date();
  
  if (start > now) {
    return 'upcoming';
  }
  return 'completed';
}

export default function BookingCard({ booking, onCancelClick }) {
  const status = computeBookingStatus(booking);
  const formattedDateTime = formatBookingDateTime(booking.start_time, booking.end_time);

  return (
    <div className={`relative group rounded-2xl border transition-all duration-300 overflow-hidden ${
      status === 'upcoming' 
        ? 'bg-[#14141E] border-[#2A2A3C] hover:border-[#F5793A]/50 shadow-lg hover:shadow-[#F5793A]/5' 
        : status === 'completed'
        ? 'bg-[#12121A] border-[#1F1F2E] opacity-80'
        : 'bg-[#12121A]/60 border-rose-950/40 opacity-60'
    }`}>
      {/* Decorative top accent glow for upcoming */}
      {status === 'upcoming' && (
        <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-[#F5793A] to-orange-400 opacity-90" />
      )}

      <div className="p-5 sm:p-6 space-y-4">
        {/* Header: Facility + Status Badge */}
        <div className="flex items-start justify-between gap-4">
          <div>
            <div className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-[#94A3B8] mb-1">
              <Trophy className="w-3.5 h-3.5 text-[#F5793A]" />
              <span>{booking.sport_type || 'Sports Facility'}</span>
            </div>
            <h3 className={`text-lg font-bold tracking-tight text-[#F1F5F9] ${
              status === 'cancelled' ? 'line-through text-slate-400' : ''
            }`}>
              {booking.facility_name} — <span className="text-[#F5793A]">{booking.court_label}</span>
            </h3>
          </div>

          {/* Status Chips */}
          <div>
            {status === 'upcoming' && (
              <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-[#F5793A]/10 text-[#F5793A] border border-[#F5793A]/40 shadow-sm">
                <span className="w-1.5 h-1.5 rounded-full bg-[#F5793A] animate-pulse" />
                Upcoming
              </span>
            )}
            {status === 'completed' && (
              <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium bg-slate-800/60 text-slate-400 border border-slate-700/60">
                <CheckCircle2 className="w-3.5 h-3.5" />
                Completed
              </span>
            )}
            {status === 'cancelled' && (
              <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium bg-rose-950/40 text-rose-400 border border-rose-900/50">
                <XCircle className="w-3.5 h-3.5" />
                Cancelled
              </span>
            )}
          </div>
        </div>

        {/* Details Grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-2 border-t border-[#1F1F2E]">
          <div className="flex items-center gap-2 text-sm text-[#94A3B8]">
            <Clock className="w-4 h-4 text-[#F5793A]" />
            <span className="font-medium text-[#F1F5F9]">{formattedDateTime}</span>
          </div>

          <div className="flex items-center gap-2 text-sm text-[#94A3B8]">
            <Users className="w-4 h-4 text-[#F5793A]" />
            <span>
              <strong className="text-[#F1F5F9] font-medium">{booking.player_count}</strong> {booking.player_count === 1 ? 'Player' : 'Players'}
            </span>
          </div>
        </div>

        {/* Action Button: Cancel (only on upcoming) */}
        {status === 'upcoming' && (
          <div className="pt-2 flex justify-end">
            <button
              onClick={() => onCancelClick(booking)}
              className="inline-flex items-center gap-1.5 px-4 py-2 rounded-xl text-xs font-semibold text-rose-400 hover:text-white bg-rose-950/20 hover:bg-rose-600 border border-rose-900/40 hover:border-rose-600 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-rose-500/40"
            >
              <XCircle className="w-3.5 h-3.5" />
              Cancel Booking
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
