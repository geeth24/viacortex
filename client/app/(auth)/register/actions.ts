'use server';

type RegisterResult = {
  success: boolean;
  error?: string;
};

export async function register(formData: FormData): Promise<RegisterResult> {
  const name = formData.get('name') as string;
  const email = formData.get('email') as string;
  const password = formData.get('password') as string;

  if (!name || !email || !password) {
    return {
      success: false,
      error: 'All fields are required',
    };
  }

  try {
    // Make a request to the backend API for registration
    const response = await fetch(`${process.env.API_URL}/api/auth/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ name, email, password, role: 'user' }),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      return {
        success: false,
        error: errorData.message || 'Registration failed',
      };
    }

    return { success: true };
  } catch (error) {
    console.error('Registration error:', error);
    return {
      success: false,
      error: 'Registration failed',
    };
  }
} 