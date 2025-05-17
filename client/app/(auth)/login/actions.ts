'use server';

import { cookies } from 'next/headers';

type LoginResult = {
  success: boolean;
  error?: string;
};

export async function login(formData: FormData): Promise<LoginResult> {
  const email = formData.get('email') as string;
  const password = formData.get('password') as string;

  if (!email || !password) {
    return {
      success: false,
      error: 'Email and password are required',
    };
  }

  try {
    // Make a request to the backend API for authentication
    const response = await fetch(`${process.env.API_URL}/api/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ email, password }),
    });

    // First check if the response is OK
    if (!response.ok) {
      // Try to get error details if available
      let errorMsg = 'Invalid credentials';
      try {
        const errorData = await response.json();
        errorMsg = errorData.message || errorMsg;
      } catch (e) {
        // Ignore JSON parsing errors
      }
      
      return {
        success: false,
        error: errorMsg,
      };
    }

    // Then try to parse the response data
    let data;
    try {
      data = await response.json();
    } catch (e) {
      console.error('JSON parsing error:', e);
      return {
        success: false,
        error: 'Invalid server response format',
      };
    }

    // Check if we have the expected token
    if (!data.access_token) {
      return {
        success: false,
        error: 'Invalid server response - no access token',
      };
    }

    // Set session cookie with the access token
    const cookieStore = await cookies();
    cookieStore.set('session', data.access_token, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      maxAge: 60 * 60 * 24 * 7, // 1 week
      path: '/',
      sameSite: 'lax',
    });

    return { success: true };
  } catch (error) {
    console.error('Login error:', error);
    return {
      success: false,
      error: 'Authentication failed',
    };
  }
} 