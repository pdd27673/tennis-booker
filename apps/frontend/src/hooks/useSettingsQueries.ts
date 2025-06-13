import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { settingsApi } from '@/services/settingsApi'
import { useAppStore } from '@/stores/appStore'
import type { 
  UserPreferencesResponse,
  SystemActionResponse
} from '@/services/settingsApi'
import React from 'react'

// Query keys for React Query
export const settingsQueryKeys = {
  userPreferences: ['userPreferences'] as const,
  systemStatus: ['systemStatus'] as const,
  systemHealth: ['systemHealth'] as const,
}

// User Preferences Hooks
export const useUserPreferences = () => {
  const { setUserPreferences, setPreferencesError } = useAppStore()

  const query = useQuery({
    queryKey: settingsQueryKeys.userPreferences,
    queryFn: settingsApi.userPreferences.getUserPreferences,
    staleTime: 5 * 60 * 1000, // 5 minutes
  })

  // Handle success/error in useEffect
  React.useEffect(() => {
    if (query.data) {
      setUserPreferences(query.data)
      setPreferencesError(null)
    }
    if (query.error) {
      console.error('Failed to fetch user preferences:', query.error)
      setPreferencesError(query.error.message)
    }
  }, [query.data, query.error, setUserPreferences, setPreferencesError])

  return query
}

export const useUpdateUserPreferences = () => {
  const queryClient = useQueryClient()
  const { 
    setUserPreferences, 
    setPreferencesLoading, 
    setPreferencesError, 
    addNotification 
  } = useAppStore()

  return useMutation({
    mutationFn: settingsApi.userPreferences.updateUserPreferences,
    onMutate: () => {
      setPreferencesLoading(true)
      setPreferencesError(null)
    },
    onSuccess: (data: UserPreferencesResponse) => {
      // Update the store
      setUserPreferences(data)
      setPreferencesError(null)
      
      // Invalidate and refetch user preferences
      queryClient.invalidateQueries({ queryKey: settingsQueryKeys.userPreferences })
      
      // Show success notification
      addNotification({
        type: 'success',
        title: 'Preferences Saved',
        message: 'Your preferences have been updated successfully.',
      })
    },
    onError: (error: Error) => {
      console.error('Failed to update user preferences:', error)
      setPreferencesError(error.message)
      
      // Show error notification
      addNotification({
        type: 'error',
        title: 'Save Failed',
        message: error.message || 'There was an error saving your preferences. Please try again.',
      })
    },
    onSettled: () => {
      setPreferencesLoading(false)
    },
  })
}

export const useResetUserPreferences = () => {
  const queryClient = useQueryClient()
  const { 
    setUserPreferences, 
    setPreferencesLoading, 
    setPreferencesError, 
    addNotification 
  } = useAppStore()

  return useMutation({
    mutationFn: settingsApi.userPreferences.resetUserPreferences,
    onMutate: () => {
      setPreferencesLoading(true)
      setPreferencesError(null)
    },
    onSuccess: (data: UserPreferencesResponse) => {
      // Update the store
      setUserPreferences(data)
      setPreferencesError(null)
      
      // Invalidate and refetch user preferences
      queryClient.invalidateQueries({ queryKey: settingsQueryKeys.userPreferences })
      
      // Show success notification
      addNotification({
        type: 'info',
        title: 'Preferences Reset',
        message: 'Your preferences have been reset to defaults.',
      })
    },
    onError: (error: Error) => {
      console.error('Failed to reset user preferences:', error)
      setPreferencesError(error.message)
      
      // Show error notification
      addNotification({
        type: 'error',
        title: 'Reset Failed',
        message: error.message || 'There was an error resetting your preferences. Please try again.',
      })
    },
    onSettled: () => {
      setPreferencesLoading(false)
    },
  })
}

// System Control Hooks
export const useSystemStatus = () => {
  const { 
    setSystemControlError,
    updateSystemInfo,
    setSystemStatus 
  } = useAppStore()

  const query = useQuery({
    queryKey: settingsQueryKeys.systemStatus,
    queryFn: settingsApi.systemControl.getSystemStatus,
    staleTime: 30 * 1000, // 30 seconds
    refetchInterval: 60 * 1000, // Refetch every minute
  })

  // Handle success/error in useEffect
  React.useEffect(() => {
    if (query.data) {
      // Update system info in store
      updateSystemInfo(query.data.systemInfo)
      setSystemStatus(query.data.systemStatus)
      setSystemControlError(null)
    }
    if (query.error) {
      console.error('Failed to fetch system status:', query.error)
      setSystemControlError(query.error.message)
    }
  }, [query.data, query.error, updateSystemInfo, setSystemStatus, setSystemControlError])

  return query
}

export const usePauseScraping = () => {
  const queryClient = useQueryClient()
  const { 
    setSystemControlLoading, 
    setSystemControlError, 
    setSystemStatus,
    addNotification 
  } = useAppStore()

  return useMutation({
    mutationFn: settingsApi.systemControl.pauseScraping,
    onMutate: () => {
      setSystemControlLoading(true)
      setSystemControlError(null)
    },
    onSuccess: (data: SystemActionResponse) => {
      // Update system status
      setSystemStatus(data.status)
      setSystemControlError(null)
      
      // Invalidate system status query
      queryClient.invalidateQueries({ queryKey: settingsQueryKeys.systemStatus })
      
      // Show success notification
      addNotification({
        type: 'info',
        title: 'Scraping Paused',
        message: data.message,
      })
    },
    onError: (error: Error) => {
      console.error('Failed to pause scraping:', error)
      setSystemControlError(error.message)
      
      // Show error notification
      addNotification({
        type: 'error',
        title: 'Pause Failed',
        message: error.message || 'Failed to pause scraping system. Please try again.',
      })
    },
    onSettled: () => {
      setSystemControlLoading(false)
    },
  })
}

export const useResumeScraping = () => {
  const queryClient = useQueryClient()
  const { 
    setSystemControlLoading, 
    setSystemControlError, 
    setSystemStatus,
    addNotification 
  } = useAppStore()

  return useMutation({
    mutationFn: settingsApi.systemControl.resumeScraping,
    onMutate: () => {
      setSystemControlLoading(true)
      setSystemControlError(null)
    },
    onSuccess: (data: SystemActionResponse) => {
      // Update system status
      setSystemStatus(data.status)
      setSystemControlError(null)
      
      // Invalidate system status query
      queryClient.invalidateQueries({ queryKey: settingsQueryKeys.systemStatus })
      
      // Show success notification
      addNotification({
        type: 'success',
        title: 'Scraping Resumed',
        message: data.message,
      })
    },
    onError: (error: Error) => {
      console.error('Failed to resume scraping:', error)
      setSystemControlError(error.message)
      
      // Show error notification
      addNotification({
        type: 'error',
        title: 'Resume Failed',
        message: error.message || 'Failed to resume scraping system. Please try again.',
      })
    },
    onSettled: () => {
      setSystemControlLoading(false)
    },
  })
}

export const useRestartSystem = () => {
  const queryClient = useQueryClient()
  const { 
    setSystemControlLoading, 
    setSystemControlError, 
    setSystemStatus,
    updateLastScanTime,
    addNotification 
  } = useAppStore()

  return useMutation({
    mutationFn: settingsApi.systemControl.restartSystem,
    onMutate: () => {
      setSystemControlLoading(true)
      setSystemControlError(null)
    },
    onSuccess: (data: SystemActionResponse) => {
      // Update system status and scan time
      setSystemStatus(data.status)
      updateLastScanTime()
      setSystemControlError(null)
      
      // Invalidate system status query
      queryClient.invalidateQueries({ queryKey: settingsQueryKeys.systemStatus })
      
      // Show success notification
      addNotification({
        type: 'warning',
        title: 'System Restarted',
        message: data.message,
      })
    },
    onError: (error: Error) => {
      console.error('Failed to restart system:', error)
      setSystemControlError(error.message)
      
      // Show error notification
      addNotification({
        type: 'error',
        title: 'Restart Failed',
        message: error.message || 'Failed to restart system. Please try again.',
      })
    },
    onSettled: () => {
      setSystemControlLoading(false)
    },
  })
}

export const useSystemHealth = () => {
  const { setSystemControlError } = useAppStore()

  const query = useQuery({
    queryKey: settingsQueryKeys.systemHealth,
    queryFn: settingsApi.systemControl.getSystemHealth,
    staleTime: 10 * 1000, // 10 seconds
    refetchInterval: 30 * 1000, // Refetch every 30 seconds
  })

  // Handle error in useEffect
  React.useEffect(() => {
    if (query.error) {
      console.error('Failed to fetch system health:', query.error)
      setSystemControlError(query.error.message)
    }
  }, [query.error, setSystemControlError])

  return query
}

// Combined hook for all settings operations
export const useSettings = () => {
  const userPreferencesQuery = useUserPreferences()
  const systemStatusQuery = useSystemStatus()
  const systemHealthQuery = useSystemHealth()
  
  const updateUserPreferencesMutation = useUpdateUserPreferences()
  const resetUserPreferencesMutation = useResetUserPreferences()
  
  const pauseScrapingMutation = usePauseScraping()
  const resumeScrapingMutation = useResumeScraping()
  const restartSystemMutation = useRestartSystem()

  return {
    // Queries
    userPreferences: userPreferencesQuery,
    systemStatus: systemStatusQuery,
    systemHealth: systemHealthQuery,
    
    // User Preferences Mutations
    updateUserPreferences: updateUserPreferencesMutation,
    resetUserPreferences: resetUserPreferencesMutation,
    
    // System Control Mutations
    pauseScraping: pauseScrapingMutation,
    resumeScraping: resumeScrapingMutation,
    restartSystem: restartSystemMutation,
    
    // Loading states
    isLoading: userPreferencesQuery.isLoading || systemStatusQuery.isLoading,
    isUpdating: updateUserPreferencesMutation.isPending || 
                pauseScrapingMutation.isPending || 
                resumeScrapingMutation.isPending || 
                restartSystemMutation.isPending,
  }
} 