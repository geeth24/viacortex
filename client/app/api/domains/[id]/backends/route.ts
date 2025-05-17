import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { cookies } from 'next/headers'

export async function GET(
  request: NextRequest,
  context: { params: { id: string } }
) {
  try {
    // Get the session cookie
    const cookieStore = await cookies()
    const session = cookieStore.get('session')
    
    if (!session?.value) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    // Extract domain ID from params with proper await
    const { id: domainId } = await Promise.resolve(context.params)
    
    // Make request to backend API
    const response = await fetch(`${process.env.BACKEND_URL}/api/domains/${domainId}/backends`, {
      headers: {
        'Authorization': `Bearer ${session.value}`,
        'Content-Type': 'application/json',
      },
    })

    if (!response.ok) {
      const errorResponse = response.clone()
      try {
        const errorData = await response.json()
        return NextResponse.json(errorData, { status: response.status })
      } catch (error) {
        // If JSON parsing fails, try to get text
        const errorText = await errorResponse.text()
        return NextResponse.json({ error: errorText || 'Unknown error' }, { status: response.status })
      }
    }

    const data = await response.json()
    return NextResponse.json(data)
  } catch (error) {
    console.error('Error fetching backend servers:', error)
    return NextResponse.json({ error: 'Internal Server Error' }, { status: 500 })
  }
}

export async function POST(
  request: NextRequest,
  context: { params: { id: string } }
) {
  try {
    // Get the session cookie
    const cookieStore = await cookies()
    const session = cookieStore.get('session')
    
    if (!session?.value) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    // Extract domain ID from params with proper await
    const { id: domainId } = await Promise.resolve(context.params)
    
    const body = await request.json()

    // Make request to backend API
    const response = await fetch(`${process.env.BACKEND_URL}/api/domains/${domainId}/backends`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${session.value}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
    })

    if (!response.ok) {
      const errorResponse = response.clone()
      try {
        const errorData = await response.json()
        return NextResponse.json(errorData, { status: response.status })
      } catch (error) {
        // If JSON parsing fails, try to get text
        const errorText = await errorResponse.text()
        return NextResponse.json({ error: errorText || 'Unknown error' }, { status: response.status })
      }
    }

    const data = await response.json()
    return NextResponse.json(data)
  } catch (error) {
    console.error('Error creating backend server:', error)
    return NextResponse.json({ error: 'Internal Server Error' }, { status: 500 })
  }
} 