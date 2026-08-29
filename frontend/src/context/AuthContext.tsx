import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import type { User } from '../api/types';
import { authApi, getStoredToken, setStoredToken } from '../api/client';

interface AuthContextType {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  requestOtp: (email: string) => Promise<string | undefined>;
  verifyOtp: (email: string, code: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(getStoredToken());
  const [isLoading, setIsLoading] = useState<boolean>(true);

  useEffect(() => {
    async function loadUser() {
      const stored = getStoredToken();
      if (!stored) {
        setIsLoading(false);
        return;
      }

      try {
        const currentUser = await authApi.getMe();
        setUser(currentUser);
        setToken(stored);
      } catch (err) {
        console.warn('Session expired or invalid:', err);
        setStoredToken(null);
        setUser(null);
        setToken(null);
      } finally {
        setIsLoading(false);
      }
    }

    loadUser();

    const handleStorageChange = (e: StorageEvent) => {
      if (e.key === 'token') {
        if (!e.newValue) {
          setUser(null);
          setToken(null);
        } else {
          loadUser();
        }
      }
    };
    window.addEventListener('storage', handleStorageChange);
    return () => window.removeEventListener('storage', handleStorageChange);
  }, []);

  const requestOtp = async (email: string): Promise<string | undefined> => {
    const res = await authApi.requestOTP(email);
    return res.code;
  };

  const verifyOtp = async (email: string, code: string): Promise<void> => {
    const res = await authApi.verifyOTP(email, code);
    setStoredToken(res.token);
    setToken(res.token);
    setUser(res.user);
  };

  const logout = () => {
    setStoredToken(null);
    setToken(null);
    setUser(null);
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        isLoading,
        isAuthenticated: !!user && !!token,
        requestOtp,
        verifyOtp,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
