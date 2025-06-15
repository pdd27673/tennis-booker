import axios, { AxiosError } from 'axios'
import config from '@/config/config'
import { tokenStorage } from '@/lib/tokenStorage'
import type { LoginFormData, RegisterFormData } from '@/lib/auth-schemas'

// Create axios instance for auth API calls
const authApiClient = axios.create({
  baseURL: config.apiUrl,
  timeout: config.apiTimeout,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Types for API responses
interface AuthResponse {
  success: boolean
  error?: string
  data?: {
    user: {
      id: string
      name: string
      username: string
      email: string
    }
    accessToken: string
    refreshToken: string
  }
}

interface UserInfoResponse {
  success: boolean
  error?: string
  data?: {
    user: {
      id: string
      name: string
      username: string
      email: string
    }
  }
}

// Handle API errors consistently
const handleAuthError = (error: AxiosError) => {
  if (error.response) {
    const errorData = error.response.data as any
    throw new Error(errorData.message || errorData.error || 'Authentication failed')
  } else if (error.request) {
    throw new Error('Network error. Please check your connection.')
  } else {
    throw new Error('An unexpected error occurred.')
  }
}

export const authApi = {
  // Login user
  async login(credentials: LoginFormData): Promise<AuthResponse> {
    try {
      console.log('🔐 AuthAPI: Attempting login for:', credentials.email)
      const response = await authApiClient.post('/api/auth/login', credentials)
      console.log('✅ AuthAPI: Login successful:', response.data)
      
      return {
        success: true,
        data: response.data
      }
    } catch (error) {
      console.error('❌ AuthAPI: Login failed:', error)
      const axiosError = error as AxiosError
      
      if (axiosError.response) {
        const errorData = axiosError.response.data as any
        return {
          success: false,
          error: errorData.message || errorData.error || 'Login failed'
        }
      }
      
      return {
        success: false,
        error: 'Network error. Please check your connection.'
      }
    }
  },

  // Register user
  async register(userData: RegisterFormData): Promise<AuthResponse> {
    try {
      console.log('📝 AuthAPI: Attempting registration for:', userData.email)
      const response = await authApiClient.post('/api/auth/register', userData)
      console.log('✅ AuthAPI: Registration successful:', response.data)
      
      return {
        success: true,
        data: response.data
      }
    } catch (error) {
      console.error('❌ AuthAPI: Registration failed:', error)
      const axiosError = error as AxiosError
      
      if (axiosError.response) {
        const errorData = axiosError.response.data as any
        return {
          success: false,
          error: errorData.message || errorData.error || 'Registration failed'
        }
      }
      
      return {
        success: false,
        error: 'Network error. Please check your connection.'
      }
    }
  },

  // Get current user info
  async getMe(accessToken?: string): Promise<UserInfoResponse> {
    try {
      const token = accessToken || tokenStorage.getAccessToken()
      console.log('👤 AuthAPI: Fetching user info with token:', token ? 'Present' : 'Missing')
      
      if (!token) {
        return {
          success: false,
          error: 'No access token available'
        }
      }

      const response = await authApiClient.get('/api/auth/me', {
        headers: {
          Authorization: `Bearer ${token}`
        }
      })
      console.log('✅ AuthAPI: User info fetched:', response.data)
      
      return {
        success: true,
        data: response.data
      }
    } catch (error) {
      console.error('❌ AuthAPI: Failed to fetch user info:', error)
      const axiosError = error as AxiosError
      
      if (axiosError.response) {
        const errorData = axiosError.response.data as any
        return {
          success: false,
          error: errorData.message || errorData.error || 'Failed to fetch user info'
        }
      }
      
      return {
        success: false,
        error: 'Network error. Please check your connection.'
      }
    }
  },

  // Refresh access token
  async refreshToken(refreshToken: string): Promise<AuthResponse> {
    try {
      console.log('🔄 AuthAPI: Refreshing access token')
      const response = await authApiClient.post('/api/auth/refresh', {
        refreshToken
      })
      console.log('✅ AuthAPI: Token refreshed successfully')
      
      return {
        success: true,
        data: response.data
      }
    } catch (error) {
      console.error('❌ AuthAPI: Token refresh failed:', error)
      const axiosError = error as AxiosError
      
      if (axiosError.response) {
        const errorData = axiosError.response.data as any
        return {
          success: false,
          error: errorData.message || errorData.error || 'Token refresh failed'
        }
      }
      
      return {
        success: false,
        error: 'Network error. Please check your connection.'
      }
    }
  },

  // Logout user (optional - mainly for clearing server-side sessions if needed)
  async logout(): Promise<{ success: boolean; error?: string }> {
    try {
      const token = tokenStorage.getAccessToken()
      console.log('🚪 AuthAPI: Logging out user')
      
      if (token) {
        await authApiClient.post('/api/auth/logout', {}, {
          headers: {
            Authorization: `Bearer ${token}`
          }
        })
      }
      
      console.log('✅ AuthAPI: Logout successful')
      return { success: true }
    } catch (error) {
      console.error('❌ AuthAPI: Logout failed:', error)
      // Even if logout fails on server, we consider it successful on client
      return { success: true }
    }
  }
}

export default authApi 