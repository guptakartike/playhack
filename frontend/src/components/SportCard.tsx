interface SportCardProps {
  name: string;
  courtsAvailable: number;
  icon: React.ReactNode;
  onClick: () => void;
}

export default function SportCard({ name, courtsAvailable, icon, onClick }: SportCardProps) {
  return (
    <button
      onClick={onClick}
      className="group w-full bg-white rounded-2xl py-4 px-2 flex flex-col items-center gap-2 shadow-sm hover:shadow-md transition-all duration-300 hover:-translate-y-0.5 active:scale-[0.97]"
    >
      <div className="w-12 h-12 rounded-full bg-blush flex items-center justify-center group-hover:bg-blush-dark transition-colors duration-300">
        {icon}
      </div>
      <span className="text-xs font-semibold text-plum text-center leading-tight">{name}</span>
      <span className="text-[8px] font-semibold tracking-wider uppercase text-plum-muted">
        {courtsAvailable} {courtsAvailable === 1 ? 'court' : 'courts'}
      </span>
    </button>
  );
}
