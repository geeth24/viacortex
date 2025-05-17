'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';

export default function LoginPage() {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [debugInfo, setDebugInfo] = useState<string | null>(null);
  const [connectionStatus, setConnectionStatus] = useState<string | null>(null);

  // Check server connection on load
  useEffect(() => {
    checkConnection();
  }, []);

  async function checkConnection() {
    try {
      setConnectionStatus('Checking server connection...');
      const response = await fetch('/api/test');
      const data = await response.json();
      
      if (data.success) {
        setConnectionStatus(`Connected to server at ${data.apiUrl}`);
      } else {
        setConnectionStatus(`Failed to connect to server: ${data.error}`);
      }
    } catch (err) {
      console.error('Connection test error:', err);
      setConnectionStatus('Failed to check server connection');
    }
  }

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsLoading(true);
    setError(null);
    setDebugInfo(null);
    
    const formData = new FormData(event.currentTarget);
    const email = formData.get('email') as string;
    const password = formData.get('password') as string;
    
    try {
      // Get the full URL to the login API endpoint
      const apiUrl = `${process.env.NEXT_PUBLIC_API_URL}/api/auth/login`;
      console.log('Making login request to:', apiUrl);
      
      const response = await fetch(apiUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, password }),
      });
      
      let data;
      try {
        data = await response.json();
      } catch (e) {
        console.error('Failed to parse JSON response:', e);
        const text = await response.text();
        setDebugInfo(`Status: ${response.status}, Response: ${text}`);
        throw new Error('Invalid response from server');
      }
      
      if (data.success) {
        router.push('/dashboard');
        router.refresh();
      } else {
        setError(data.message || 'Invalid credentials');
        if (data.debug) {
          setDebugInfo(data.debug);
        }
      }
    } catch (err) {
      console.error('Login error:', err);
      setError('An unexpected error occurred. Please try again later.');
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>Login</CardTitle>
        <CardDescription>Enter your credentials to access your account</CardDescription>
      </CardHeader>
      <CardContent>
        {connectionStatus && (
          <div className={`p-3 mb-4 rounded-md text-sm ${
            connectionStatus.includes('Failed') 
              ? 'bg-red-50 text-red-700' 
              : 'bg-green-50 text-green-700'
          }`}>
            {connectionStatus}
          </div>
        )}
        
        <form onSubmit={handleSubmit}>
          <div className="grid gap-4">
            <div className="grid gap-2">
              <Input
                id="email"
                name="email"
                placeholder="name@example.com"
                autoCapitalize="none"
                autoComplete="email"
                autoCorrect="off"
                disabled={isLoading}
                required
              />
            </div>
            <div className="grid gap-2">
              <Input
                id="password"
                name="password"
                type="password"
                autoCapitalize="none"
                autoComplete="current-password"
                disabled={isLoading}
                required
              />
            </div>
            {error && <p className="text-sm text-red-500">{error}</p>}
            {debugInfo && (
              <div className="text-xs p-2 bg-gray-100 rounded-md overflow-auto max-h-[100px]">
                <pre>{debugInfo}</pre>
              </div>
            )}
            <Button type="submit" disabled={isLoading}>
              {isLoading ? 'Logging in...' : 'Login'}
            </Button>
          </div>
        </form>
      </CardContent>
      <CardFooter>
        <div className="text-sm text-center w-full">
          Don&apos;t have an account? <a href="/register" className="underline">Register</a>
          {' | '}
          <a href="/test" className="underline">Connection Test</a>
        </div>
      </CardFooter>
    </Card>
  );
} 