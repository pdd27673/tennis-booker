import axios, { AxiosError } from 'axios'
import type { AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import { useAppStore } from '@/stores/appStore'
import { tokenStorage } from '@/lib/tokenStorage'
import { mockAuthApi } from './mockAuthApi'

// Create a separate axios instance for API calls
export const apiClient = axios.create({
  baseURL: '/api', // This would be your actual API base URL
  timeout: 10000,
})

// Flag to prevent multiple refresh attempts
let isRefreshing = false
let failedQueue: Array<{
  resolve: (value: string) => void
  reject: (error: any) => void
}> = []

// Process the queue of failed requests after token refresh
const processQueue = (error: any, token: string | null = null) => {
  failedQueue.forEach(({ resolve, reject }) => {
    if (error) {
      reject(error)
    } else {
      resolve(token!)
    }
  })
  
  failedQueue = []
}

// Request interceptor to add auth token
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = tokenStorage.getAccessToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor to handle token refresh
apiClient.interceptors.response.use(
  (response: AxiosResponse) => {
    return response
  },
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean }
    
    // Check if error is 401 and we haven't already tried to refresh
    if (error.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        // If already refreshing, queue this request
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject })
        }).then(token => {
          originalRequest.headers.Authorization = `Bearer ${token}`
          return apiClient(originalRequest)
        }).catch(err => {
          return Promise.reject(err)
        })
      }

      originalRequest._retry = true
      isRefreshing = true

      try {
        const refreshToken = tokenStorage.getRefreshToken()
        
        if (!refreshToken) {
          throw new Error('No refresh token available')
        }

        console.log('🔄 Attempting token refresh...')
        
        // Call the refresh token endpoint
        const refreshResponse = await mockAuthApi.refreshTokenEndpoint(refreshToken)
        
        if (refreshResponse.success && refreshResponse.data) {
          const { tokens } = refreshResponse.data
          
          // Update tokens in storage and store
          tokenStorage.setAccessToken(tokens.accessToken)
          tokenStorage.setRefreshToken(tokens.refreshToken)
          
          // Update Zustand store
          const { updateTokens, addNotification } = useAppStore.getState()
          updateTokens(tokens.accessToken, tokens.refreshToken)
          
          console.log('✅ Token refresh successful')
          
          addNotification({
            title: 'Session Renewed',
            message: 'Your session has been automatically renewed',
            type: 'info',
          })
          
          // Process queued requests
          processQueue(null, tokens.accessToken)
          
          // Retry original request with new token
          originalRequest.headers.Authorization = `Bearer ${tokens.accessToken}`
          return apiClient(originalRequest)
        } else {
          throw new Error(refreshResponse.error || 'Token refresh failed')
        }
      } catch (refreshError) {
        console.error('❌ Token refresh failed:', refreshError)
        
        // Process queue with error
        processQueue(refreshError, null)
        
        // Clear auth state and redirect to login
        const { clearAuthState, addNotification } = useAppStore.getState()
        clearAuthState()
        
        addNotification({
          title: 'Session Expired',
          message: 'Please log in again to continue',
          type: 'warning',
        })
        
        // Redirect to login page
        window.location.href = '/login'
        
        return Promise.reject(refreshError)
      } finally {
        isRefreshing = false
      }
    }

    return Promise.reject(error)
  }
)

// Helper function to make authenticated API calls
export const makeAuthenticatedRequest = async <T>(
  method: 'GET' | 'POST' | 'PUT' | 'DELETE',
  url: string,
  data?: any
): Promise<T> => {
  const response = await apiClient.request({
    method,
    url,
    data,
  })
  
  return response.data
}

// Convenience methods for different HTTP verbs
export const apiGet = <T>(url: string): Promise<T> => 
  makeAuthenticatedRequest<T>('GET', url)

export const apiPost = <T>(url: string, data?: any): Promise<T> => 
  makeAuthenticatedRequest<T>('POST', url, data)

export const apiPut = <T>(url: string, data?: any): Promise<T> => 
  makeAuthenticatedRequest<T>('PUT', url, data)

export const apiDelete = <T>(url: string): Promise<T> => 
  makeAuthenticatedRequest<T>('DELETE', url)

// Test function to simulate protected API call
export const testProtectedEndpoint = async (): Promise<{
  success: boolean
  data?: any
  error?: string
}> => {
  try {
    const accessToken = tokenStorage.getAccessToken()
    if (!accessToken) {
      return {
        success: false,
        error: 'No access token available'
      }
    }
    
    const response = await mockAuthApi.getProtectedData(accessToken)
    return response
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Unknown error'
    }
  }
}

// Test function to simulate expired token scenario
export const testExpiredTokenEndpoint = async (): Promise<{
  success: boolean
  data?: any
  error?: string
}> => {
  try {
    const response = await mockAuthApi.getProtectedDataWithExpiredToken()
    return response
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Unknown error'
    }
  }
} 