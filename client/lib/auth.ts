import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';

export async function getAuthToken(): Promise<string> {
  const cookieStore = await cookies();
  const session = cookieStore.get('session');
  
  if (!session?.value) {
    redirect('/login');
  }
  
  return session.value;
}

export async function fetchWithAuth<T>(url: string, options?: RequestInit): Promise<T> {
  const token = await getAuthToken();
  
  // Check if the URL is using the correct API path format
  // Make sure we're calling the API with the right endpoints
  let apiUrl = url;
  if (url.includes('/user/profile')) {
    // The endpoint might actually be: /api/profile instead
    apiUrl = url.replace('/user/profile', '/api/profile');
  }
  
  console.log('Fetching with auth:', apiUrl);
  
  try {
    const response = await fetch(apiUrl, {
      ...options,
      headers: {
        ...options?.headers,
        Authorization: `Bearer ${token}`,
      },
    });
    
    if (!response.ok) {
      if (response.status === 401) {
        // Clear the invalid session
        const cookieStore = await cookies();
        cookieStore.delete('session');
        redirect('/login');
      }
      
      // Try to get more information from the response
      const errorText = await response.text();
      console.error('API request failed:', {
        status: response.status,
        url: apiUrl,
        error: errorText
      });
      
      throw new Error(`API request failed with status ${response.status}: ${errorText}`);
    }
    
    const data = await response.json();
    return data;
  } catch (error) {
    console.error('Fetch error:', error);
    throw error;
  }
} 