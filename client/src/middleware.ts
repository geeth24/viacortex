import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

export async function middleware(request: NextRequest) {
  // Get the pathname of the request
  const path = request.nextUrl.pathname

  // Public paths that don't require authentication
  const isPublicPath = path === '/login' || path === '/register'

  // Get token from cookie
  const token = request.cookies.get('token')?.value

  try {
    // Check if there are any users in the system
    const response = await fetch(`${process.env.API_URL}/api/check-users`, {
      headers: {
        'Cookie': request.headers.get('cookie') || ''
      }
    })
    const { count } = await response.json()

    // If no users exist and not on register page, redirect to register
    if (count === 0 && path !== '/register') {
      return NextResponse.redirect(new URL('/register', request.url))
    }

    // If users exist and trying to access register, redirect to login
    if (count > 0 && path === '/register') {
      return NextResponse.redirect(new URL('/login', request.url))
    }
  } catch (error) {
    console.error('Error checking users:', error)
  }

  // If the path is public and user is logged in, redirect to dashboard
  if ((isPublicPath || path === '/') && token) {
    return NextResponse.redirect(new URL('/dashboard', request.url))
  }

  // If the path is protected and user is not logged in, redirect to login
  if (!isPublicPath && path !== '/' && !token) {
    const response = NextResponse.redirect(new URL('/login', request.url))
    
    // Add return-to path to search params
    const returnTo = new URL(request.url).pathname
    response.cookies.set('returnTo', returnTo, {
      path: '/',
      httpOnly: true,
      sameSite: 'lax',
      secure: process.env.NODE_ENV === 'production'
    })
    
    return response
  }

  return NextResponse.next()
}

// Configure the paths that should be handled by the middleware
export const config = {
  matcher: ['/', '/login', '/register', '/dashboard/:path*']
} 