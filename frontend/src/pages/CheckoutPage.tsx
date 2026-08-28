import { useNavigate } from 'react-router-dom';
import Header from '../components/Header';

const players = [
  { id: 1, name: 'You', isStar: true },
  { id: 2, name: 'Alex' },
  { id: 3, name: 'Jordan' },
];

const lineItems = [
  { label: 'Court Hire (90 min)', amount: 45.00 },
  { label: 'Racket Rental (x2)', amount: 10.00 },
  { label: 'Service Fee', amount: 2.50 },
];

export default function CheckoutPage() {
  const navigate = useNavigate();
  const total = lineItems.reduce((sum, item) => sum + item.amount, 0);

  return (
    <div className="min-h-dvh bg-blush pb-28">
      <Header title="Checkout" showBack showAvatar />

      {/* Player Squad */}
      <div className="mx-5 mt-2 bg-gradient-to-br from-plum/90 to-plum rounded-2xl p-5 shadow-lg shadow-plum/10">
        <h3 className="text-sm font-semibold text-white/70">Player Squad</h3>
        <div className="flex items-center gap-3 mt-4">
          {players.map((player) => (
            <div key={player.id} className="relative">
              <div className="w-12 h-12 rounded-full bg-white/20 overflow-hidden flex items-center justify-center ring-2 ring-white/30">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
                  <circle cx="12" cy="7" r="4" />
                </svg>
              </div>
              {player.isStar && (
                <span className="absolute -bottom-0.5 -left-0.5 w-5 h-5 bg-warning rounded-full flex items-center justify-center text-[8px]">⭐</span>
              )}
            </div>
          ))}
          {/* Add Player Button */}
          <button className="w-12 h-12 rounded-full border-2 border-dashed border-white/30 flex items-center justify-center hover:border-white/50 transition-colors">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" opacity="0.6">
              <line x1="12" y1="5" x2="12" y2="19" />
              <line x1="5" y1="12" x2="19" y2="12" />
            </svg>
          </button>
        </div>
        <p className="text-xs text-white/50 mt-3">3/4 players confirmed. You can invite more later.</p>
      </div>

      {/* Booking Details */}
      <div className="px-5 mt-6">
        <h3 className="text-lg font-bold text-plum mb-3">Booking Details</h3>
        <div className="bg-white rounded-2xl p-4 flex items-center gap-4 shadow-sm">
          <div className="w-14 h-14 rounded-xl bg-blush flex items-center justify-center shrink-0">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#271F30" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
              <circle cx="12" cy="12" r="10" />
              <path d="M18.09 6.41A9.93 9.93 0 0 0 12 4a9.93 9.93 0 0 0-6.09 2.41" />
              <path d="M5.91 17.59A9.93 9.93 0 0 0 12 20a9.93 9.93 0 0 0 6.09-2.41" />
              <line x1="12" y1="2" x2="12" y2="22" />
            </svg>
          </div>
          <div>
            <h4 className="font-semibold text-plum">Oasis Padel Club</h4>
            <div className="flex items-center gap-1.5 text-sm text-plum-muted mt-0.5">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
                <line x1="16" y1="2" x2="16" y2="6" />
                <line x1="8" y1="2" x2="8" y2="6" />
                <line x1="3" y1="10" x2="21" y2="10" />
              </svg>
              Tomorrow, 18:00 - 19:30
            </div>
            <div className="flex items-center gap-1.5 text-sm text-plum-muted mt-0.5">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z" />
                <circle cx="12" cy="10" r="3" />
              </svg>
              Court 4 (Indoor)
            </div>
          </div>
        </div>
      </div>

      {/* Payment Summary */}
      <div className="px-5 mt-6">
        <h3 className="text-lg font-bold text-plum mb-3">Payment Summary</h3>
        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <div className="flex flex-col gap-3">
            {lineItems.map((item) => (
              <div key={item.label} className="flex items-center justify-between">
                <span className="text-sm text-plum/80">{item.label}</span>
                <span className="text-sm font-semibold text-plum">${item.amount.toFixed(2)}</span>
              </div>
            ))}
          </div>

          <div className="border-t border-blush-dark/30 mt-4 pt-4 flex items-center justify-between">
            <span className="text-base font-bold text-plum">Total</span>
            <span className="text-2xl font-extrabold text-plum">${total.toFixed(2)}</span>
          </div>

          <div className="mt-4 flex items-start gap-2 bg-blush rounded-xl p-3">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#9B8FA6" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="shrink-0 mt-0.5">
              <circle cx="12" cy="12" r="10" />
              <line x1="12" y1="16" x2="12" y2="12" />
              <line x1="12" y1="8" x2="12.01" y2="8" />
            </svg>
            <p className="text-xs text-plum-muted">Split with squad available after booking</p>
          </div>
        </div>
      </div>

      {/* Sticky Confirm Button */}
      <div className="fixed bottom-0 left-1/2 -translate-x-1/2 w-full max-w-[430px] px-5 pb-[max(1.25rem,env(safe-area-inset-bottom))] pt-3 bg-gradient-to-t from-blush via-blush to-blush/0">
        <button
          onClick={() => navigate('/bookings')}
          className="w-full bg-plum text-white py-4 rounded-2xl font-semibold text-base flex items-center justify-center gap-2 hover:bg-plum-light transition-colors active:scale-[0.98] shadow-lg shadow-plum/20"
        >
          Confirm & Pay
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <line x1="5" y1="12" x2="19" y2="12" />
            <polyline points="12 5 19 12 12 19" />
          </svg>
        </button>
      </div>
    </div>
  );
}
