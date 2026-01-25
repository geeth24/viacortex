import { NextResponse } from 'next/server';

export async function GET() {
  const apiUrl = process.env.API_URL || 'http://localhost:8080';

  try {
    const response = await fetch(`${apiUrl}/api/status`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (response.ok) {
      const data = await response.json();
      return NextResponse.json({
        success: true,
        apiUrl,
        status: data
      });
    } else {
      return NextResponse.json({
        success: false,
        apiUrl,
        error: `Server returned ${response.status}`
      });
    }
  } catch (error) {
    return NextResponse.json({
      success: false,
      apiUrl,
      error: error instanceof Error ? error.message : 'Unknown error'
    });
  }
}
