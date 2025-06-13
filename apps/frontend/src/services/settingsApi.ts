import type { UserPreferences, SystemControlState, SystemStatus } from '@/stores/appStore'

// Mock API delay for realistic behavior
const mockDelay = (ms: number = 1000) => new Promise(resolve => setTimeout(resolve, ms))

// User Preferences API
export const userPreferencesApi = {
  // Get user preferences
  getUserPreferences: async (): Promise<UserPreferences> => {
    await mockDelay(500)
    
    // Simulate API call - in real implementation, this would fetch from backend
    const mockPreferences: UserPreferences = {
          preferredClubs: ['Tennis Club Central', 'Riverside Tennis Club'],
    preferredTimeSlots: ['18:00-20:00', '09:00-11:00'],
    notificationEmail: 'user@example.com',
    enableEmailNotifications: true,
    enablePushNotifications: false,
    maxDistance: 10,
    enableAutoBooking: false,
    }
    
    return mockPreferences
  },

  // Update user preferences
  updateUserPreferences: async (preferences: UserPreferences): Promise<UserPreferences> => {
    await mockDelay(1000)
    
    // Simulate potential API errors (5% chance)
    if (Math.random() < 0.05) {
      throw new Error('Failed to save preferences. Please try again.')
    }
    
    // Simulate API call - in real implementation, this would send to backend
    console.log('API: Updating user preferences:', preferences)
    
    // Return the updated preferences (backend would validate and return)
    return preferences
  },

  // Reset user preferences to defaults
  resetUserPreferences: async (): Promise<UserPreferences> => {
    await mockDelay(800)
    
    const defaultPreferences: UserPreferences = {
      preferredClubs: ['Tennis Club Central'],
      preferredTimeSlots: ['18:00-20:00'],
      notificationEmail: '',
      enableEmailNotifications: true,
      enablePushNotifications: false,
      maxDistance: 10,
      enableAutoBooking: false,
    }
    
    console.log('API: Resetting user preferences to defaults')
    return defaultPreferences
  },
}

// System Control API
export const systemControlApi = {
  // Get current system status
  getSystemStatus: async (): Promise<SystemControlState> => {
    await mockDelay(300)
    
    // Simulate API call - in real implementation, this would fetch from backend
    const mockSystemState: SystemControlState = {
      systemStatus: 'RUNNING',
      lastUpdate: new Date(),
      isScrapingActive: true,
      systemInfo: {
        monitoredClubs: 12,
        lastScan: new Date(Date.now() - 2 * 60 * 1000), // 2 minutes ago
        nextScan: new Date(Date.now() + 3 * 60 * 1000), // 3 minutes from now
        averageResponseTime: 1.2,
        successRate: 98.5,
        courtsFoundToday: 47,
      },
    }
    
    return mockSystemState
  },

  // Pause scraping system
  pauseScraping: async (): Promise<{ status: SystemStatus; message: string }> => {
    await mockDelay(800)
    
    // Simulate potential API errors (3% chance)
    if (Math.random() < 0.03) {
      throw new Error('Failed to pause scraping system. Please try again.')
    }
    
    console.log('API: Pausing scraping system')
    
    return {
      status: 'PAUSED',
      message: 'Scraping system has been paused successfully.',
    }
  },

  // Resume scraping system
  resumeScraping: async (): Promise<{ status: SystemStatus; message: string }> => {
    await mockDelay(800)
    
    // Simulate potential API errors (3% chance)
    if (Math.random() < 0.03) {
      throw new Error('Failed to resume scraping system. Please try again.')
    }
    
    console.log('API: Resuming scraping system')
    
    return {
      status: 'RUNNING',
      message: 'Scraping system has been resumed successfully.',
    }
  },

  // Restart scraping system
  restartSystem: async (): Promise<{ status: SystemStatus; message: string }> => {
    await mockDelay(1500) // Restart takes longer
    
    // Simulate potential API errors (5% chance)
    if (Math.random() < 0.05) {
      throw new Error('Failed to restart system. Please try again.')
    }
    
    console.log('API: Restarting scraping system')
    
    return {
      status: 'RUNNING',
      message: 'System has been restarted successfully.',
    }
  },

  // Update system configuration
  updateSystemConfig: async (config: {
    monitoredClubs?: number
    scanInterval?: number
    maxRetries?: number
  }): Promise<{ message: string }> => {
    await mockDelay(1000)
    
    console.log('API: Updating system configuration:', config)
    
    return {
      message: 'System configuration updated successfully.',
    }
  },

  // Get system health metrics
  getSystemHealth: async (): Promise<{
    uptime: number
    memoryUsage: number
    cpuUsage: number
    activeConnections: number
  }> => {
    await mockDelay(400)
    
    // Simulate real-time metrics
    return {
      uptime: Math.floor(Math.random() * 86400), // Random uptime in seconds
      memoryUsage: Math.floor(Math.random() * 80) + 20, // 20-100%
      cpuUsage: Math.floor(Math.random() * 60) + 10, // 10-70%
      activeConnections: Math.floor(Math.random() * 50) + 5, // 5-55 connections
    }
  },
}

// Combined API object for easy importing
export const settingsApi = {
  userPreferences: userPreferencesApi,
  systemControl: systemControlApi,
}

// API response types
export type UserPreferencesResponse = UserPreferences
export type SystemControlResponse = SystemControlState
export type SystemActionResponse = { status: SystemStatus; message: string }
export type SystemHealthResponse = {
  uptime: number
  memoryUsage: number
  cpuUsage: number
  activeConnections: number
} 