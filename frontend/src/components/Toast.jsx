import React, { useEffect } from 'react';
import { AlertCircle, CheckCircle2, X } from 'lucide-react';

export default function Toast({ type = 'error', message, onClose, duration = 4000 }) {
  useEffect(() => {
    if (!message) return;
    const timer = setTimeout(() => {
      onClose();
    }, duration);
    return () => clearTimeout(timer);
  }, [message, duration, onClose]);

  if (!message) return null;

  return (
    <div className="fixed bottom-6 right-6 z-50 flex items-center gap-3 px-4 py-3 rounded-xl border shadow-xl bg-[#14141E] animate-bounce-short ${
      type === 'error' ? 'border-rose-800/60 text-rose-300' : 'border-emerald-800/60 text-emerald-300'
    }">
      {type === 'error' ? (
        <AlertCircle className="w-5 h-5 text-rose-400 shrink-0" />
      ) : (
        <CheckCircle2 className="w-5 h-5 text-emerald-400 shrink-0" />
      )}
      <span className="text-xs font-medium">{message}</span>
      <button onClick={onClose} className="p-1 hover:text-white shrink-0">
        <X className="w-4 h-4" />
      </button>
    </div>
  );
}
