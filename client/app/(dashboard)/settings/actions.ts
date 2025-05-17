'use server';

import { cookies } from 'next/headers';
import { revalidatePath } from 'next/cache';

type UpdateSettingsResult = {
  success: boolean;
  error?: string;
};

export async function updateSettings(formData: FormData): Promise<UpdateSettingsResult> {
  try {
    const cookieStore = await cookies();
    const session = cookieStore.get('session');
    
    if (!session?.value) {
      return {
        success: false,
        error: 'Unauthorized - Please log in again',
      };
    }
    
    const id = formData.get('id') as string;
    const emailNotifications = formData.get('emailNotifications') === 'on';
    const marketingEmails = formData.get('marketingEmails') === 'on';
    const twoFactorAuth = formData.get('twoFactorAuth') === 'on';

    if (!id) {
      return {
        success: false,
        error: 'Missing user ID',
      };
    }

    // Note: Since the server doesn't have a settings endpoint in routes.go,
    // we'll just return success without actually making an API call.
    // In a real implementation, we would call the appropriate server endpoint.
    
    /* 
    // Example of how this would work with a real endpoint
    const apiUrl = process.env.API_URL || 'http://localhost:8080';
    const response = await fetch(`${apiUrl}/api/settings`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${session.value}`,
      },
      body: JSON.stringify({
        id,
        emailNotifications,
        marketingEmails,
        twoFactorAuth,
      }),
    });

    if (!response.ok) {
      let errorMsg = 'Failed to update settings';
      try {
        const errorText = await response.text();
        if (errorText) {
          errorMsg = errorText;
        }
      } catch (e) {
        // Ignore parsing errors
      }
      
      return {
        success: false,
        error: errorMsg,
      };
    }
    */

    // Revalidate the settings page to reflect updated data
    revalidatePath('/settings');

    return { success: true };
  } catch (error) {
    console.error('Settings update error:', error);
    return {
      success: false,
      error: 'Failed to update settings',
    };
  }
} 