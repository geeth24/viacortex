import { 
  Card, 
  CardContent, 
  CardDescription, 
  CardHeader, 
  CardTitle 
} from '@/components/ui/card';
import { fetchWithAuth } from '@/lib/auth';
import { ProfileForm } from './profile-form';

async function getUserProfile() {
  try {
    // Use the correct API endpoint based on the server routes.go
    const userData = await fetchWithAuth<{
      user: {
        id: string;
        name: string;
        email: string;
        role: string;
      }
    }>(`${process.env.API_URL}/api/verify`, {
      next: { revalidate: 0 }, // Don't cache profile data
    });
    
    // Return in the format expected by the ProfileForm
    return {
      id: userData.user.id,
      name: userData.user.name,
      email: userData.user.email,
    };
  } catch (error) {
    console.error('Error fetching user profile:', error);
    throw error;
  }
}

export default async function ProfilePage() {
  const userProfile = await getUserProfile();
  
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Profile</h1>
        <p className="text-muted-foreground">
          Manage your account settings and profile information.
        </p>
      </div>
      <Card>
        <CardHeader>
          <CardTitle>Personal Information</CardTitle>
          <CardDescription>
            Update your personal information.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <ProfileForm userProfile={userProfile} />
        </CardContent>
      </Card>
    </div>
  );
} 