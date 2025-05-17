import { NextRequest, NextResponse } from 'next/server';

export async function POST(request: NextRequest) {
  try {
    console.log('Register API route called');
    
    // Get request body
    const body = await request.json();
    const { name, email, password } = body;
    console.log('Request body:', { name, email, password: password ? '******' : undefined });

    if (!name || !email || !password) {
      console.log('Missing required fields');
      return NextResponse.json(
        { success: false, message: 'All fields are required' },
        { status: 400 }
      );
    }

    // Get API URL with path correction
    const apiUrl = process.env.API_URL || 'http://localhost:8080';
    console.log('Base API URL:', apiUrl);
    
    // The correct endpoint according to routes.go is "/api/register"
    const registerUrl = `${apiUrl}/api/register`;
    console.log('Trying register endpoint:', registerUrl);

    // Forward the request to the backend
    const response = await fetch(registerUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ 
        email, 
        password, 
        role: 'user',
        name // Including name if the server uses it
      }),
    });

    console.log('Backend response status:', response.status);

    // Handle non-successful responses
    if (!response.ok) {
      let errorMessage = 'Registration failed';
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
      
      // Return success response
      return NextResponse.json({ 
        success: true,
        message: 'Account created successfully'
      });
    } catch (error) {
      console.error('Error parsing JSON response:', error);
      return NextResponse.json(
        { success: false, message: 'Invalid server response format' },
        { status: 500 }
      );
    }
  } catch (error) {
    console.error('Registration error:', error);
    return NextResponse.json(
      { success: false, message: 'Registration failed' },
      { status: 500 }
    );
  }
} 