import { useEffect, useState } from 'react'
import { useAppStore } from '@/stores/appStore'
import { authApi } from '@/services/authApi'
import { tokenStorage } from '@/lib/tokenStorage'

export function useTokenInitialization() {
  const [isInitializing, setIsInitializing] = useState(true)
  const { setAuthState, clearAuthState, refreshToken } = useAppStore()

  useEffect(() => {
    const initializeTokens = async () => {
      try {
        // Check if we have a refresh token
        const storedRefreshToken = refreshToken || tokenStorage.getRefreshToken()
        
        if (!storedRefreshToken) {
          // No refresh token, user needs to login
          setIsInitializing(false)
          return
        }

        console.log('🔄 Initializing tokens from refresh token...')
        
        // Try to refresh the access token
        const refreshResponse = await authApi.refreshToken(storedRefreshToken)
        
        if (refreshResponse.success && refreshResponse.data) {
          const { user, accessToken, refreshToken: newRefreshToken } = refreshResponse.data
          
          // Set authentication state
          setAuthState({
            user: {
              id: user.id,
              name: user.name || user.username,
              email: user.email,
              avatar: undefined,
            },
            accessToken,
            refreshToken: newRefreshToken,
          })
          
          console.log('✅ Token initialization successful')
        } else {
          // Refresh token is invalid, clear auth state
          console.log('❌ Token refresh failed, clearing auth state')
          clearAuthState()
        }
      } catch (error) {
        console.error('Token initialization error:', error)
        // Clear auth state on error
        clearAuthState()
      } finally {
        setIsInitializing(false)
      }
    }

    initializeTokens()
  }, [setAuthState, clearAuthState, refreshToken])

  return { isInitializing }
} 