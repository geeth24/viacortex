'use client';

import { useRouter } from 'next/navigation';
import { useState, useEffect } from 'react';

type User = {
  id: string;
  name?: string;
  email: string;
  role: string;
  active?: boolean;
  lastLogin?: string;
};

export function useAuth() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchUser();
  }, []);

  const fetchUser = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const response = await fetch('/api/user');
      
      if (!response.ok) {
        if (response.status === 401) {
          router.push('/login');
          return;
        }
        
        throw new Error('Failed to fetch user data');
      }
      
      const data = await response.json();
      
      if (data.user) {
        setUser(data.user);
      } else {
        setUser(data);
      }
    } catch (err) {
      console.error('Error fetching user:', err);
      setError('Failed to fetch user data');
    } finally {
      setLoading(false);
    }
  };

  const updateUser = async (userData: Partial<User>) => {
    try {
      setLoading(true);
      setError(null);
      
      const response = await fetch('/api/user', {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(userData),
      });
      
      if (!response.ok) {
        if (response.status === 401) {
          router.push('/login');
          return false;
        }
        
        throw new Error('Failed to update user data');
      }
      
      const data = await response.json();
      
      if (data.user) {
        setUser(data.user);
      } else if (data.success) {
        await fetchUser(); // Refresh user data
      }
      
      return true;
    } catch (err) {
      console.error('Error updating user:', err);
      setError('Failed to update user data');
      return false;
    } finally {
      setLoading(false);
    }
  };

  const logout = async () => {
    try {
      // This assumes you have a logout endpoint that clears the session cookie
      await fetch('/api/auth/logout', { method: 'POST' });
      router.push('/login');
    } catch (error) {
      console.error('Error logging out:', error);
    }
  };

  return {
    user,
    loading,
    error,
    fetchUser,
    updateUser,
    logout,
  };
} 