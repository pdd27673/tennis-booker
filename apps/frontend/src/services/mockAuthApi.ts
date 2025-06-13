import type { LoginFormData, RegisterFormData } from '@/lib/auth-schemas'

// Mock user database
const mockUsers = [
  {
    id: '1',
    name: 'Demo User',
    email: 'demo@example.com',
    password: 'password123', // In real app, this would be hashed
  },
  {
    id: '2',
    name: 'Test User',
    email: 'test@example.com',
    password: 'test123',
  },
]

// Mock JWT tokens
const generateMockTokens = (userId: string) => {
  const accessToken = `mock-access-token-${userId}-${Date.now()}`
  const refreshToken = `mock-refresh-token-${userId}-${Date.now()}`
  
  return {
    accessToken,
    refreshToken,
    expiresIn: 3600, // 1 hour
  }
}

// Types for our mock API responses
interface AuthResponse {
  success: boolean
  error?: string
  status?: number
  data?: {
    user: {
      id: string
      name: string
      email: string
    }
    tokens: {
      accessToken: string
      refreshToken: string
      expiresIn: number
    }
  }
}

interface ProtectedDataResponse {
  success: boolean
  error?: string
  status?: number
  data?: {
    data: string
    timestamp: number
  }
}

interface RefreshTokenResponse {
  success: boolean
  error?: string
  status?: number
  data?: {
    tokens: {
      accessToken: string
      refreshToken: string
      expiresIn: number
    }
  }
}

// Simulate network delay
const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

export const mockAuthApi = {
  async login(credentials: LoginFormData): Promise<AuthResponse> {
    await delay(800) // Simulate network delay
    
    const user = mockUsers.find(
      u => u.email === credentials.email && u.password === credentials.password
    )
    
    if (!user) {
      return {
        success: false,
        error: 'Invalid email or password',
      }
    }
    
    const tokens = generateMockTokens(user.id)
    
    return {
      success: true,
      data: {
        user: {
          id: user.id,
          name: user.name,
          email: user.email,
        },
        tokens,
      },
    }
  },

  async register(userData: RegisterFormData): Promise<AuthResponse> {
    await delay(1000) // Simulate network delay
    
    // Check if user already exists
    const existingUser = mockUsers.find(u => u.email === userData.email)
    if (existingUser) {
      return {
        success: false,
        error: 'User with this email already exists',
      }
    }
    
    // Create new user
    const newUser = {
      id: `${mockUsers.length + 1}`,
      name: userData.name,
      email: userData.email,
      password: userData.password,
    }
    
    mockUsers.push(newUser)
    
    const tokens = generateMockTokens(newUser.id)
    
    return {
      success: true,
      data: {
        user: {
          id: newUser.id,
          name: newUser.name,
          email: newUser.email,
        },
        tokens,
      },
    }
  },

  async refreshToken(refreshToken: string): Promise<AuthResponse> {
    await delay(300)
    
    // Simple validation - in real app, this would verify the refresh token
    if (!refreshToken || !refreshToken.startsWith('mock-refresh-token-')) {
      return {
        success: false,
        error: 'Invalid refresh token',
      }
    }
    
    // Extract user ID from token (mock implementation)
    const userId = refreshToken.split('-')[3]
    const user = mockUsers.find(u => u.id === userId)
    
    if (!user) {
      return {
        success: false,
        error: 'User not found',
      }
    }
    
    const tokens = generateMockTokens(user.id)
    
    return {
      success: true,
      data: {
        user: {
          id: user.id,
          name: user.name,
          email: user.email,
        },
        tokens,
      },
    }
  },

  async getMe(accessToken: string): Promise<AuthResponse> {
    await delay(200)
    
    // Simple validation - in real app, this would verify the access token
    if (!accessToken || !accessToken.startsWith('mock-access-token-')) {
      return {
        success: false,
        error: 'Invalid access token',
      }
    }
    
    // Extract user ID from token (mock implementation)
    const userId = accessToken.split('-')[3]
    const user = mockUsers.find(u => u.id === userId)
    
    if (!user) {
      return {
        success: false,
        error: 'User not found',
      }
    }
    
    return {
      success: true,
      data: {
        user: {
          id: user.id,
          name: user.name,
          email: user.email,
        },
        tokens: {
          accessToken,
          refreshToken: '', // Not needed for getMe
          expiresIn: 3600,
        },
      },
    }
  },

  // Mock refresh token endpoint
  async refreshTokenEndpoint(refreshToken: string): Promise<RefreshTokenResponse> {
    // Simulate network delay
    await delay(500)
    
    // Validate refresh token format
    if (!refreshToken || !refreshToken.startsWith('refresh_')) {
      return {
        success: false,
        error: 'Invalid refresh token format'
      }
    }
    
    // Check if refresh token exists in our mock database
    const tokenData = refreshToken.split('_')[1] // Extract user identifier
    const user = mockUsers.find(u => u.id === tokenData)
    
    if (!user) {
      return {
        success: false,
        error: 'Invalid refresh token'
      }
    }
    
    // Generate new tokens
    const newAccessToken = `access_${user.id}_${Date.now()}`
    const newRefreshToken = `refresh_${user.id}_${Date.now()}`
    
    console.log('🔄 Token refresh successful:', {
      userId: user.id,
      newAccessToken: newAccessToken.substring(0, 20) + '...',
      newRefreshToken: newRefreshToken.substring(0, 20) + '...',
    })
    
    return {
      success: true,
      data: {
        tokens: {
          accessToken: newAccessToken,
          refreshToken: newRefreshToken,
          expiresIn: 300, // 5 minutes
        }
      }
    }
  },

  // Mock protected endpoint that requires authentication
  async getProtectedData(accessToken: string): Promise<ProtectedDataResponse> {
    // Simulate network delay
    await delay(300)
    
    // Validate access token format
    if (!accessToken || !accessToken.startsWith('access_')) {
      return {
        success: false,
        error: 'Unauthorized - Invalid token format',
        status: 401
      }
    }
    
    // Extract token parts
    const tokenParts = accessToken.split('_')
    if (tokenParts.length !== 3) {
      return {
        success: false,
        error: 'Unauthorized - Malformed token',
        status: 401
      }
    }
    
    const userId = tokenParts[1]
    const timestamp = parseInt(tokenParts[2] || '0')
    
    // Check if user exists
    const user = mockUsers.find(u => u.id === userId)
    if (!user) {
      return {
        success: false,
        error: 'Unauthorized - User not found',
        status: 401
      }
    }
    
    // Check if token is expired (simulate 5 minute expiry)
    const tokenAge = Date.now() - timestamp
    const fiveMinutes = 5 * 60 * 1000
    
    if (tokenAge > fiveMinutes) {
      console.log('🔒 Access token expired:', {
        tokenAge: Math.round(tokenAge / 1000) + 's',
        maxAge: Math.round(fiveMinutes / 1000) + 's'
      })
      return {
        success: false,
        error: 'Unauthorized - Token expired',
        status: 401
      }
    }
    
    // Return protected data
    return {
      success: true,
      data: {
        data: `Protected data for ${user.name}`,
        timestamp: Date.now()
      }
    }
  },

  // Mock endpoint to simulate expired token (for testing)
  async getProtectedDataWithExpiredToken(): Promise<ProtectedDataResponse> {
    // Always return 401 to simulate expired token
    await delay(300)
    
    return {
      success: false,
      error: 'Unauthorized - Token expired',
      status: 401
    }
  },
} 