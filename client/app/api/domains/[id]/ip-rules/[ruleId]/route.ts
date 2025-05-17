import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

export async function DELETE(
  request: NextRequest,
  { params }: { params: { id: string, ruleId: string } }
) {
  try {
    const domainId = params.id
    const ruleId = params.ruleId
    const authHeader = request.headers.get('authorization')
    
    if (!authHeader) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    // Make request to backend API
    const response = await fetch(`${process.env.BACKEND_URL}/api/domains/${domainId}/ip-rules/${ruleId}`, {
      method: 'DELETE',
      headers: {
        'Authorization': authHeader,
        'Content-Type': 'application/json',
      },
    })

    if (!response.ok) {
      const errorData = await response.json()
      return NextResponse.json(errorData, { status: response.status })
    }

    return NextResponse.json({ success: true })
  } catch (error) {
    console.error('Error deleting IP rule:', error)
    return NextResponse.json({ error: 'Internal Server Error' }, { status: 500 })
  }
} 