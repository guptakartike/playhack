import { useState } from 'react';
import { useParams } from 'react-router-dom';
import Header from '../components/Header';
import TimeSlot from '../components/TimeSlot';

interface DayItem {
  day: string;
  date: number;
  key: string;
}

const days: DayItem[] = [
  { day: 'MON', date: 12, key: 'mon-12' },
  { day: 'TUE', date: 13, key: 'tue-13' },
  { day: 'WED', date: 14, key: 'wed-14' },
  { day: 'THU', date: 15, key: 'thu-15' },
];

interface SlotItem {
  time: string;
  type: 'standard' | 'peak';
  state: 'available' | 'selected' | 'locked';
}

const morningSlots: SlotItem[] = [
  { time: '6:00 - 7:00', type: 'standard', state: 'available' },
  { time: '7:15 - 8:15', type: 'peak', state: 'locked' },
  { time: '8:30 - 9:30', type: 'peak', state: 'selected' },
  { time: '9:45 - 10:45', type: 'peak', state: 'available' },
  { time: '11:00 - 12:00', type: 'standard', state: 'available' },
];

const afternoonSlots: SlotItem[] = [
  { time: '12:15 - 13:15', type: 'standard', state: 'available' },
  { time: '13:30 - 14:30', type: 'standard', state: 'available' },
  { time: '14:45 - 15:45', type: 'standard', state: 'available' },
  { time: '16:00 - 17:00', type: 'standard', state: 'available' },
  { time: '17:15 - 18:15', type: 'peak', state: 'available' },
  { time: '18:30 - 19:30', type: 'peak', state: 'available' },
  { time: '19:45 - 20:45', type: 'peak', state: 'available' },
];

export default function SelectTimePage() {
  const { courtId } = useParams<{ sport: string; courtId: string }>();
  const [selectedDay, setSelectedDay] = useState('tue-13');
  const [selectedSlot, setSelectedSlot] = useState('8:30 - 9:30');
  const [showConfirm, setShowConfirm] = useState(false);

  const courtLabel = `Court ${courtId}`;
  const selectedDayData = days.find(d => d.key === selectedDay);

  return (
    <div className="min-h-dvh bg-blush pb-28">
      <Header title="Book Court" showLogo showAvatar />

      <div className="px-5 pt-2">
        <h2 className="text-2xl font-extrabold text-plum leading-tight">Select a Time</h2>
        <p className="text-sm text-plum-muted mt-1">{courtLabel} • Hard Court</p>
      </div>

      {/* Month + Navigation */}
      <div className="px-5 mt-5 flex items-center justify-between">
        <h3 className="text-lg font-bold text-plum">October</h3>
        <div className="flex gap-2">
          <button className="w-8 h-8 rounded-full bg-plum/10 flex items-center justify-center hover:bg-plum/20 transition-colors">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#271F30" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <polyline points="15 18 9 12 15 6" />
            </svg>
          </button>
          <button className="w-8 h-8 rounded-full bg-plum/10 flex items-center justify-center hover:bg-plum/20 transition-colors">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#271F30" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <polyline points="9 18 15 12 9 6" />
            </svg>
          </button>
        </div>
      </div>

      {/* Day Selector */}
      <div className="px-5 mt-4 flex gap-3 overflow-x-auto hide-scrollbar">
        {days.map((d) => (
          <button
            key={d.key}
            onClick={() => setSelectedDay(d.key)}
            className={`flex flex-col items-center min-w-[60px] py-3 px-3 rounded-2xl transition-all duration-300 ${
              selectedDay === d.key
                ? 'bg-plum text-white shadow-md shadow-plum/20'
                : 'bg-white text-plum hover:bg-blush-dark/30'
            }`}
          >
            <span className={`text-[10px] font-semibold tracking-wider ${selectedDay === d.key ? 'text-white/70' : 'text-plum-muted'}`}>
              {d.day}
            </span>
            <span className="text-lg font-bold mt-0.5">{d.date}</span>
          </button>
        ))}
      </div>

      {/* Legend */}
      <div className="px-5 mt-6 flex items-center gap-5">
        <div className="flex items-center gap-2">
          <span className="w-3 h-3 rounded-full bg-white border border-blush-dark" />
          <span className="text-xs font-medium text-plum-muted">Standard</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="w-3.5 h-3.5 rounded-full bg-[#E05A2B]/15 border-2 border-[#E05A2B]" />
          <span className="text-xs font-medium text-plum-muted">Peak</span>
        </div>
      </div>

      {/* Morning Slots */}
      <div className="px-5 mt-5">
        <p className="text-xs font-semibold tracking-widest uppercase text-plum-muted mb-3">Morning</p>
        <div className="grid grid-cols-2 gap-2.5">
          {morningSlots.map((slot) => (
            <TimeSlot
              key={slot.time}
              time={slot.time}
              type={slot.type}
              state={slot.time === selectedSlot ? 'selected' : slot.state === 'selected' ? 'available' : slot.state}
              onClick={() => {
                if (slot.state !== 'locked') setSelectedSlot(slot.time);
              }}
            />
          ))}
        </div>
      </div>

      {/* Afternoon Slots */}
      <div className="px-5 mt-5">
        <p className="text-xs font-semibold tracking-widest uppercase text-plum-muted mb-3">Afternoon</p>
        <div className="grid grid-cols-2 gap-2.5">
          {afternoonSlots.map((slot) => (
            <TimeSlot
              key={slot.time}
              time={slot.time}
              type={slot.type}
              state={slot.time === selectedSlot ? 'selected' : slot.state}
              onClick={() => setSelectedSlot(slot.time)}
            />
          ))}
        </div>
      </div>

      {/* Sticky Bottom Button */}
      <div className="fixed bottom-0 left-1/2 -translate-x-1/2 w-full max-w-[430px] px-5 pb-[max(1.25rem,env(safe-area-inset-bottom))] pt-3 bg-gradient-to-t from-blush via-blush to-blush/0">
        <button
          onClick={() => setShowConfirm(true)}
          className="w-full bg-plum text-white py-4 rounded-2xl font-semibold text-base flex items-center justify-center gap-2 hover:bg-plum-light transition-colors active:scale-[0.98] shadow-lg shadow-plum/20"
        >
          Book Slot
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <line x1="5" y1="12" x2="19" y2="12" />
            <polyline points="12 5 19 12 12 19" />
          </svg>
        </button>
      </div>

      {/* Confirmation Popup */}
      {showConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center px-6">
          <div className="absolute inset-0 bg-plum/40 backdrop-blur-sm" onClick={() => setShowConfirm(false)} />
          <div className="relative bg-white rounded-3xl p-6 w-full max-w-sm shadow-2xl animate-[scaleIn_0.2s_ease-out]">
            {/* Success Icon */}
            <div className="w-14 h-14 rounded-full bg-success/10 flex items-center justify-center mx-auto mb-4">
              <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#22C55E" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <polyline points="20 6 9 17 4 12" />
              </svg>
            </div>

            <h3 className="text-lg font-bold text-plum text-center">Booking Confirmed!</h3>
            <p className="text-sm text-plum-muted text-center mt-1">Your slot has been booked successfully.</p>

            {/* Booking Details */}
            <div className="bg-blush rounded-2xl p-4 mt-5 flex flex-col gap-2.5">
              <div className="flex items-center justify-between">
                <span className="text-xs font-medium text-plum-muted">Court</span>
                <span className="text-sm font-semibold text-plum">{courtLabel} • Hard Court</span>
              </div>
              <div className="border-t border-blush-dark/30" />
              <div className="flex items-center justify-between">
                <span className="text-xs font-medium text-plum-muted">Date</span>
                <span className="text-sm font-semibold text-plum">
                  {selectedDayData ? `${selectedDayData.day}, October ${selectedDayData.date}` : '—'}
                </span>
              </div>
              <div className="border-t border-blush-dark/30" />
              <div className="flex items-center justify-between">
                <span className="text-xs font-medium text-plum-muted">Time</span>
                <span className="text-sm font-semibold text-plum">{selectedSlot}</span>
              </div>
            </div>

            <button
              onClick={() => setShowConfirm(false)}
              className="w-full bg-plum text-white py-3.5 rounded-xl font-semibold text-sm mt-5 hover:bg-plum-light transition-colors active:scale-[0.98]"
            >
              Done
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
