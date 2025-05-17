'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import { updateSettings } from './actions';

type UserSettings = {
  id: string;
  emailNotifications: boolean;
  marketingEmails: boolean;
  twoFactorAuth: boolean;
};

export function SettingsForm({ settings }: { settings: UserSettings }) {
  const [isLoading, setIsLoading] = useState(false);
  const [message, setMessage] = useState<{ text: string; type: 'success' | 'error' } | null>(null);
  const [formState, setFormState] = useState({
    emailNotifications: settings.emailNotifications,
    marketingEmails: settings.marketingEmails,
    twoFactorAuth: settings.twoFactorAuth,
  });

  async function handleSubmit(formData: FormData) {
    setIsLoading(true);
    setMessage(null);
    
    try {
      const result = await updateSettings(formData);
      
      if (result.success) {
        setMessage({ text: 'Settings updated successfully', type: 'success' });
      } else {
        setMessage({ text: result.error || 'Failed to update settings', type: 'error' });
      }
    } catch (err) {
      setMessage({ text: 'An unexpected error occurred', type: 'error' });
    } finally {
      setIsLoading(false);
    }
  }

  function handleSwitchChange(name: string, checked: boolean) {
    setFormState(prev => ({ ...prev, [name]: checked }));
  }

  return (
    <form action={handleSubmit} className="space-y-4">
      <input type="hidden" name="id" value={settings.id} />
      
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div className="space-y-0.5">
            <Label htmlFor="emailNotifications">Email Notifications</Label>
            <p className="text-sm text-muted-foreground">
              Receive email notifications for important updates.
            </p>
          </div>
          <Switch 
            id="emailNotifications" 
            name="emailNotifications"
            checked={formState.emailNotifications}
            onCheckedChange={(checked) => handleSwitchChange('emailNotifications', checked)}
            disabled={isLoading}
          />
        </div>
        
        <div className="flex items-center justify-between">
          <div className="space-y-0.5">
            <Label htmlFor="marketingEmails">Marketing Emails</Label>
            <p className="text-sm text-muted-foreground">
              Receive emails about new features and promotions.
            </p>
          </div>
          <Switch 
            id="marketingEmails" 
            name="marketingEmails"
            checked={formState.marketingEmails}
            onCheckedChange={(checked) => handleSwitchChange('marketingEmails', checked)}
            disabled={isLoading}
          />
        </div>
        
        <div className="flex items-center justify-between">
          <div className="space-y-0.5">
            <Label htmlFor="twoFactorAuth">Two-Factor Authentication</Label>
            <p className="text-sm text-muted-foreground">
              Enable two-factor authentication for enhanced security.
            </p>
          </div>
          <Switch 
            id="twoFactorAuth" 
            name="twoFactorAuth"
            checked={formState.twoFactorAuth}
            onCheckedChange={(checked) => handleSwitchChange('twoFactorAuth', checked)}
            disabled={isLoading}
          />
        </div>
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