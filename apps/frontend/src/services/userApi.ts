import { AxiosError } from 'axios'
import { apiClient } from './tokenRefreshService'
import type { UserPreferences } from '@/stores/appStore'

// Default preferences that match backend structure
const getDefaultPreferences = (): UserPreferences => ({
  times: [],
  weekdayTimes: [{ start: '18:00', end: '20:00' }],
  weekendTimes: [{ start: '09:00', end: '11:00' }],
  preferredVenues: [],
  excludedVenues: [],
  preferredDays: ['monday', 'tuesday', 'wednesday', 'thursday', 'friday'],
  maxPrice: 100.0,
  notificationSettings: {
    email: true,
    instantAlerts: true,
    maxAlertsPerHour: 10,
    maxAlertsPerDay: 50,
    alertTimeWindowStart: "07:00",
    alertTimeWindowEnd: "22:00",
    unsubscribed: false,
  },
})

// Handle API errors consistently
const handleUserError = (error: AxiosError) => {
  if (error.response) {
    const errorData = error.response.data as { message?: string; error?: string }
    throw new Error(errorData.message || errorData.error || 'User operation failed')
  } else if (error.request) {
    throw new Error('Network error. Please check your connection.')
  } else {
    throw new Error('An unexpected error occurred.')
  }
}

export const userApi = {
  // Get user preferences
  async getUserPreferences(): Promise<UserPreferences> {
    try {
      console.log('📋 UserAPI: Fetching user preferences...')
      // Use the dedicated preferences endpoint
      const response = await apiClient.get('/users/preferences')
      console.log('✅ UserAPI: Successfully fetched preferences:', response.data)
      return response.data
    } catch (error) {
      console.error('❌ UserAPI: Failed to fetch user preferences:', error)
      // Return default preferences if API call fails
      const defaults = getDefaultPreferences()
      console.log('🔄 UserAPI: Using default preferences:', defaults)
      return defaults
    }
  },

  // Update user preferences
  async updateUserPreferences(preferences: UserPreferences): Promise<UserPreferences> {
    try {
      console.log('💾 UserAPI: Updating user preferences:', preferences)
      // Send preferences directly to backend (no transformation needed)
      const response = await apiClient.put('/users/preferences', preferences)
      console.log('✅ UserAPI: Successfully updated preferences:', response.data)
      return response.data
    } catch (error) {
      console.error('❌ UserAPI: Failed to update preferences:', error)
      handleUserError(error as AxiosError)
      throw error
    }
  },

  // Reset user preferences to defaults
  async resetUserPreferences(): Promise<UserPreferences> {
    try {
      // Send the default preferences to the backend
      return await this.updateUserPreferences(getDefaultPreferences())
    } catch (error) {
      console.error('Failed to reset preferences on backend:', error)
      // Return default preferences even if backend call fails
      return getDefaultPreferences()
    }
  },
}

export default userApi 