import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from './AuthContext';
import type { NotificationPayload } from '../api/types';

interface NotificationItem extends NotificationPayload {
  id: string;
  receivedAt: Date;
}

interface NotificationContextType {
  notifications: NotificationItem[];
  dismissNotification: (id: string) => void;
  clearAll: () => void;
}

const NotificationContext = createContext<NotificationContextType | undefined>(undefined);

export function NotificationProvider({ children }: { children: ReactNode }) {
  const { user, token } = useAuth();
  const navigate = useNavigate();
  const [notifications, setNotifications] = useState<NotificationItem[]>([]);

  useEffect(() => {
    if (!user || !token) {
      setNotifications([]);
      return;
    }

    let es: EventSource | null = null;
    let reconnectTimeout: number | undefined;

    function connect() {
      const url = `/api/notifications/stream?token=${encodeURIComponent(token!)}`;
      es = new EventSource(url);

      es.onmessage = (event) => {
        try {
          const payload: NotificationPayload = JSON.parse(event.data);
          const item: NotificationItem = {
            ...payload,
            id: `${payload.slot_id}-${Date.now()}`,
            receivedAt: new Date(),
          };

          setNotifications((prev) => [item, ...prev]);

          // Auto-dismiss after 15 seconds
          setTimeout(() => {
            setNotifications((prev) => prev.filter((n) => n.id !== item.id));
          }, 15000);
        } catch (err) {
          console.error('Failed to parse SSE notification:', err);
        }
      };

      es.onerror = () => {
        if (es) {
          es.close();
        }
        // Attempt reconnect after 5s if still authenticated
        reconnectTimeout = window.setTimeout(() => {
          if (token) connect();
        }, 5000);
      };
    }

    connect();

    return () => {
      if (es) es.close();
      if (reconnectTimeout) clearTimeout(reconnectTimeout);
    };
  }, [user, token]);

  const dismissNotification = (id: string) => {
    setNotifications((prev) => prev.filter((n) => n.id !== id));
  };

  const clearAll = () => {
    setNotifications([]);
  };

  return (
    <NotificationContext.Provider
      value={{
        notifications,
        dismissNotification,
        clearAll,
      }}
    >
      {children}

      {/* Floating In-App Live Alert Banners */}
      <div className="fixed top-4 left-1/2 -translate-x-1/2 w-full max-w-[400px] px-4 z-50 flex flex-col gap-2.5 pointer-events-none">
        {notifications.map((n) => (
          <div
            key={n.id}
            className="pointer-events-auto bg-plum text-white rounded-2xl p-4 shadow-2xl border border-white/10 animate-[scaleIn_0.2s_ease-out] flex flex-col gap-2"
          >
            <div className="flex items-start justify-between">
              <div className="flex items-center gap-2">
                <span className="w-6 h-6 rounded-full bg-warning/20 text-warning flex items-center justify-center text-xs">
                  🔔
                </span>
                <span className="text-xs font-bold uppercase tracking-wider text-warning">
                  Waitlist Alert!
                </span>
              </div>
              <button
                onClick={() => dismissNotification(n.id)}
                className="text-white/40 hover:text-white transition-colors text-xs"
              >
                ✕
              </button>
            </div>

            <p className="text-sm font-medium text-white/90">
              {n.message || `A slot at ${n.facility_name} (${n.court_label}) is now open!`}
            </p>

            <div className="flex items-center gap-2 mt-1">
              <button
                onClick={() => {
                  dismissNotification(n.id);
                  navigate('/sports');
                }}
                className="flex-1 bg-white text-plum py-2 px-3 rounded-xl text-xs font-bold hover:bg-blush transition-colors text-center"
              >
                Book Now →
              </button>
              <button
                onClick={() => dismissNotification(n.id)}
                className="bg-white/10 text-white/80 py-2 px-3 rounded-xl text-xs font-medium hover:bg-white/20 transition-colors"
              >
                Dismiss
              </button>
            </div>
          </div>
        ))}
      </div>
    </NotificationContext.Provider>
  );
}

export function useNotifications() {
  const context = useContext(NotificationContext);
  if (!context) {
    throw new Error('useNotifications must be used within a NotificationProvider');
  }
  return context;
}
