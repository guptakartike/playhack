import Header from '../components/Header';
import BottomNav from '../components/BottomNav';
import BookingCard from '../components/BookingCard';

const upcomingBookings = [
  {
    id: 1,
    venue: 'Alpha Court',
    title: 'Singles Match vs. Dave',
    dateTime: 'Tomorrow, 18:00 - 19:30',
    status: 'confirmed' as const,
    opponent: { name: 'Dave "The Serve"' },
    showCheckIn: true,
  },
  {
    id: 2,
    venue: 'Bravo Court',
    title: 'Doubles Practice',
    dateTime: 'Sat, Oct 28, 10:00 - 12:00',
    status: 'pending' as const,
  },
];

const archivedBookings = [
  { id: 3, venue: 'Charlie Court', date: 'Oct 15', status: 'Completed' },
  { id: 4, venue: 'Alpha Court', date: 'Oct 02', status: 'Completed' },
];

export default function MyBookingsPage() {
  return (
    <div className="min-h-dvh bg-blush pb-24">
      <Header title="My Bookings" showLogo showAvatar />

      <div className="px-5 pt-2 pb-2">
        <h2 className="text-2xl font-extrabold text-plum leading-tight">Mission Log</h2>
        <p className="text-sm text-plum-muted mt-1">Your upcoming tennis deployments.</p>
      </div>

      {/* Upcoming Missions */}
      <div className="px-5 mt-4">
        <h3 className="text-sm font-semibold tracking-widest uppercase text-plum-muted mb-3">Upcoming Missions</h3>
        <div className="flex flex-col gap-3">
          {upcomingBookings.map((booking) => (
            <BookingCard
              key={booking.id}
              venue={booking.venue}
              title={booking.title}
              dateTime={booking.dateTime}
              status={booking.status}
              opponent={booking.opponent}
              showCheckIn={booking.showCheckIn}
              onCheckIn={() => {}}
            />
          ))}
        </div>
      </div>

      {/* Archive Log */}
      <div className="px-5 mt-8">
        <h3 className="text-lg font-bold text-plum mb-3">Archive Log</h3>
        <div className="flex flex-col gap-2">
          {archivedBookings.map((booking) => (
            <div key={booking.id} className="bg-white rounded-2xl px-5 py-4 flex items-center justify-between shadow-sm">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-full bg-blush flex items-center justify-center">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#9B8FA6" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <circle cx="12" cy="12" r="10" />
                    <polyline points="12 6 12 12 16 14" />
                  </svg>
                </div>
                <div>
                  <p className="text-sm font-semibold text-plum">{booking.venue}</p>
                  <p className="text-xs text-plum-muted">{booking.date} • {booking.status}</p>
                </div>
              </div>
              <button className="text-xs font-semibold text-plum hover:text-plum-light transition-colors">
                View Stats
              </button>
            </div>
          ))}
        </div>
      </div>

      <BottomNav />
    </div>
  );
}
