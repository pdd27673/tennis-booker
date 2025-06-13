import { useEffect } from 'react'
import type { ReactNode } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { useAppStore } from '@/stores/appStore'

interface ProtectedRouteProps {
  children: ReactNode
  redirectTo?: string
}

export function ProtectedRoute({ 
  children, 
  redirectTo = '/login' 
}: ProtectedRouteProps) {
  const { isAuthenticated, accessToken } = useAppStore()
  const location = useLocation()

  // Check if user is authenticated
  // We check both isAuthenticated flag and presence of accessToken
  const isUserAuthenticated = isAuthenticated && accessToken

  useEffect(() => {
    // Log authentication check for debugging
    console.log('ProtectedRoute check:', {
      path: location.pathname,
      isAuthenticated,
      hasAccessToken: !!accessToken,
      willRedirect: !isUserAuthenticated,
    })
  }, [location.pathname, isAuthenticated, accessToken, isUserAuthenticated])

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