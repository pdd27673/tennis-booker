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

export const authApi = {
  // Login user
  async login(credentials: LoginFormData): Promise<AuthResponse> {
    try {
      const response = await authApiClient.post('/api/auth/login', credentials)
      
      return {
        success: true,
        data: response.data
      }
      } catch (error) {
      const axiosError = error as AxiosError
      
      if (axiosError.response) {
        const errorData = axiosError.response.data as { message?: string; error?: string }
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
      const response = await authApiClient.post('/api/auth/register', userData)
      
      return {
        success: true,
        data: response.data
      }
      } catch (error) {
      const axiosError = error as AxiosError
      
      if (axiosError.response) {
        const errorData = axiosError.response.data as { message?: string; error?: string }
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

      
      return {
        success: true,
        data: response.data
      }
    } catch (error) {

      const axiosError = error as AxiosError
      
      if (axiosError.response) {
        const errorData = axiosError.response.data as { message?: string; error?: string }
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

      const response = await authApiClient.post('/api/auth/refresh', {
        refreshToken
      })

      
      return {
        success: true,
        data: response.data
      }
    } catch (error) {

      const axiosError = error as AxiosError
      
      if (axiosError.response) {
        const errorData = axiosError.response.data as { message?: string; error?: string }
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

      
      if (token) {
        await authApiClient.post('/api/auth/logout', {}, {
          headers: {
            Authorization: `Bearer ${token}`
          }
        })
      }
      

      return { success: true }
    } catch {

      // Even if logout fails on server, we consider it successful on client
      return { success: true }
    }
  }
}

export default authApi 