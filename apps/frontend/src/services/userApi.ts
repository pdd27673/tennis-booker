import { AxiosError } from 'axios'
import { apiClient } from './tokenRefreshService'
import type { UserPreferences } from '@/stores/appStore'

// Notification types
export interface UserNotification {
  id: string
  userId: string
  venueName: string
  courtName: string
  date: string
  time: string
  price: number
  emailSent: boolean
  emailStatus: string
  slotKey: string
  createdAt: string
  type: string
}

export interface NotificationResponse {
  notifications: UserNotification[]
  pagination: {
    page: number
    limit: number
    total: number
    totalPages: number
  }
}

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
      const response = await apiClient.get('/api/users/preferences')
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
      const response = await apiClient.put('/api/users/preferences', preferences)
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

  // Get user notifications
  async getUserNotifications(page: number = 0, limit: number = 50): Promise<NotificationResponse> {
    try {
      console.log(`📩 UserAPI: Fetching notifications page ${page}, limit ${limit}...`)
      const response = await apiClient.get(`/api/users/notifications?page=${page}&limit=${limit}`)
      console.log('✅ UserAPI: Successfully fetched notifications:', response.data)
      return response.data
    } catch (error) {
      console.error('❌ UserAPI: Failed to fetch notifications:', error)
      handleUserError(error as AxiosError)
      throw error
    }
  },
}

export default userApi 