import React, { useEffect, useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { useNavigate } from 'react-router-dom'
import { userApi } from '@/services/userApi'
import { useAppStore, type UserPreferences } from '@/stores/appStore'
import { 
  Settings, 
  Clock, 
  MapPin, 
  DollarSign, 
  Bell,
  Calendar,
  Loader2,
  AlertCircle
} from 'lucide-react'

export default function UserPreferencesCard() {
  const navigate = useNavigate()
  const { userPreferences, setUserPreferences, addNotification } = useAppStore()
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const loadPreferences = async () => {
      try {
        setIsLoading(true)
        setError(null)
        console.log('🔄 UserPreferencesCard: Loading user preferences...')
        
        const prefs = await userApi.getUserPreferences()
        console.log('✅ UserPreferencesCard: Loaded preferences:', prefs)
        setUserPreferences(prefs)
      } catch (error) {
        console.error('❌ UserPreferencesCard: Failed to load preferences:', error)
        setError('Failed to load preferences')
        addNotification({
          title: 'Error Loading Preferences',
          message: 'Failed to load your preferences. Please try again.',
          type: 'error',
        })
      } finally {
        setIsLoading(false)
      }
    }

    loadPreferences()
  }, [setUserPreferences, addNotification])

  const formatTimeSlots = (timeSlots?: Array<{ start: string; end: string }>) => {
    if (!timeSlots || timeSlots.length === 0) return 'None set'
    return timeSlots.map(slot => `${slot.start} - ${slot.end}`).join(', ')
  }

  const formatDays = (days?: string[]) => {
    if (!days || days.length === 0) return 'None selected'
    const dayNames: Record<string, string> = {
      monday: 'Mon',
      tuesday: 'Tue', 
      wednesday: 'Wed',
      thursday: 'Thu',
      friday: 'Fri',
      saturday: 'Sat',
      sunday: 'Sun'
    }
    return days.map(day => dayNames[day] || day).join(', ')
  }

  const formatVenues = (venues?: string[]) => {
    if (!venues || venues.length === 0) return 'None selected'
    return venues.join(', ')
  }

  if (isLoading) {
    return (
      <Card className="border-0 shadow-lg bg-white/70 dark:bg-gray-800/70 backdrop-blur-sm">
        <CardHeader>
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
              <Settings className="w-5 h-5 text-blue-600 dark:text-blue-400" />
            </div>
            <div>
              <CardTitle className="text-xl text-gray-900 dark:text-white">
                Your Preferences
              </CardTitle>
              <CardDescription className="text-gray-600 dark:text-gray-400">
                Current notification and booking preferences
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-3">
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-3/4" />
            <Skeleton className="h-4 w-1/2" />
          </div>
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <Card className="border-0 shadow-lg bg-white/70 dark:bg-gray-800/70 backdrop-blur-sm">
        <CardHeader>
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-red-100 dark:bg-red-900 rounded-lg">
              <AlertCircle className="w-5 h-5 text-red-600 dark:text-red-400" />
            </div>
            <div>
              <CardTitle className="text-xl text-gray-900 dark:text-white">
                Preferences Error
              </CardTitle>
              <CardDescription className="text-gray-600 dark:text-gray-400">
                {error}
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <Button 
            onClick={() => navigate('/settings')}
            variant="outline"
            className="w-full"
          >
            Go to Settings
          </Button>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className="border-0 shadow-lg bg-white/70 dark:bg-gray-800/70 backdrop-blur-sm">
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
              <Settings className="w-5 h-5 text-blue-600 dark:text-blue-400" />
            </div>
            <div>
              <CardTitle className="text-xl text-gray-900 dark:text-white">
                Your Preferences
              </CardTitle>
              <CardDescription className="text-gray-600 dark:text-gray-400">
                Current notification and booking preferences
              </CardDescription>
            </div>
          </div>
          <Button 
            onClick={() => navigate('/settings')}
            variant="outline"
            size="sm"
            className="hover:scale-105 transition-transform duration-200"
          >
            <Settings className="w-4 h-4 mr-2" />
            Edit
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Time Preferences */}
        <div className="space-y-3">
          <div className="flex items-center space-x-2">
            <Clock className="w-4 h-4 text-green-600 dark:text-green-400" />
            <h4 className="font-medium text-gray-900 dark:text-white">Time Preferences</h4>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 ml-6">
            <div>
              <p className="text-sm font-medium text-gray-700 dark:text-gray-300">Weekdays</p>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                {formatTimeSlots(userPreferences?.weekdayTimes || userPreferences?.times)}
              </p>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-700 dark:text-gray-300">Weekends</p>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                {formatTimeSlots(userPreferences?.weekendTimes || userPreferences?.times)}
              </p>
            </div>
          </div>
        </div>

        {/* Venue Preferences */}
        <div className="space-y-3">
          <div className="flex items-center space-x-2">
            <MapPin className="w-4 h-4 text-purple-600 dark:text-purple-400" />
            <h4 className="font-medium text-gray-900 dark:text-white">Preferred Venues</h4>
          </div>
          <div className="ml-6">
            <p className="text-sm text-gray-600 dark:text-gray-400">
              {formatVenues(userPreferences?.preferredVenues)}
            </p>
          </div>
        </div>

        {/* Days Preferences */}
        <div className="space-y-3">
          <div className="flex items-center space-x-2">
            <Calendar className="w-4 h-4 text-orange-600 dark:text-orange-400" />
            <h4 className="font-medium text-gray-900 dark:text-white">Preferred Days</h4>
          </div>
          <div className="ml-6">
            <p className="text-sm text-gray-600 dark:text-gray-400">
              {formatDays(userPreferences?.preferredDays)}
            </p>
          </div>
        </div>

        {/* Price Preference */}
        <div className="space-y-3">
          <div className="flex items-center space-x-2">
            <DollarSign className="w-4 h-4 text-yellow-600 dark:text-yellow-400" />
            <h4 className="font-medium text-gray-900 dark:text-white">Max Price</h4>
          </div>
          <div className="ml-6">
            <Badge variant="secondary" className="bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-100">
              ${userPreferences?.maxPrice || 0}/hour
            </Badge>
          </div>
        </div>

        {/* Notification Settings */}
        <div className="space-y-3">
          <div className="flex items-center space-x-2">
            <Bell className="w-4 h-4 text-red-600 dark:text-red-400" />
            <h4 className="font-medium text-gray-900 dark:text-white">Notifications</h4>
          </div>
          <div className="ml-6 space-y-2">
            <div className="flex items-center space-x-2">
              <Badge 
                variant={userPreferences?.notificationSettings?.email ? "default" : "secondary"}
                className={userPreferences?.notificationSettings?.email 
                  ? "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100" 
                  : "bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-100"
                }
              >
                Email: {userPreferences?.notificationSettings?.email ? 'Enabled' : 'Disabled'}
              </Badge>
            </div>
            {userPreferences?.notificationSettings?.email && (
              <p className="text-xs text-gray-500 dark:text-gray-400">
                Max {userPreferences?.notificationSettings?.maxAlertsPerHour || 10}/hour, 
                {userPreferences?.notificationSettings?.maxAlertsPerDay || 50}/day
              </p>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )
} 