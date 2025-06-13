import { useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAppStore } from '@/stores/appStore'
import { mockAuthApi } from '@/services/mockAuthApi'
import type { LoginFormData, RegisterFormData } from '@/lib/auth-schemas'

export interface UseAuthReturn {
  // Authentication state
  isAuthenticated: boolean
  user: {
    id: string
    name: string
    email: string
    avatar?: string
  } | null
  accessToken: string | null
  refreshToken: string | null
  
  // Authentication functions
  login: (credentials: LoginFormData) => Promise<{
    success: boolean
    error?: string
  }>
  register: (userData: RegisterFormData) => Promise<{
    success: boolean
    error?: string
  }>
  logout: () => void
  
  // User information functions
  refreshUserInfo: () => Promise<{
    success: boolean
    error?: string
  }>
  
  // Loading states
  isLoading: boolean
}

export function useAuth(): UseAuthReturn {
  const {
    isAuthenticated,
    userProfile: user,
    accessToken,
    refreshToken,
    setAuthState,
    clearAuthState,
    addNotification,
  } = useAppStore()
  
  const navigate = useNavigate()

  // Login function with enhanced user information fetching
  const login = useCallback(async (credentials: LoginFormData) => {
    try {
      const response = await mockAuthApi.login(credentials)
      
      if (response.success && response.data) {
        const { user: userData, tokens } = response.data
        
        // Set authentication state
        setAuthState({
          user: userData,
          accessToken: tokens.accessToken,
          refreshToken: tokens.refreshToken,
        })
        
        // Fetch additional user information using the access token
        try {
          const userInfoResponse = await mockAuthApi.getMe(tokens.accessToken)
          if (userInfoResponse.success && userInfoResponse.data) {
            // Update user information if we got more details
            setAuthState({
              user: userInfoResponse.data.user,
              accessToken: tokens.accessToken,
              refreshToken: tokens.refreshToken,
            })
          }
        } catch (error) {
          console.warn('Failed to fetch additional user info:', error)
          // Continue with basic user info from login response
        }
        
        addNotification({
          title: 'Login Successful',
          message: `Welcome back, ${userData.name}!`,
          type: 'success',
        })
        
        return { success: true }
      } else {
        addNotification({
          title: 'Login Failed',
          message: response.error || 'Invalid credentials',
          type: 'error',
        })
        
        return { 
          success: false, 
          error: response.error || 'Login failed' 
        }
      }
    } catch (error) {
      const errorMessage = 'An unexpected error occurred during login'
      addNotification({
        title: 'Login Error',
        message: errorMessage,
        type: 'error',
      })
      
      return { 
        success: false, 
        error: errorMessage 
      }
    }
  }, [setAuthState, addNotification])

  // Register function
  const register = useCallback(async (userData: RegisterFormData) => {
    try {
      const response = await mockAuthApi.register(userData)
      
      if (response.success && response.data) {
        const { user: newUser, tokens } = response.data
        
        // Set authentication state
        setAuthState({
          user: newUser,
          accessToken: tokens.accessToken,
          refreshToken: tokens.refreshToken,
        })
        
        addNotification({
          title: 'Registration Successful',
          message: `Welcome, ${newUser.name}! Your account has been created.`,
          type: 'success',
        })
        
        return { success: true }
      } else {
        addNotification({
          title: 'Registration Failed',
          message: response.error || 'Failed to create account',
          type: 'error',
        })
        
        return { 
          success: false, 
          error: response.error || 'Registration failed' 
        }
      }
    } catch (error) {
      const errorMessage = 'An unexpected error occurred during registration'
      addNotification({
        title: 'Registration Error',
        message: errorMessage,
        type: 'error',
      })
      
      return { 
        success: false, 
        error: errorMessage 
      }
    }
  }, [setAuthState, addNotification])

  // Enhanced logout function
  const logout = useCallback(() => {
    clearAuthState()
    addNotification({
      title: 'Logged Out',
      message: 'You have been successfully logged out',
      type: 'info',
    })
    navigate('/login')
  }, [clearAuthState, addNotification, navigate])

  // Refresh user information
  const refreshUserInfo = useCallback(async () => {
    if (!accessToken) {
      return { 
        success: false, 
        error: 'No access token available' 
      }
    }
    
    try {
      const response = await mockAuthApi.getMe(accessToken)
      
      if (response.success && response.data) {
        // Update user information while preserving tokens
        setAuthState({
          user: response.data.user,
          accessToken,
          refreshToken: refreshToken || '',
        })
        
        return { success: true }
      } else {
        return { 
          success: false, 
          error: response.error || 'Failed to fetch user information' 
        }
      }
    } catch (error) {
      return { 
        success: false, 
        error: 'An unexpected error occurred while fetching user information' 
      }
    }
  }, [accessToken, refreshToken, setAuthState])

  return {
    // Authentication state
    isAuthenticated,
    user,
    accessToken,
    refreshToken,
    
    // Authentication functions
    login,
    register,
    logout,
    
    // User information functions
    refreshUserInfo,
    
    // Loading states (for future enhancement)
    isLoading: false, // TODO: Implement loading state management
  }
} 