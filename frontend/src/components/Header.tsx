import { useNavigate } from 'react-router-dom';

interface HeaderProps {
  title: string;
  showBack?: boolean;
  showAvatar?: boolean;
  showLogo?: boolean;
}

export default function Header({ title, showBack = false, showAvatar = true, showLogo = false }: HeaderProps) {
  const navigate = useNavigate();

  return (
    <header className="flex items-center justify-between px-5 pt-[max(1rem,env(safe-area-inset-top))] pb-3">
      <div className="flex items-center gap-3">
        {showBack && (
          <button
            onClick={() => navigate(-1)}
            className="w-9 h-9 flex items-center justify-center rounded-full hover:bg-blush-dark/40 transition-colors"
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#271F30" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <path d="M19 12H5" />
              <path d="M12 19l-7-7 7-7" />
            </svg>
          </button>
        )}
        {showLogo && (
          <div className="w-8 h-8 rounded-lg bg-white shadow-xs border border-blush-dark/30 flex items-center justify-center p-1 overflow-hidden">
            <img src="/icon.svg" alt="Huddle Up Logo" className="w-full h-full object-contain" />
          </div>
        )}
        <h1 className="text-lg font-bold text-plum">{title}</h1>
      </div>

      {showAvatar && (
        <div className="w-9 h-9 rounded-full bg-plum/20 overflow-hidden flex items-center justify-center ring-2 ring-plum/10">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#271F30" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
            <circle cx="12" cy="7" r="4" />
          </svg>
        </div>
      )}
    </header>
  );
}
