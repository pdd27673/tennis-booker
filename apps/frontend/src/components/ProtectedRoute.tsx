import { useEffect } from 'react'
import type { ReactNode } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { useAppStore } from '@/stores/appStore'
import { useTokenInitialization } from '@/hooks/useTokenInitialization'

interface ProtectedRouteProps {
  children: ReactNode
  redirectTo?: string
}

export function ProtectedRoute({ 
  children, 
  redirectTo = '/login' 
}: ProtectedRouteProps) {
  const { isAuthenticated, accessToken, refreshToken } = useAppStore()
  const { isInitializing } = useTokenInitialization()
  const location = useLocation()

  // Check if user is authenticated
  // We check isAuthenticated flag OR if we have both access and refresh tokens
  const hasValidTokens = accessToken && refreshToken
  const isUserAuthenticated = isAuthenticated || hasValidTokens

  // IMPORTANT: All hooks must be called before any conditional returns
  useEffect(() => {
    // Log authentication check for debugging (only when not initializing)
    if (!isInitializing) {
      console.log('ProtectedRoute check:', {
        path: location.pathname,
        isAuthenticated,
        hasAccessToken: !!accessToken,
        hasRefreshToken: !!refreshToken,
        willRedirect: !isUserAuthenticated,
      })
    }
  }, [location.pathname, isAuthenticated, accessToken, refreshToken, isUserAuthenticated, isInitializing])

  // Show loading while initializing tokens
  if (isInitializing) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-2 text-gray-600">Initializing...</p>
        </div>
      </div>
    )
  }

  // If not authenticated, redirect to login with return URL
  if (!isUserAuthenticated) {
    // Store the attempted URL so we can redirect back after login
    const returnUrl = location.pathname + location.search
    const redirectUrl = `${redirectTo}?returnUrl=${encodeURIComponent(returnUrl)}`
    
    console.log('Redirecting unauthenticated user:', {
      from: returnUrl,
      to: redirectUrl,
    })
    
    return <Navigate to={redirectUrl} replace />
  }

  // User is authenticated, render the protected content
  return <>{children}</>
}

// Higher-order component version for easier usage
export function withProtectedRoute<P extends object>(
  Component: React.ComponentType<P>,
  redirectTo?: string
) {
  return function ProtectedComponent(props: P) {
    return (
      <ProtectedRoute redirectTo={redirectTo}>
        <Component {...props} />
      </ProtectedRoute>
    )
  }
} 