'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { updateProfile } from './actions';

type UserProfile = {
  id: string;
  name: string;
  email: string;
};

export function ProfileForm({ userProfile }: { userProfile: UserProfile }) {
  const [isLoading, setIsLoading] = useState(false);
  const [message, setMessage] = useState<{ text: string; type: 'success' | 'error' } | null>(null);

  async function handleSubmit(formData: FormData) {
    setIsLoading(true);
    setMessage(null);
    
    try {
      const result = await updateProfile(formData);
      
      if (result.success) {
        setMessage({ text: 'Profile updated successfully', type: 'success' });
      } else {
        setMessage({ text: result.error || 'Failed to update profile', type: 'error' });
      }
    } catch (err) {
      setMessage({ text: 'An unexpected error occurred', type: 'error' });
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <form action={handleSubmit} className="space-y-4">
      <input type="hidden" name="id" value={userProfile.id} />
      
      <div className="space-y-2">
        <Label htmlFor="name">Name</Label>
        <Input 
          id="name" 
          name="name" 
          defaultValue={userProfile.name} 
          disabled={isLoading} 
          required 
        />
      </div>
      
      <div className="space-y-2">
        <Label htmlFor="email">Email</Label>
        <Input 
          id="email" 
          name="email" 
          type="email" 
          defaultValue={userProfile.email} 
          disabled={isLoading || true} // Email cannot be changed
          required 
        />
        <p className="text-sm text-muted-foreground">
          Your email address cannot be changed.
        </p>
      </div>
      
      {message && (
        <div className={`p-3 rounded-md ${
          message.type === 'success' ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'
        }`}>
          {message.text}
        </div>
      )}
      
      <Button type="submit" disabled={isLoading}>
        {isLoading ? 'Saving...' : 'Save changes'}
      </Button>
    </form>
  );
} 