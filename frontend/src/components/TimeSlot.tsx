interface TimeSlotProps {
  time: string;
  type: 'standard' | 'peak';
  state: 'available' | 'selected' | 'locked';
  onClick: () => void;
}

export default function TimeSlot({ time, type, state, onClick }: TimeSlotProps) {
  const base =
    'rounded-2xl px-3 py-3 text-xs sm:text-sm font-semibold tracking-tight transition-all duration-200 relative flex items-center justify-center text-center cursor-pointer';

  const stateClasses = {
    available:
      type === 'peak'
        ? 'bg-[#E05A2B]/10 border-2 border-[#E05A2B] text-plum shadow-xs hover:bg-[#E05A2B]/20 active:scale-[0.97]'
        : 'bg-white border-2 border-transparent text-plum shadow-xs hover:bg-blush-dark/30 active:scale-[0.97]',
    selected:
      type === 'peak'
        ? 'bg-[#E05A2B]/15 border-2 border-[#E05A2B] text-plum ring-2 ring-plum shadow-md scale-[1.02]'
        : 'bg-white border-2 border-transparent text-plum ring-2 ring-plum shadow-md scale-[1.02]',
    locked:
      'bg-blush-dark/40 border-2 border-dashed border-plum/20 hover:border-[#E05A2B] hover:bg-[#E05A2B]/10 text-plum/70 shadow-xs active:scale-[0.97]',
  };

  return (
    <button
      type="button"
      onClick={onClick}
      className={`${base} ${stateClasses[state]}`}
    >
      <div className="flex flex-col items-center">
        <span>{time}</span>
        {state === 'locked' && (
          <span className="text-[10px] font-semibold text-[#E05A2B] tracking-tight mt-0.5">
            Booked • Waitlist
          </span>
        )}
      </div>

      {state === 'selected' && (
        <span className="absolute top-1.5 right-1.5 w-4 h-4 rounded-full bg-plum flex items-center justify-center">
          <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
            <polyline points="20 6 9 17 4 12" />
          </svg>
        </span>
      )}

      {state === 'locked' && (
        <span className="absolute top-1.5 right-1.5 text-xs">
          🔒
        </span>
      )}
    </button>
  );
}
