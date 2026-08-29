import { useState, useEffect } from 'react';
import { useNavigate, useParams, useSearchParams } from 'react-router-dom';
import Header from '../components/Header';
import CourtCard from '../components/CourtCard';
import { facilityApi } from '../api/client';
import type { Court } from '../api/types';

interface CourtViewItem {
  id: string;
  name: string;
  surface: string;
  type: string;
  maxPlayers: number;
  status: 'available' | 'maintenance';
  nextAvailable: string;
}

const fallbackCourts: CourtViewItem[] = [
  { id: '1', name: 'Court 1', type: 'Indoor', surface: 'Hard Court', maxPlayers: 4, status: 'available', nextAvailable: '2:00 PM' },
  { id: '2', name: 'Court 2', type: 'Outdoor', surface: 'Clay', maxPlayers: 4, status: 'available', nextAvailable: '3:30 PM' },
  { id: '3', name: 'Court 3', type: 'Indoor', surface: 'Hard Court', maxPlayers: 4, status: 'maintenance', nextAvailable: 'Tomorrow' },
  { id: '4', name: 'Court 4', type: 'Outdoor', surface: 'Grass', maxPlayers: 4, status: 'available', nextAvailable: '4:00 PM' },
];

export default function CourtDetailsPage() {
  const { sport } = useParams<{ sport: string }>();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const [courts, setCourts] = useState<CourtViewItem[]>(fallbackCourts);
  const [selectedCourtId, setSelectedCourtId] = useState<string>('1');
  const [facilityName, setFacilityName] = useState<string>(searchParams.get('facility') || 'Sports Complex');
  const [facilityId, setFacilityId] = useState<string>(sport || '');
  const [loading, setLoading] = useState<boolean>(true);

  const isUUID = (str?: string) =>
    !!str && /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(str);

  useEffect(() => {
    async function loadCourts() {
      if (!sport) {
        setLoading(false);
        return;
      }

      try {
        let fId = sport;

        // If sport parameter is not a UUID, try to find facility by sport_type
        if (!isUUID(sport)) {
          const facilities = await facilityApi.listFacilities();
          const match = facilities.find(
            (f) =>
              f.sport_type.toLowerCase() === sport.toLowerCase() ||
              f.sport_type.toLowerCase().replace(/\s+/g, '-') === sport.toLowerCase()
          );

          if (match) {
            fId = match.id;
            setFacilityId(match.id);
            setFacilityName(match.name);
          }
        }

        if (isUUID(fId)) {
          const backendCourts = await facilityApi.listCourts(fId);
          if (backendCourts && backendCourts.length > 0) {
            const mapped: CourtViewItem[] = backendCourts.map((c: Court, idx: number) => ({
              id: c.id,
              name: c.label || `Court ${idx + 1}`,
              surface: idx % 2 === 0 ? 'Synthetic Hard' : 'Clay Surface',
              type: idx % 2 === 0 ? 'Indoor' : 'Outdoor',
              maxPlayers: c.max_players || 4,
              status: 'available',
              nextAvailable: 'Next slot available',
            }));

            setCourts(mapped);
            setSelectedCourtId(mapped[0].id);
          }
        }
      } catch (err) {
        console.warn('Could not load courts from backend, using fallback:', err);
      } finally {
        setLoading(false);
      }
    }

    loadCourts();
  }, [sport]);

  const selectedCourt = courts.find((c) => c.id === selectedCourtId) || courts[0];

  const handleContinue = () => {
    const params = new URLSearchParams();
    if (selectedCourt) {
      params.set('courtName', selectedCourt.name);
    }
    if (facilityName) {
      params.set('facilityName', facilityName);
    }
    params.set('maxPlayers', String(selectedCourt?.maxPlayers || 4));

    navigate(`/book/${facilityId || sport}/court/${selectedCourtId}/time?${params.toString()}`);
  };

  return (
    <div className="min-h-dvh bg-blush pb-28 flex flex-col">
      <Header title="Court Details" showBack showAvatar />

      <div className="px-5 pt-2 flex-1">
        <h2 className="text-2xl font-extrabold text-plum leading-tight">Select a Court</h2>
        <p className="text-sm text-plum-muted mt-1">{facilityName} • Campus Sports</p>

        {loading ? (
          <div className="py-12 flex flex-col items-center justify-center gap-3">
            <div className="w-8 h-8 border-3 border-plum/20 border-t-plum rounded-full animate-spin" />
            <p className="text-xs text-plum-muted">Loading available courts...</p>
          </div>
        ) : (
          <div className="mt-5 flex flex-col gap-3">
            {courts.map((court) => (
              <CourtCard
                key={court.id}
                name={court.name}
                surface={court.surface}
                type={court.type}
                status={court.status}
                nextAvailable={court.nextAvailable}
                selected={selectedCourtId === court.id}
                onClick={() => {
                  if (court.status === 'available') {
                    setSelectedCourtId(court.id);
                  }
                }}
              />
            ))}
          </div>
        )}
      </div>

      {/* Sticky Bottom Button */}
      <div className="fixed bottom-0 left-1/2 -translate-x-1/2 w-full max-w-[430px] px-5 pb-[max(1.25rem,env(safe-area-inset-bottom))] pt-3 bg-gradient-to-t from-blush via-blush to-blush/0">
        <button
          onClick={handleContinue}
          disabled={loading || courts.length === 0}
          className="w-full bg-plum text-white py-4 rounded-2xl font-semibold text-base flex items-center justify-center gap-2 hover:bg-plum-light transition-colors active:scale-[0.98] shadow-lg shadow-plum/20 disabled:opacity-50"
        >
          Continue with {selectedCourt?.name || 'Court'}
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <line x1="5" y1="12" x2="19" y2="12" />
            <polyline points="12 5 19 12 12 19" />
          </svg>
        </button>
      </div>
    </div>
  );
}
