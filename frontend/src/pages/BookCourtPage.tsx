import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import Header from '../components/Header';
import BottomNav from '../components/BottomNav';
import SportCard from '../components/SportCard';
import { facilityApi } from '../api/client';
import type { Facility } from '../api/types';

const icon = (d: string) => (
  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#271F30" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
    <path d={d} />
  </svg>
);

interface SportMeta {
  name: string;
  defaultCourts: number;
  icon: React.ReactNode;
}

const sportsList: SportMeta[] = [
  { name: 'Football',      defaultCourts: 2, icon: icon('M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20zm0 3l2.5 2-1 3h-3l-1-3zm-5.5 5.5L9 9l1 3-2 2.5H5.5zm3 6L11 14h2l1.5 2.5-1 3h-3zm8-6H15l-2-2.5 1-3 2.5-1.5 2 2.5zm-2 6l-1-3 2-2.5 2.5 1.5v3z') },
  { name: 'Cricket',       defaultCourts: 3, icon: icon('M4.5 19.5l15-15M9 15l-4.5 4.5M15 9l4.5-4.5M12 12a2 2 0 1 0 0-4 2 2 0 0 0 0 4z') },
  { name: 'Hockey',        defaultCourts: 1, icon: icon('M6 3c0 0 0 8 0 12s4 6 6 6 6-2 6-6 0-12 0-12M6 15h12') },
  { name: 'Basketball',    defaultCourts: 2, icon: icon('M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20zM2 12h20M12 2a15.3 15.3 0 0 1 0 20M12 2a15.3 15.3 0 0 0 0 20') },
  { name: 'Volleyball',    defaultCourts: 1, icon: icon('M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20zM2 12h20M12 2v20') },
  { name: 'Tennis',        defaultCourts: 2, icon: icon('M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20zM12 2v20M4 6c2.5 2 5 3 8 3s5.5-1 8-3M4 18c2.5-2 5-3 8-3s5.5 1 8 3') },
  { name: 'Athletics',     defaultCourts: 1, icon: icon('M13 4a1.5 1.5 0 1 0 0-3 1.5 1.5 0 0 0 0 3zM7 21l3-7 2.5 2V21h2v-6l-2.5-2 .5-3c1.5 1.5 3.5 2.5 6 2.5v-2c-2 0-3.5-1-4.5-2.5l-1-1.5c-.5-.5-1-1-2-1s-1 0-1.5.5L5 11v4h2V12l2-1.5') },
  { name: 'Swimming',      defaultCourts: 4, icon: icon('M2 16c1-1 2.5-2 4-2s3 1 4 2 2.5 2 4 2 3-1 4-2 2.5-2 4-2M2 20c1-1 2.5-2 4-2s3 1 4 2 2.5 2 4 2 3-1 4-2 2.5-2 4-2M2 12c1-1 2.5-2 4-2s3 1 4 2 2.5 2 4 2 3-1 4-2 2.5-2 4-2') },
  { name: 'Water Polo',    defaultCourts: 1, icon: icon('M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20zM2 16c1-1 2.5-2 4-2s3 1 4 2 2.5 2 4 2 3-1 4-2 2.5-2 4-2M8 8h.01M16 8h.01') },
  { name: 'Kho-Kho',       defaultCourts: 1, icon: icon('M5 12h14M12 5v14M9 3v4M15 3v4M9 17v4M15 17v4') },
  { name: 'Badminton',     defaultCourts: 2, icon: icon('M12.5 2.5l-1 1c-2 2-3 5-1 7l4 4c2 2 5 0 7-1l1-1M6 18l-4 4M14 10l-8 8') },
  { name: 'Table Tennis',  defaultCourts: 6, icon: icon('M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20zM12 12a3 3 0 1 0 0-6 3 3 0 0 0 0 6zM2 12h6') },
  { name: 'Squash',        defaultCourts: 2, icon: icon('M4 4h16v16H4zM4 4l16 16M20 4L4 20M12 8a2 2 0 1 0 0 4 2 2 0 0 0 0-4z') },
];

export default function BookCourtPage() {
  const navigate = useNavigate();
  const [facilities, setFacilities] = useState<Facility[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function loadFacilities() {
      try {
        const data = await facilityApi.listFacilities();
        setFacilities(data || []);
      } catch (err) {
        console.warn('Could not load facilities from backend:', err);
      } finally {
        setLoading(false);
      }
    }
    loadFacilities();
  }, []);

  const handleSportClick = (sport: SportMeta) => {
    // Find matching facility from backend
    const match = facilities.find(
      (f) => f.sport_type.toLowerCase() === sport.name.toLowerCase()
    );

    if (match) {
      navigate(`/courts/${match.id}?sport=${encodeURIComponent(sport.name)}&facility=${encodeURIComponent(match.name)}`);
    } else {
      // Navigate using the sport slug (will allow fallback courts or display)
      navigate(`/courts/${sport.name.toLowerCase().replace(/\s+/g, '-')}`);
    }
  };

  return (
    <div className="min-h-dvh bg-blush pb-24">
      <Header title="Book Court" showLogo showAvatar />

      <div className="px-5 pt-2 pb-4">
        <h2 className="text-2xl font-extrabold text-plum leading-tight">Select a Sport</h2>
        <p className="text-sm text-plum-muted mt-1">Choose a court and book your session instantly.</p>
      </div>

      {loading ? (
        <div className="px-5 py-12 flex flex-col items-center justify-center gap-3">
          <div className="w-8 h-8 border-3 border-plum/20 border-t-plum rounded-full animate-spin" />
          <p className="text-xs text-plum-muted">Loading campus sports facilities...</p>
        </div>
      ) : (
        <div className="px-4 grid grid-cols-3 gap-3">
          {sportsList.map((sport) => {
            const match = facilities.find(
              (f) => f.sport_type.toLowerCase() === sport.name.toLowerCase()
            );

            return (
              <SportCard
                key={sport.name}
                name={sport.name}
                courtsAvailable={match ? sport.defaultCourts : sport.defaultCourts}
                icon={sport.icon}
                onClick={() => handleSportClick(sport)}
              />
            );
          })}
        </div>
      )}

      <BottomNav />
    </div>
  );
}
