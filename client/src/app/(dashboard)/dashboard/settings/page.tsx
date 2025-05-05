'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/contexts/auth-context';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { AlertCircle, CheckCircle } from 'lucide-react';

export default function SettingsPage() {
  const { user, updateProfile } = useAuth();
  const [name, setName] = useState(user?.name || '');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  // Update name state when user data changes
  useEffect(() => {
    if (user?.name) {
      setName(user.name);
    }
  }, [user?.name]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError('');
    setSuccess('');

    try {
      if (!name.trim()) {
        throw new Error('Name cannot be empty');
      }

      await updateProfile(name);
      setSuccess('Profile updated successfully');
    } catch (err: unknown) {
      console.error('Profile update error:', err);
      if (err instanceof Response) {
        // Handle HTTP error responses
        const text = await err.text();
        setError(text || 'Failed to update profile');
      } else {
        setError(err instanceof Error ? err.message : 'Failed to update profile');
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="container max-w-2xl">
      <Card>
        <CardHeader>
          <CardTitle>Profile Settings</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <label htmlFor="name" className="text-sm font-medium">
                Name
              </label>
              <Input
                id="name"
                value={name}
                onChange={(e) => {
                  setName(e.target.value);
                  // Clear messages when user starts typing
                  setError('');
                  setSuccess('');
                }}
                placeholder="Enter your name"
                disabled={isLoading}
                className={error ? 'border-red-500' : ''}
              />
              {error && (
                <Alert variant="destructive">
                  <AlertCircle className="h-4 w-4" />
                  <AlertTitle>Error</AlertTitle>
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              )}
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Email</label>
              <Input value={user?.email || ''} disabled />
              <p className="text-sm text-muted-foreground">Email cannot be changed</p>
            </div>
            {success && (
              <Alert>
                <CheckCircle className="h-4 w-4" />
                <AlertTitle>Profile updated successfully</AlertTitle>
                <AlertDescription>Your profile has been updated successfully.</AlertDescription>
              </Alert>
            )}
            <Button type="submit" disabled={isLoading || !name.trim()} className="w-full sm:w-auto">
              {isLoading ? (
                <span className="flex items-center gap-2">
                  <span className="animate-spin">⏳</span> Saving...
                </span>
              ) : (
                'Save Changes'
              )}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
