import type { UserPreferences, SystemControlState } from '@/stores/appStore'
import { userApi } from './userApi'
import { systemApi } from './systemApi'

// User Preferences API
export const userPreferencesApi = {
  // Get user preferences
  getUserPreferences: userApi.getUserPreferences,

  // Update user preferences
  updateUserPreferences: userApi.updateUserPreferences,

  // Reset user preferences to defaults
  resetUserPreferences: userApi.resetUserPreferences,
}

// System Control API
export const systemControlApi = {
  // Get current system status
  getSystemStatus: systemApi.getSystemStatus,

  // Pause scraping system
  pauseScraping: systemApi.pauseScraping,

  // Resume scraping system
  resumeScraping: systemApi.resumeScraping,

  // Restart scraping system
  restartSystem: systemApi.restartSystem,

  // Update system configuration - placeholder for future implementation
  updateSystemConfig: async (config: {
    monitoredClubs?: number
    scanInterval?: number
    maxRetries?: number
  }): Promise<{ message: string }> => {
    console.log('API: Updating system configuration:', config)
    // This endpoint doesn't exist in backend yet
    return {
      message: 'System configuration feature is not yet implemented.',
    }
  },

  // Get system health metrics
  getSystemHealth: systemApi.getHealth,
}

// Combined API object for easy importing
export const settingsApi = {
  userPreferences: userPreferencesApi,
  systemControl: systemControlApi,
}

// API response types
export type UserPreferencesResponse = UserPreferences
export type SystemControlResponse = SystemControlState
export type SystemActionResponse = { status: string; message: string }
export type SystemHealthResponse = {
  uptime: number
  memoryUsage: number
  cpuUsage: number
  activeConnections: number
} 