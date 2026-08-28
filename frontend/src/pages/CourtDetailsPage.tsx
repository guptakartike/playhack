import { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import Header from '../components/Header';
import CourtCard from '../components/CourtCard';

const courts = [
  { id: 1, name: 'Court 1', type: 'Indoor', surface: 'Hard Court', status: 'available' as const, nextAvailable: '2:00 PM' },
  { id: 2, name: 'Court 2', type: 'Outdoor', surface: 'Clay', status: 'available' as const, nextAvailable: '3:30 PM' },
  { id: 3, name: 'Court 3', type: 'Indoor', surface: 'Hard Court', status: 'maintenance' as const, maintenanceNote: 'Back online tomorrow' },
  { id: 4, name: 'Court 4', type: 'Outdoor', surface: 'Grass', status: 'available' as const, nextAvailable: '4:00 PM' },
];

export default function CourtDetailsPage() {
  const { sport } = useParams<{ sport: string }>();
  const navigate = useNavigate();
  const [selectedCourt, setSelectedCourt] = useState<number>(1);

  return (
    <div className="min-h-dvh bg-blush pb-28 flex flex-col">
      <Header title="Court Details" showBack showAvatar />

      <div className="px-5 pt-2 flex-1">
        <h2 className="text-2xl font-extrabold text-plum leading-tight">Select a Court</h2>
        <p className="text-sm text-plum-muted mt-1">Choose your preferred playing surface.</p>

        <div className="mt-5 flex flex-col gap-3">
          {courts.map((court) => (
            <CourtCard
              key={court.id}
              name={court.name}
              surface={court.surface}
              type={court.type}
              status={court.status}
              nextAvailable={court.nextAvailable}
              maintenanceNote={court.maintenanceNote}
              selected={selectedCourt === court.id}
              onClick={() => {
                if (court.status === 'available') {
                  setSelectedCourt(court.id);
                }
              }}
            />
          ))}
        </div>
      </div>

      {/* Sticky Bottom Button */}
      <div className="fixed bottom-0 left-1/2 -translate-x-1/2 w-full max-w-[430px] px-5 pb-[max(1.25rem,env(safe-area-inset-bottom))] pt-3 bg-gradient-to-t from-blush via-blush to-blush/0">
        <button
          onClick={() => navigate(`/book/${sport}/court/${selectedCourt}/time`)}
          className="w-full bg-plum text-white py-4 rounded-2xl font-semibold text-base flex items-center justify-center gap-2 hover:bg-plum-light transition-colors active:scale-[0.98] shadow-lg shadow-plum/20"
        >
          Continue with {courts.find(c => c.id === selectedCourt)?.name}
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <line x1="5" y1="12" x2="19" y2="12" />
            <polyline points="12 5 19 12 12 19" />
          </svg>
        </button>
      </div>
    </div>
  );
}
