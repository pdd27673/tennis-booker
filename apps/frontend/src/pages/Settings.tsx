import React from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { useAuth } from '@/hooks/useAuth'
import { useAppStore } from '@/stores/appStore'
import { 
  useUpdateUserPreferences,
  usePauseScraping,
  useResumeScraping,
  useRestartSystem,
  useSystemStatus,
  useUserPreferences,
  useResetUserPreferences,
  useSystemHealth
} from '@/hooks/useSettingsQueries'
import SystemStatusDisplay from '@/components/SystemStatusDisplay'
import { useNavigate } from 'react-router-dom'
import { 
  Clock, 
  MapPin, 
  Bell, 
  Settings as SettingsIcon, 
  Save, 
  ArrowLeft,
  Play,
  Pause,
  RotateCcw,
  Zap
} from 'lucide-react'
import { Separator } from '@/components/ui/separator'

// Form validation schema
const userPreferencesSchema = z.object({
  preferredClubs: z.array(z.string()).min(1, 'Select at least one club'),
  preferredTimeSlots: z.array(z.string()).min(1, 'Select at least one time slot'),
  notificationEmail: z.string().email('Please enter a valid email address'),
  maxDistance: z.number().min(1, 'Distance must be at least 1 km').max(100, 'Distance cannot exceed 100 km'),
  enableEmailNotifications: z.boolean(),
  enablePushNotifications: z.boolean(),
  enableAutoBooking: z.boolean(),
})

type UserPreferencesForm = z.infer<typeof userPreferencesSchema>

const Settings: React.FC = () => {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  
  // Zustand store state (for UI state and notifications)
  const {
    userPreferences,
    systemControl,
    preferencesError,
    systemControlError,
  } = useAppStore()

  // React Query hooks
  const userPreferencesQuery = useUserPreferences()
  const systemStatusQuery = useSystemStatus()
  const updateUserPreferencesMutation = useUpdateUserPreferences()
  const resetUserPreferencesMutation = useResetUserPreferences()
  const pauseScrapingMutation = usePauseScraping()
  const resumeScrapingMutation = useResumeScraping()
  const restartSystemMutation = useRestartSystem()
  const systemHealthQuery = useSystemHealth()

  const form = useForm<UserPreferencesForm>({
    resolver: zodResolver(userPreferencesSchema),
    defaultValues: {
      preferredClubs: userPreferences?.preferredClubs || [],
      preferredTimeSlots: userPreferences?.preferredTimeSlots || [],
      notificationEmail: userPreferences?.notificationEmail || '',
      maxDistance: userPreferences?.maxDistance || 10,
      enableEmailNotifications: userPreferences?.enableEmailNotifications || true,
      enablePushNotifications: userPreferences?.enablePushNotifications || false,
      enableAutoBooking: userPreferences?.enableAutoBooking || false,
    },
  })

  // Initialize form with store data
  React.useEffect(() => {
    if (userPreferences) {
      form.reset({
        preferredClubs: userPreferences.preferredClubs,
        preferredTimeSlots: userPreferences.preferredTimeSlots,
        notificationEmail: userPreferences.notificationEmail,
        maxDistance: userPreferences.maxDistance,
        enableEmailNotifications: userPreferences.enableEmailNotifications,
        enablePushNotifications: userPreferences.enablePushNotifications,
        enableAutoBooking: userPreferences.enableAutoBooking,
      })
    }
  }, [userPreferences, form])

  const onSubmit = (data: UserPreferencesForm) => {
    updateUserPreferencesMutation.mutate(data)
  }

  // Reset preferences handler
  const handleResetPreferences = () => {
    resetUserPreferencesMutation.mutate()
  }

  // Format time for display
  const formatTimeAgo = (date: Date | null) => {
    if (!date) return 'Never'
    
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMinutes = Math.floor(diffMs / (1000 * 60))
    
    if (diffMinutes < 1) return 'Just now'
    if (diffMinutes < 60) return `${diffMinutes} minutes ago`
    
    const diffHours = Math.floor(diffMinutes / 60)
    if (diffHours < 24) return `${diffHours} hours ago`
    
    const diffDays = Math.floor(diffHours / 24)
    return `${diffDays} days ago`
  }

  const formatTimeUntil = (date: Date | null) => {
    if (!date) return 'Unknown'
    
    const now = new Date()
    const diffMs = date.getTime() - now.getTime()
    const diffMinutes = Math.floor(diffMs / (1000 * 60))
    
    if (diffMinutes < 1) return 'Now'
    if (diffMinutes < 60) return `in ${diffMinutes} minutes`
    
    const diffHours = Math.floor(diffMinutes / 60)
    if (diffHours < 24) return `in ${diffHours} hours`
    
    const diffDays = Math.floor(diffHours / 24)
    return `in ${diffDays} days`
  }

  // System control handlers - now using React Query mutations
  const handlePauseScraping = () => {
    console.log('Pause scraping requested')
    pauseScrapingMutation.mutate()
  }

  const handleResumeScraping = () => {
    console.log('Resume scraping requested')
    resumeScrapingMutation.mutate()
  }

  const handleRestartSystem = () => {
    console.log('System restart requested')
    restartSystemMutation.mutate()
  }

  const handleLogout = () => {
    logout()
  }

  // Loading states from React Query
  const isPreferencesLoading = updateUserPreferencesMutation.isPending || resetUserPreferencesMutation.isPending
  const isSystemControlLoading = pauseScrapingMutation.isPending || 
                                  resumeScrapingMutation.isPending || 
                                  restartSystemMutation.isPending

  // Available options
  const availableClubs = [
    'Tennis Club Central',
    'City Sports Complex',
    'Riverside Tennis Center',
    'Elite Tennis Academy',
    'Community Recreation Center'
  ]

  const availableTimeSlots = [
    '06:00-08:00',
    '08:00-10:00',
    '10:00-12:00',
    '12:00-14:00',
    '14:00-16:00',
    '16:00-18:00',
    '18:00-20:00',
    '20:00-22:00'
  ]

  if (isPreferencesLoading || isSystemControlLoading) {
    return (
      <div className="container mx-auto p-6">
        <div className="flex items-center justify-center h-64">
          <div className="text-lg">Loading settings...</div>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-gray-50 to-gray-100 dark:from-gray-900 dark:via-gray-900 dark:to-gray-800">
      {/* Header */}
      <header className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm shadow-lg border-b border-gray-200/50 dark:border-gray-700/50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-6">
            <div className="flex items-center space-x-4">
              <Button
                variant="outline"
                size="sm"
                onClick={() => navigate('/dashboard')}
                className="hover:scale-105 transition-transform duration-200"
              >
                <ArrowLeft className="w-4 h-4 mr-2" />
                Back to Dashboard
              </Button>
              <div className="flex items-center space-x-3">
                <SettingsIcon className="w-8 h-8 text-blue-600 dark:text-blue-400" />
                <div>
                  <h1 className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white">
                    Settings
                  </h1>
                  <p className="text-gray-600 dark:text-gray-400">
                    Manage your preferences and system settings
                  </p>
                </div>
              </div>
            </div>
            <div className="flex items-center space-x-3">
              <Badge variant="secondary" className="bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100">
                {user?.name}
              </Badge>
              <Button 
                variant="destructive" 
                onClick={handleLogout}
                className="hover:scale-105 transition-transform duration-200"
              >
                Logout
              </Button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-4xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
        <div className="space-y-8">
          
          {/* Error Display */}
          {(preferencesError || systemControlError) && (
            <Card className="border-red-200 bg-red-50 dark:border-red-800 dark:bg-red-900/20">
              <CardContent className="pt-6">
                <div className="flex items-center space-x-2 text-red-800 dark:text-red-200">
                  <Bell className="w-5 h-5" />
                  <p className="font-medium">
                    {preferencesError || systemControlError}
                  </p>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Loading Indicator */}
          {(userPreferencesQuery.isLoading || systemStatusQuery.isLoading) && (
            <Card className="border-blue-200 bg-blue-50 dark:border-blue-800 dark:bg-blue-900/20">
              <CardContent className="pt-6">
                <div className="flex items-center space-x-2 text-blue-800 dark:text-blue-200">
                  <div className="w-5 h-5 border-2 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
                  <p className="font-medium">Loading settings...</p>
                </div>
              </CardContent>
            </Card>
          )}

          {/* User Preferences Section */}
          <Card className="border-0 shadow-lg bg-white/70 dark:bg-gray-800/70 backdrop-blur-sm">
            <CardHeader>
              <div className="flex items-center space-x-3">
                <div className="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
                  <Bell className="w-5 h-5 text-blue-600 dark:text-blue-400" />
                </div>
                <div>
                  <CardTitle className="text-xl text-gray-900 dark:text-white">
                    User Preferences
                  </CardTitle>
                  <CardDescription className="text-gray-600 dark:text-gray-400">
                    Configure your tennis court monitoring preferences
                  </CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                  
                  {/* Preferred Clubs */}
                  <FormField
                    control={form.control}
                    name="preferredClubs"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel className="flex items-center space-x-2">
                          <MapPin className="w-4 h-4" />
                          <span>Preferred Tennis Clubs</span>
                        </FormLabel>
                        <FormControl>
                          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                            {availableClubs.map((club) => (
                              <div key={club} className="flex items-center space-x-2">
                                <input
                                  type="checkbox"
                                  id={club}
                                  checked={field.value.includes(club)}
                                  onChange={(e) => {
                                    if (e.target.checked) {
                                      field.onChange([...field.value, club])
                                    } else {
                                      field.onChange(field.value.filter((c) => c !== club))
                                    }
                                  }}
                                  className="rounded border-gray-300"
                                />
                                <label htmlFor={club} className="text-sm font-medium">
                                  {club}
                                </label>
                              </div>
                            ))}
                          </div>
                        </FormControl>
                        <FormDescription>
                          Select the clubs you'd like to monitor for available courts
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  {/* Preferred Time Slots */}
                  <FormField
                    control={form.control}
                    name="preferredTimeSlots"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel className="flex items-center space-x-2">
                          <Clock className="w-4 h-4" />
                          <span>Preferred Time Slots</span>
                        </FormLabel>
                        <FormControl>
                          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-2">
                            {availableTimeSlots.map((slot) => (
                              <div key={slot} className="flex items-center space-x-2">
                                <input
                                  type="checkbox"
                                  id={slot}
                                  checked={field.value.includes(slot)}
                                  onChange={(e) => {
                                    if (e.target.checked) {
                                      field.onChange([...field.value, slot])
                                    } else {
                                      field.onChange(field.value.filter((s) => s !== slot))
                                    }
                                  }}
                                  className="rounded border-gray-300"
                                />
                                <label htmlFor={slot} className="text-sm font-medium text-center">
                                  {slot}
                                </label>
                              </div>
                            ))}
                          </div>
                        </FormControl>
                        <FormDescription>
                          Select your preferred playing times
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  {/* Notification Email */}
                  <FormField
                    control={form.control}
                    name="notificationEmail"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Notification Email</FormLabel>
                        <FormControl>
                          <Input
                            type="email"
                            placeholder="your.email@example.com"
                            {...field}
                            className="transition-all duration-200 focus:ring-2 focus:ring-blue-500"
                          />
                        </FormControl>
                        <FormDescription>
                          Email address for court availability notifications
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  {/* Max Distance */}
                  <FormField
                    control={form.control}
                    name="maxDistance"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Maximum Distance (km)</FormLabel>
                        <FormControl>
                          <Input
                            type="number"
                            min="1"
                            max="100"
                            {...field}
                            onChange={(e) => field.onChange(parseInt(e.target.value))}
                          />
                        </FormControl>
                        <FormDescription>
                          Maximum distance from your location to search for courts
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  {/* Notification Settings */}
                  <div className="space-y-4">
                    <h3 className="text-lg font-medium">Notification Settings</h3>
                    
                    <FormField
                      control={form.control}
                      name="enableEmailNotifications"
                      render={({ field }) => (
                        <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                          <div className="space-y-0.5">
                            <FormLabel className="text-base">Email Notifications</FormLabel>
                            <FormDescription>
                              Receive email alerts when courts become available
                            </FormDescription>
                          </div>
                          <FormControl>
                            <Switch
                              checked={field.value}
                              onCheckedChange={field.onChange}
                            />
                          </FormControl>
                        </FormItem>
                      )}
                    />

                    <FormField
                      control={form.control}
                      name="enablePushNotifications"
                      render={({ field }) => (
                        <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                          <div className="space-y-0.5">
                            <FormLabel className="text-base">Push Notifications</FormLabel>
                            <FormDescription>
                              Receive browser push notifications for immediate alerts
                            </FormDescription>
                          </div>
                          <FormControl>
                            <Switch
                              checked={field.value}
                              onCheckedChange={field.onChange}
                            />
                          </FormControl>
                        </FormItem>
                      )}
                    />
                  </div>

                  {/* Auto Booking */}
                  <FormField
                    control={form.control}
                    name="enableAutoBooking"
                    render={({ field }) => (
                      <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                        <div className="space-y-0.5">
                          <FormLabel className="text-base">Auto-Booking</FormLabel>
                          <FormDescription>
                            Automatically book courts when they match your preferences
                          </FormDescription>
                        </div>
                        <FormControl>
                          <Switch
                            checked={field.value}
                            onCheckedChange={field.onChange}
                          />
                        </FormControl>
                      </FormItem>
                    )}
                  />

                  {/* Form Actions */}
                  <div className="flex space-x-4 pt-6">
                    <Button
                      type="submit"
                      disabled={isPreferencesLoading}
                      className="hover:scale-105 transition-transform duration-200 min-w-[120px]"
                    >
                      {updateUserPreferencesMutation.isPending ? (
                        <div className="flex items-center space-x-2">
                          <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                          <span>Saving...</span>
                        </div>
                      ) : (
                        <div className="flex items-center space-x-2">
                          <Save className="w-4 h-4" />
                          <span>Save Preferences</span>
                        </div>
                      )}
                    </Button>
                    <Button
                      type="button"
                      variant="outline"
                      onClick={handleResetPreferences}
                      disabled={isPreferencesLoading}
                      className="hover:scale-105 transition-transform duration-200 min-w-[120px]"
                    >
                      {resetUserPreferencesMutation.isPending ? (
                        <div className="flex items-center space-x-2">
                          <div className="w-4 h-4 border-2 border-gray-600 border-t-transparent rounded-full animate-spin"></div>
                          <span>Resetting...</span>
                        </div>
                      ) : (
                        <span>Reset to Defaults</span>
                      )}
                    </Button>
                  </div>
                </form>
              </Form>
            </CardContent>
          </Card>

          <Separator />

          {/* System Control Section */}
          <Card className="border-0 shadow-lg bg-white/70 dark:bg-gray-800/70 backdrop-blur-sm">
            <CardHeader>
              <div className="flex items-center space-x-3">
                <div className="p-2 bg-purple-100 dark:bg-purple-900 rounded-lg">
                  <Zap className="w-5 h-5 text-purple-600 dark:text-purple-400" />
                </div>
                <div>
                  <CardTitle className="text-xl text-gray-900 dark:text-white">
                    System Control
                  </CardTitle>
                  <CardDescription className="text-gray-600 dark:text-gray-400">
                    Monitor and control the tennis court scraping system
                  </CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent className="space-y-6">
              
              {/* System Status Display */}
              <SystemStatusDisplay 
                status={systemControl.systemStatus} 
                lastUpdate={systemControl.lastUpdate || undefined}
              />
              
              {/* Control Buttons */}
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <Button
                  variant="outline"
                  onClick={handlePauseScraping}
                  className="flex items-center space-x-2 hover:scale-105 transition-transform duration-200 h-12"
                  disabled={systemControl.systemStatus === 'PAUSED' || isSystemControlLoading}
                >
                  <Pause className="w-4 h-4" />
                  <span>Pause Monitoring</span>
                </Button>
                
                <Button
                  variant="default"
                  onClick={handleResumeScraping}
                  className="flex items-center space-x-2 hover:scale-105 transition-transform duration-200 h-12"
                  disabled={systemControl.systemStatus === 'RUNNING' || isSystemControlLoading}
                >
                  <Play className="w-4 h-4" />
                  <span>Resume Monitoring</span>
                </Button>
                
                <Button
                  variant="secondary"
                  onClick={handleRestartSystem}
                  className="flex items-center space-x-2 hover:scale-105 transition-transform duration-200 h-12"
                  disabled={isSystemControlLoading}
                >
                  <RotateCcw className="w-4 h-4" />
                  <span>Restart System</span>
                </Button>
              </div>
              
              {/* System Information */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pt-4 border-t border-gray-200 dark:border-gray-700">
                <div className="space-y-2">
                  <h4 className="font-medium text-gray-900 dark:text-white">System Information</h4>
                  <div className="text-sm text-gray-600 dark:text-gray-400 space-y-1">
                    <p>• Monitoring {systemControl.systemInfo.monitoredClubs} tennis clubs</p>
                    <p>• Last scan: {formatTimeAgo(systemControl.systemInfo.lastScan)}</p>
                    <p>• Next scan: {formatTimeUntil(systemControl.systemInfo.nextScan)}</p>
                  </div>
                </div>
                <div className="space-y-2">
                  <h4 className="font-medium text-gray-900 dark:text-white">Performance</h4>
                  <div className="text-sm text-gray-600 dark:text-gray-400 space-y-1">
                    <p>• Average response time: {systemControl.systemInfo.averageResponseTime}s</p>
                    <p>• Success rate: {systemControl.systemInfo.successRate}%</p>
                    <p>• Courts found today: {systemControl.systemInfo.courtsFoundToday}</p>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* System Health */}
          {systemHealthQuery.data && (
            <div>
              <h3 className="text-lg font-medium mb-4">System Health</h3>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <p className="text-sm font-medium">CPU Usage</p>
                    <Badge variant={systemHealthQuery.data.cpuUsage > 80 ? 'destructive' : 'secondary'}>
                      {systemHealthQuery.data.cpuUsage}%
                    </Badge>
                  </div>
                </div>
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <p className="text-sm font-medium">Memory Usage</p>
                    <Badge variant={systemHealthQuery.data.memoryUsage > 80 ? 'destructive' : 'secondary'}>
                      {systemHealthQuery.data.memoryUsage}%
                    </Badge>
                  </div>
                </div>
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <p className="text-sm font-medium">Active Connections</p>
                    <Badge variant="secondary">
                      {systemHealthQuery.data.activeConnections}
                    </Badge>
                  </div>
                </div>
              </div>
            </div>
          )}

        </div>
      </main>
    </div>
  )
}

export default Settings 