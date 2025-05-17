import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

export async function GET(request: NextRequest) {
  try {
    const searchParams = request.nextUrl.searchParams
    const timeRange = searchParams.get('timeRange') || '24h'
    const domainId = searchParams.get('domainId')
    
    const authHeader = request.headers.get('authorization')
    if (!authHeader) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    // Construct the API URL
    let url = `${process.env.BACKEND_URL}/api/metrics?timeRange=${timeRange}`
    if (domainId) {
      url = `${process.env.BACKEND_URL}/api/metrics/${domainId}?timeRange=${timeRange}`
    }

    // Make request to backend API
    const response = await fetch(url, {
      headers: {
        'Authorization': authHeader,
        'Content-Type': 'application/json',
      },
    })

    if (!response.ok) {
      const errorData = await response.json()
      return NextResponse.json(errorData, { status: response.status })
    }

    const data = await response.json()
    return NextResponse.json(data)
  } catch (error) {
    console.error('Error fetching metrics:', error)
    return NextResponse.json({ error: 'Internal Server Error' }, { status: 500 })
  }
} 