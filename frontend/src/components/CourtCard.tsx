import StatusBadge from './StatusBadge';

interface CourtCardProps {
  name: string;
  surface: string;
  type: string;
  status: 'available' | 'maintenance';
  nextAvailable?: string;
  selected?: boolean;
  onClick: () => void;
  maintenanceNote?: string;
}

export default function CourtCard({
  name,
  surface,
  type,
  status,
  nextAvailable,
  selected = false,
  onClick,
  maintenanceNote,
}: CourtCardProps) {
  return (
    <button
      onClick={onClick}
      disabled={status === 'maintenance'}
      className={`w-full rounded-2xl p-5 text-left transition-all duration-300 relative ${
        selected
          ? 'bg-plum text-white shadow-lg shadow-plum/20 scale-[1.02]'
          : status === 'maintenance'
            ? 'bg-blush-dark/50 opacity-70 cursor-not-allowed'
            : 'bg-white hover:shadow-md hover:-translate-y-0.5 active:scale-[0.98]'
      }`}
    >
      <div className="flex items-start justify-between mb-2">
        <div>
          <h3 className={`text-xl font-bold ${selected ? 'text-white' : 'text-plum'}`}>{name}</h3>
          <p className={`text-sm mt-0.5 ${selected ? 'text-white/70' : 'text-plum-muted'}`}>
            {type} • {surface}
          </p>
        </div>
        <StatusBadge status={status} />
      </div>

      <div className={`flex items-center gap-2 mt-3 text-sm ${selected ? 'text-white/70' : 'text-plum-muted'}`}>
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <circle cx="12" cy="12" r="10" />
          <polyline points="12 6 12 12 16 14" />
        </svg>
        {status === 'maintenance' && maintenanceNote
          ? maintenanceNote
          : `Next available: ${nextAvailable}`
        }
      </div>

      {selected && (
        <div className="absolute bottom-4 right-4 w-8 h-8 rounded-full bg-white/20 flex items-center justify-center">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
            <polyline points="20 6 9 17 4 12" />
          </svg>
        </div>
      )}
    </button>
  );
}
