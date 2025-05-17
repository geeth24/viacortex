'use server';

import { cookies } from 'next/headers';
import { revalidatePath } from 'next/cache';

type UpdateProfileResult = {
  success: boolean;
  error?: string;
};

export async function updateProfile(formData: FormData): Promise<UpdateProfileResult> {
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
    const name = formData.get('name') as string;

    if (!id || !name) {
      return {
        success: false,
        error: 'Missing required fields',
      };
    }

    // The correct endpoint according to routes.go is "/api/profile"
    const apiUrl = process.env.API_URL || 'http://localhost:8080';
    const response = await fetch(`${apiUrl}/api/profile`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${session.value}`,
      },
      body: JSON.stringify({ id, name }),
    });

    // Handle the response
    if (!response.ok) {
      // Try to get error details
      let errorMsg = 'Failed to update profile';
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

    // Revalidate the profile page to reflect updated data
    revalidatePath('/profile');
    revalidatePath('/dashboard');

    return { success: true };
  } catch (error) {
    console.error('Profile update error:', error);
    return {
      success: false,
      error: 'Failed to update profile',
    };
  }
} 