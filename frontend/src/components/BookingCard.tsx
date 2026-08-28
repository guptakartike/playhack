import StatusBadge from './StatusBadge';

interface BookingCardProps {
  venue: string;
  title: string;
  dateTime: string;
  status: 'confirmed' | 'pending' | 'completed';
  opponent?: {
    name: string;
    avatar?: string;
  };
  showCheckIn?: boolean;
  onCheckIn?: () => void;
}

export default function BookingCard({
  venue,
  title,
  dateTime,
  status,
  opponent,
  showCheckIn = false,
  onCheckIn,
}: BookingCardProps) {
  return (
    <div className="bg-white rounded-2xl p-5 shadow-sm">
      <div className="flex items-start justify-between mb-1">
        <span className="text-xs font-semibold tracking-widest uppercase text-plum-muted">{venue}</span>
        <StatusBadge status={status} />
      </div>

      <h3 className="text-xl font-bold text-plum mt-1 leading-tight">{title}</h3>
      <p className="text-sm text-plum-muted mt-1">{dateTime}</p>

      {(opponent || showCheckIn) && (
        <div className="flex items-center justify-between mt-4 pt-3 border-t border-blush-dark/30">
          {opponent && (
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-blush overflow-hidden flex items-center justify-center">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#9B8FA6" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
                  <circle cx="12" cy="7" r="4" />
                </svg>
              </div>
              <div>
                <p className="text-xs text-plum-muted">Opponent</p>
                <p className="text-sm font-semibold text-plum">{opponent.name}</p>
              </div>
            </div>
          )}

          {showCheckIn && (
            <button
              onClick={onCheckIn}
              className="flex items-center gap-2 bg-plum text-white px-4 py-2 rounded-full text-sm font-semibold hover:bg-plum-light transition-colors active:scale-[0.97]"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z" />
                <circle cx="12" cy="10" r="3" />
              </svg>
              Check-in
            </button>
          )}
        </div>
      )}
    </div>
  );
}
