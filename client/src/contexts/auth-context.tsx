'use client';

import {
  createContext,
  useContext,
  useState,
  useEffect,
  ReactNode,
  useCallback,
  useMemo,
} from 'react';
import { useRouter } from 'next/navigation';
import { getCookie, setCookie, deleteCookie } from 'cookies-next/client';

interface User {
  id: string;
  email: string;
  name: string;
  last_login: string;
  active: boolean;
  role: string;
}

interface AuthContextType {
  user: User | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  isLoading: boolean;
  register: (email: string, password: string) => Promise<void>;
  updateProfile: (name: string) => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const router = useRouter();

  const checkAuth = useCallback(async () => {
    try {
      const token = getCookie('token');
      if (!token) {
        setIsLoading(false);
        return;
      }

      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/verify`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setUser(data.user);
      } else {
        setUser(null);
        deleteCookie('token');
        router.push('/login');
      }
    } catch (error) {
      console.error('Auth check failed:', error);
      setUser(null);
      deleteCookie('token');
      router.push('/login');
    } finally {
      setIsLoading(false);
    }
  }, [router]);

  useEffect(() => {
    checkAuth();
  }, [checkAuth]);

  const login = useCallback(
    async (email: string, password: string) => {
      try {
        const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/login`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ email, password }),
        });

        if (!response.ok) {
          const error = await response.text();
          throw new Error(error || 'Login failed');
        }

        const data = await response.json();
        setCookie('token', data.access_token, {
          maxAge: 60 * 60 * 24, // 1 day
          path: '/',
          secure: process.env.NODE_ENV === 'production',
          sameSite: 'strict',
        });
        setUser(data.user);
        router.refresh();
        router.push('/dashboard');
      } catch (error) {
        console.error('Login failed:', error);
        throw error;
      }
    },
    [router],
  );

  const register = useCallback(
    async (email: string, password: string) => {
      try {
        const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/register`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ email, password }),
        });

        if (!response.ok) {
          const error = await response.text();
          throw new Error(error || 'Registration failed');
        }

        const data = await response.json();
        setCookie('token', data.access_token, {
          maxAge: 60 * 60 * 24, // 1 day
          path: '/',
          secure: process.env.NODE_ENV === 'production',
          sameSite: 'strict',
        });
        setUser(data.user);
        router.refresh();
      } catch (error) {
        console.error('Registration failed:', error);
        throw error;
      }
    },
    [router],
  );

  const logout = useCallback(() => {
    deleteCookie('token');
    setUser(null);
    router.refresh();
    router.push('/login');
  }, [router]);

  const updateProfile = useCallback(async (name: string) => {
    try {
      const token = getCookie('token');
      if (!token) {
        throw new Error('Not authenticated');
      }

      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/profile`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ name }),
      });

      if (!response.ok) {
        throw response;
      }

      const data = await response.json();

      if (data.user) {
        setUser(data.user);
      }

      return data;
    } catch (error) {
      console.error('Profile update failed:', error);
      throw error;
    }
  }, []);

  const contextValue = useMemo(
    () => ({
      user,
      login,
      logout,
      isLoading,
      register,
      updateProfile,
    }),
    [user, login, logout, isLoading, register, updateProfile],
  );

  const MemoizedChildren = useMemo(() => children, [children]);

  return <AuthContext.Provider value={contextValue}>{MemoizedChildren}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
