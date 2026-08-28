interface TimeSlotProps {
  time: string;
  type: 'standard' | 'peak';
  state: 'available' | 'selected' | 'locked';
  onClick: () => void;
}

export default function TimeSlot({ time, type, state, onClick }: TimeSlotProps) {
  const base = 'rounded-2xl px-3 py-3 text-xs sm:text-sm font-semibold tracking-tight transition-all duration-200 relative flex items-center justify-center text-center';

  const stateClasses = {
    available: type === 'peak'
      ? 'bg-[#E05A2B]/10 border-2 border-[#E05A2B] text-plum shadow-xs hover:bg-[#E05A2B]/20 active:scale-[0.97]'
      : 'bg-white border-2 border-transparent text-plum shadow-xs hover:bg-blush-dark/30 active:scale-[0.97]',
    selected: type === 'peak'
      ? 'bg-[#E05A2B]/15 border-2 border-[#E05A2B] text-plum ring-2 ring-plum shadow-md scale-[1.02]'
      : 'bg-white border-2 border-transparent text-plum ring-2 ring-plum shadow-md scale-[1.02]',
    locked: 'bg-blush-dark/30 border-2 border-transparent text-plum-muted/40 cursor-not-allowed',
  };

  return (
    <button
      onClick={onClick}
      disabled={state === 'locked'}
      className={`${base} ${stateClasses[state]}`}
    >
      <span>{time}</span>
      {state === 'selected' && (
        <span className="absolute top-1.5 right-1.5 w-4 h-4 rounded-full bg-plum flex items-center justify-center">
          <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
            <polyline points="20 6 9 17 4 12" />
          </svg>
        </span>
      )}
      {state === 'locked' && (
        <span className="absolute top-1.5 right-1.5">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="#9B8FA6" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
            <path d="M7 11V7a5 5 0 0 1 10 0v4" />
          </svg>
        </span>
      )}
    </button>
  );
}
