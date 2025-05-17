import { NextRequest, NextResponse } from 'next/server';

export async function POST(request: NextRequest) {
  try {
    console.log('Login API route called');
    
    // Get request body
    const body = await request.json();
    const { email, password } = body;
    console.log('Request body:', { email, password: password ? '******' : undefined });

    if (!email || !password) {
      console.log('Missing email or password');
      return NextResponse.json(
        { success: false, message: 'Email and password are required' },
        { status: 400 }
      );
    }

    // Get API URL with path correction
    const apiUrl = process.env.API_URL || 'http://localhost:8080';
    console.log('Base API URL:', apiUrl);
    
    // Use the correct endpoint according to the server's routes.go
    const loginUrl = `${apiUrl}/api/login`;
    console.log('Trying login endpoint:', loginUrl);
    
    const response = await fetch(loginUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ email, password }),
    });

    console.log('Backend response status:', response.status);
    
    // Handle non-successful responses
    if (!response.ok) {
      let errorMessage = 'Invalid credentials';
      try {
        const errorData = await response.text();
        console.log('Error response text:', errorData);
        if (errorData) {
          errorMessage = errorData;
        }
      } catch (error) {
        console.error('Error parsing error response:', error);
      }
      
      return NextResponse.json(
        { success: false, message: errorMessage },
        { status: response.status }
      );
    }
    
    // Parse the successful response
    try {
      const responseText = await response.text();
      console.log('Response text:', responseText);
      
      let responseData;
      try {
        responseData = JSON.parse(responseText);
      } catch (error) {
        console.error('Failed to parse JSON:', error);
        return NextResponse.json(
          { success: false, message: 'Invalid JSON response from server' },
          { status: 500 }
        );
      }
      
      console.log('Parsed response data:', responseData);
      
      // Check if we have the expected token
      if (!responseData?.access_token) {
        console.error('No access token in response:', responseData);
        return NextResponse.json(
          { success: false, message: 'Invalid server response: no access token' },
          { status: 500 }
        );
      }

      // Create the response with the cookie
      const clientResponse = NextResponse.json({ 
        success: true,
        user: responseData.user
      });
      
      console.log('Setting cookie with access token');
      
      // Set the cookie in the response
      clientResponse.cookies.set({
        name: 'session',
        value: responseData.access_token,
        httpOnly: true,
        secure: process.env.NODE_ENV === 'production',
        maxAge: 60 * 60 * 24 * 7, // 1 week
        path: '/',
        sameSite: 'lax',
      });

      return clientResponse;
    } catch (error) {
      console.error('Error handling response:', error);
      return NextResponse.json(
        { success: false, message: 'Invalid server response format' },
        { status: 500 }
      );
    }
  } catch (error) {
    console.error('Login error:', error);
    return NextResponse.json(
      { success: false, message: 'Authentication failed' },
      { status: 500 }
    );
  }
} 