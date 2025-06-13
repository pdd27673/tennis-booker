// Token storage utility for managing JWT tokens
// Access token is stored in memory for security
// Refresh token is stored in localStorage for persistence

let accessToken: string | null = null

export const tokenStorage = {
  // Access Token (in-memory storage)
  setAccessToken: (token: string | null) => {
    accessToken = token
  },

  getAccessToken: (): string | null => {
    return accessToken
  },

  clearAccessToken: () => {
    accessToken = null
  },

  // Refresh Token (localStorage for persistence)
  setRefreshToken: (token: string | null) => {
    if (token) {
      localStorage.setItem('refreshToken', token)
    } else {
      localStorage.removeItem('refreshToken')
    }
  },

  getRefreshToken: (): string | null => {
    try {
      return localStorage.getItem('refreshToken')
    } catch (error) {
      console.warn('Failed to access localStorage for refresh token:', error)
      return null
    }
  },

  clearRefreshToken: () => {
    try {
      localStorage.removeItem('refreshToken')
    } catch (error) {
      console.warn('Failed to clear refresh token from localStorage:', error)
    }
  },

  // Clear all tokens
  clearAllTokens: () => {
    accessToken = null
    try {
      localStorage.removeItem('refreshToken')
    } catch (error) {
      console.warn('Failed to clear tokens:', error)
    }
  },

  // Check if user has valid tokens
  hasValidTokens: (): boolean => {
    return !!(accessToken && tokenStorage.getRefreshToken())
  },

  // Initialize tokens from storage on app start
  initializeFromStorage: () => {
    // Access token is not persisted (in-memory only)
    // Only refresh token is restored from localStorage
    const refreshToken = tokenStorage.getRefreshToken()
    return {
      accessToken: null, // Always null on app start for security
      refreshToken,
    }
  },
}

// Security note: 
// - Access tokens are stored in memory only and lost on page refresh
// - Refresh tokens are stored in localStorage for persistence
// - In production, consider using HttpOnly cookies managed by the backend
// - This approach balances security with user experience for a demo app 