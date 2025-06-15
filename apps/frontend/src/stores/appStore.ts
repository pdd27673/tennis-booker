import { create } from 'zustand'
import { devtools, persist } from 'zustand/middleware'
import { tokenStorage } from '@/lib/tokenStorage'

export type Theme = 'light' | 'dark' | 'system'
export type SystemStatus = 'IDLE' | 'RUNNING' | 'PAUSED' | 'ERROR'

export interface UserProfile {
  id: string
  name: string
  email: string
  avatar?: string
}

// Authentication state interface
export interface AuthState {
  isAuthenticated: boolean
  user: UserProfile | null
  accessToken: string | null
  refreshToken: string | null
}

// Time range interface to match backend
export interface TimeRange {
  start: string // Format: "HH:MM"
  end: string   // Format: "HH:MM"
}

// Notification settings interface to match backend
export interface NotificationSettings {
  email: boolean
  emailAddress?: string
  instantAlerts: boolean
  maxAlertsPerHour: number
  maxAlertsPerDay: number
  alertTimeWindowStart: string
  alertTimeWindowEnd: string
  unsubscribed: boolean
}

// User preferences interface to match backend
export interface UserPreferences {
  id?: string
  userId?: string
  times?: TimeRange[]           // Legacy field for backward compatibility
  weekdayTimes?: TimeRange[]    // Monday-Friday preferred times
  weekendTimes?: TimeRange[]    // Saturday-Sunday preferred times
  preferredVenues: string[]
  excludedVenues: string[]
  preferredDays: string[]
  maxPrice: number
  notificationSettings: NotificationSettings
  createdAt?: string
  updatedAt?: string
}

// System control interface
export interface SystemControlState {
  systemStatus: SystemStatus
  lastUpdate: Date | null
  isScrapingActive: boolean
  systemInfo: {
    monitoredClubs: number
    lastScan: Date | null
    nextScan: Date | null
    averageResponseTime: number
    successRate: number
    courtsFoundToday: number
  }
}

interface AppState {
  // Authentication state
  isAuthenticated: boolean
  userProfile: UserProfile | null
  accessToken: string | null
  refreshToken: string | null
  
  // UI preferences
  theme: Theme
  sidebarCollapsed: boolean
  
  // Application state
  activeTab: string
  notifications: Array<{
    id: string
    type: 'info' | 'success' | 'warning' | 'error'
    title: string
    message: string
    timestamp: Date
  }>
  
  // User preferences state
  userPreferences: UserPreferences
  isPreferencesLoading: boolean
  preferencesError: string | null
  
  // System control state
  systemControl: SystemControlState
  isSystemControlLoading: boolean
  systemControlError: string | null
  
  // Authentication actions
  setAuthState: (authData: {
    user: UserProfile
    accessToken: string
    refreshToken: string
  }) => void
  clearAuthState: () => void
  updateTokens: (accessToken: string, refreshToken?: string) => void
  
  // User preferences actions
  setUserPreferences: (preferences: UserPreferences) => void
  updateUserPreferences: (preferences: Partial<UserPreferences>) => void
  setPreferencesLoading: (loading: boolean) => void
  setPreferencesError: (error: string | null) => void
  resetUserPreferences: () => void
  
  // System control actions
  setSystemStatus: (status: SystemStatus) => void
  updateSystemInfo: (info: Partial<SystemControlState['systemInfo']>) => void
  setSystemControlLoading: (loading: boolean) => void
  setSystemControlError: (error: string | null) => void
  requestPauseScraping: () => void
  requestResumeScraping: () => void
  requestRestartSystem: () => void
  updateLastScanTime: () => void
  
  // Existing actions
  login: (profile: UserProfile) => void
  logout: () => void
  setTheme: (theme: Theme) => void
  setSidebarCollapsed: (collapsed: boolean) => void
  setActiveTab: (tab: string) => void
  addNotification: (notification: Omit<AppState['notifications'][0], 'id' | 'timestamp'>) => void
  removeNotification: (id: string) => void
  clearNotifications: () => void
}

// Default user preferences
const defaultUserPreferences: UserPreferences = {
  weekdayTimes: [
    { start: '18:00', end: '20:00' }
  ],
  weekendTimes: [
    { start: '09:00', end: '11:00' }
  ],
  preferredVenues: ['Tennis Club Central', 'Riverside Tennis Club'],
  excludedVenues: [],
  preferredDays: ['monday', 'tuesday', 'wednesday', 'thursday', 'friday'],
  maxPrice: 100.0,
  notificationSettings: {
    email: true,
    instantAlerts: true,
    maxAlertsPerHour: 10,
    maxAlertsPerDay: 50,
    alertTimeWindowStart: '07:00',
    alertTimeWindowEnd: '22:00',
    unsubscribed: false,
  },
}

// Default system control state
const defaultSystemControl: SystemControlState = {
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

export const useAppStore = create<AppState>()(
  devtools(
    persist(
      (set, get) => ({
        // Initialize authentication state from token storage
        ...(() => {
          const tokens = tokenStorage.initializeFromStorage()
          return {
            isAuthenticated: false, // Will be set to true after token validation
            userProfile: null,
            accessToken: tokens.accessToken,
            refreshToken: tokens.refreshToken,
          }
        })(),
        
        // UI preferences
        theme: 'system',
        sidebarCollapsed: false,
        activeTab: 'dashboard',
        notifications: [],

        // User preferences state
        userPreferences: defaultUserPreferences,
        isPreferencesLoading: false,
        preferencesError: null,

        // System control state
        systemControl: defaultSystemControl,
        isSystemControlLoading: false,
        systemControlError: null,

        // Authentication actions
        setAuthState: (authData) => {
          // Store tokens using token storage utility
          tokenStorage.setAccessToken(authData.accessToken)
          tokenStorage.setRefreshToken(authData.refreshToken)
          
          set(
            {
              isAuthenticated: true,
              userProfile: authData.user,
              accessToken: authData.accessToken,
              refreshToken: authData.refreshToken,
              // Update notification email with user's email
              userPreferences: {
                ...get().userPreferences,
                notificationSettings: {
                  ...get().userPreferences.notificationSettings,
                  emailAddress: authData.user.email,
                },
              },
            },
            false,
            'setAuthState'
          )
        },

        clearAuthState: () => {
          // Clear tokens from storage
          tokenStorage.clearAllTokens()
          
          set(
            {
              isAuthenticated: false,
              userProfile: null,
              accessToken: null,
              refreshToken: null,
              // Reset preferences to defaults on logout
              userPreferences: defaultUserPreferences,
              // Reset system control to defaults on logout
              systemControl: defaultSystemControl,
            },
            false,
            'clearAuthState'
          )
        },

        updateTokens: (accessToken, refreshToken) => {
          // Update tokens in storage
          tokenStorage.setAccessToken(accessToken)
          if (refreshToken) {
            tokenStorage.setRefreshToken(refreshToken)
          }
          
          set(
            (state) => ({
              accessToken,
              refreshToken: refreshToken || state.refreshToken,
            }),
            false,
            'updateTokens'
          )
        },

        // User preferences actions
        setUserPreferences: (preferences) => {
          set(
            { userPreferences: preferences },
            false,
            'setUserPreferences'
          )
        },

        updateUserPreferences: (preferences) => {
          set(
            (state) => ({
              userPreferences: {
                ...state.userPreferences,
                ...preferences,
              },
            }),
            false,
            'updateUserPreferences'
          )
        },

        setPreferencesLoading: (loading) => {
          set(
            { isPreferencesLoading: loading },
            false,
            'setPreferencesLoading'
          )
        },

        setPreferencesError: (error) => {
          set(
            { preferencesError: error },
            false,
            'setPreferencesError'
          )
        },

        resetUserPreferences: () => {
          set(
            { userPreferences: defaultUserPreferences },
            false,
            'resetUserPreferences'
          )
        },

        // System control actions
        setSystemStatus: (status) => {
          set(
            (state) => ({
              systemControl: {
                ...state.systemControl,
                systemStatus: status,
                lastUpdate: new Date(),
                isScrapingActive: status === 'RUNNING',
              },
            }),
            false,
            'setSystemStatus'
          )
        },

        updateSystemInfo: (info) => {
          set(
            (state) => ({
              systemControl: {
                ...state.systemControl,
                systemInfo: {
                  ...state.systemControl.systemInfo,
                  ...info,
                },
                lastUpdate: new Date(),
              },
            }),
            false,
            'updateSystemInfo'
          )
        },

        setSystemControlLoading: (loading) => {
          set(
            { isSystemControlLoading: loading },
            false,
            'setSystemControlLoading'
          )
        },

        setSystemControlError: (error) => {
          set(
            { systemControlError: error },
            false,
            'setSystemControlError'
          )
        },

        requestPauseScraping: () => {
          const { addNotification } = get()
          
          set(
            (state) => ({
              systemControl: {
                ...state.systemControl,
                systemStatus: 'PAUSED',
                isScrapingActive: false,
                lastUpdate: new Date(),
              },
            }),
            false,
            'requestPauseScraping'
          )

          addNotification({
            type: 'info',
            title: 'Scraping Paused',
            message: 'Court monitoring has been temporarily paused.',
          })
        },

        requestResumeScraping: () => {
          const { addNotification } = get()
          
          set(
            (state) => ({
              systemControl: {
                ...state.systemControl,
                systemStatus: 'RUNNING',
                isScrapingActive: true,
                lastUpdate: new Date(),
                systemInfo: {
                  ...state.systemControl.systemInfo,
                  nextScan: new Date(Date.now() + 3 * 60 * 1000), // 3 minutes from now
                },
              },
            }),
            false,
            'requestResumeScraping'
          )

          addNotification({
            type: 'success',
            title: 'Scraping Resumed',
            message: 'Court monitoring has been resumed.',
          })
        },

        requestRestartSystem: () => {
          const { addNotification } = get()
          
          set(
            (state) => ({
              systemControl: {
                ...state.systemControl,
                systemStatus: 'RUNNING',
                isScrapingActive: true,
                lastUpdate: new Date(),
                systemInfo: {
                  ...state.systemControl.systemInfo,
                  lastScan: new Date(),
                  nextScan: new Date(Date.now() + 5 * 60 * 1000), // 5 minutes from now
                },
              },
            }),
            false,
            'requestRestartSystem'
          )

          addNotification({
            type: 'warning',
            title: 'System Restarting',
            message: 'The monitoring system is being restarted...',
          })
        },

        updateLastScanTime: () => {
          set(
            (state) => ({
              systemControl: {
                ...state.systemControl,
                systemInfo: {
                  ...state.systemControl.systemInfo,
                  lastScan: new Date(),
                  nextScan: new Date(Date.now() + 5 * 60 * 1000), // 5 minutes from now
                },
                lastUpdate: new Date(),
              },
            }),
            false,
            'updateLastScanTime'
          )
        },

        // Legacy login action (kept for backward compatibility)
        login: (profile: UserProfile) => {
          set(
            {
              isAuthenticated: true,
              userProfile: profile,
              // Update notification email with user's email
              userPreferences: {
                ...get().userPreferences,
                notificationSettings: {
                  ...get().userPreferences.notificationSettings,
                  emailAddress: profile.email,
                },
              },
            },
            false,
            'login'
          )
        },

        // Enhanced logout action
        logout: () => {
          // Clear tokens from storage
          tokenStorage.clearAllTokens()
          
          set(
            {
              isAuthenticated: false,
              userProfile: null,
              accessToken: null,
              refreshToken: null,
              // Reset preferences to defaults on logout
              userPreferences: defaultUserPreferences,
              // Reset system control to defaults on logout
              systemControl: defaultSystemControl,
            },
            false,
            'logout'
          )
        },

        setTheme: (theme: Theme) => {
          set({ theme }, false, 'setTheme')
        },

        setSidebarCollapsed: (collapsed: boolean) => {
          set({ sidebarCollapsed: collapsed }, false, 'setSidebarCollapsed')
        },

        setActiveTab: (tab: string) => {
          set({ activeTab: tab }, false, 'setActiveTab')
        },

        addNotification: (notification) => {
          const newNotification = {
            ...notification,
            id: crypto.randomUUID(),
            timestamp: new Date(),
          }
          set(
            (state) => ({
              notifications: [...state.notifications, newNotification],
            }),
            false,
            'addNotification'
          )
        },

        removeNotification: (id: string) => {
          set(
            (state) => ({
              notifications: state.notifications.filter((n) => n.id !== id),
            }),
            false,
            'removeNotification'
          )
        },

        clearNotifications: () => {
          set({ notifications: [] }, false, 'clearNotifications')
        },
      }),
      {
        name: 'tennis-booker-app-store',
        partialize: (state) => ({
          theme: state.theme,
          sidebarCollapsed: state.sidebarCollapsed,
          activeTab: state.activeTab,
          userPreferences: state.userPreferences,
          systemControl: state.systemControl,
          // Note: Authentication state is NOT persisted in localStorage
          // Tokens are managed separately by tokenStorage utility
          // UI preferences, user preferences, and system control state are persisted
        }),
      }
    ),
    {
      name: 'AppStore',
    }
  )
) 