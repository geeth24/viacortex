import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { fetchWithAuth } from '@/lib/auth';
import { Separator } from '@/components/ui/separator';
import { SettingsForm } from './settings-form';

async function getUserSettings() {
  try {
    // Use the correct API endpoint
    // Since we don't have a specific settings endpoint in the server routes,
    // we'll use the verify endpoint and create default settings
    const userData = await fetchWithAuth<{
      user: {
        id: string;
        email: string;
        name: string;
        role: string;
      }
    }>(`${process.env.API_URL}/api/verify`, {
      next: { revalidate: 0 },
    });
    
    // Return settings format expected by the form
    // Using default values since the server doesn't have a settings endpoint
    return {
      id: userData.user.id,
      emailNotifications: true,
      marketingEmails: false,
      twoFactorAuth: false,
    };
  } catch (error) {
    console.error('Error fetching user settings:', error);
    throw error;
  }
}

export default async function SettingsPage() {
  const settings = await getUserSettings();
  
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Settings</h1>
        <p className="text-muted-foreground">
          Manage your account settings and preferences.
        </p>
      </div>
      <Separator />
      <div className="grid gap-4">
        <Card>
          <CardHeader>
            <CardTitle>Notifications</CardTitle>
            <CardDescription>
              Configure how you receive notifications.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <SettingsForm settings={settings} />
          </CardContent>
        </Card>
      </div>
    </div>
  );
} 