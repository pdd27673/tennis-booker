import React, { useState, useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { useNavigate } from 'react-router-dom'
import { useAppStore } from '@/stores/appStore'
import { userApi } from '@/services/userApi'
import { 
  ArrowLeft,
  Settings2, 
  Clock, 
  Save,
  Bell,
  MapPin,
  DollarSign,
  Loader2
} from 'lucide-react'

// Form validation schema matching backend UserPreferences
const userPreferencesSchema = z.object({
  preferredVenues: z.array(z.string()).min(1, 'Select at least one venue'),
  excludedVenues: z.array(z.string()),
  times: z.array(z.object({
    start: z.string().regex(/^\d{2}:\d{2}$/, 'Time must be in HH:MM format'),
    end: z.string().regex(/^\d{2}:\d{2}$/, 'Time must be in HH:MM format')
  })).optional(), // Legacy field for backward compatibility
  weekdayTimes: z.array(z.object({
    start: z.string().regex(/^\d{2}:\d{2}$/, 'Time must be in HH:MM format'),
    end: z.string().regex(/^\d{2}:\d{2}$/, 'Time must be in HH:MM format')
  })).min(1, 'Add at least one weekday time slot'),
  weekendTimes: z.array(z.object({
    start: z.string().regex(/^\d{2}:\d{2}$/, 'Time must be in HH:MM format'),
    end: z.string().regex(/^\d{2}:\d{2}$/, 'Time must be in HH:MM format')
  })).min(1, 'Add at least one weekend time slot'),
  preferredDays: z.array(z.string()).min(1, 'Select at least one day'),
  maxPrice: z.number().min(1, 'Price must be at least $1').max(1000, 'Price cannot exceed $1000'),
  notificationSettings: z.object({
    email: z.boolean(),
    emailAddress: z.string().email('Please enter a valid email address').optional().or(z.literal('')),
    instantAlerts: z.boolean(),
    maxAlertsPerHour: z.number().min(1).max(100),
    maxAlertsPerDay: z.number().min(1).max(1000),
    alertTimeWindowStart: z.string().regex(/^\d{2}:\d{2}$/, 'Time must be in HH:MM format'),
    alertTimeWindowEnd: z.string().regex(/^\d{2}:\d{2}$/, 'Time must be in HH:MM format'),
    unsubscribed: z.boolean(),
  }),
})

type UserPreferencesForm = z.infer<typeof userPreferencesSchema>

const Settings: React.FC = () => {
  const navigate = useNavigate()
  const { userProfile: user, clearAuthState, addNotification, setUserPreferences } = useAppStore()
  
  const [isLoading, setIsLoading] = useState(false)
  const [isSaving, setIsSaving] = useState(false)
  const [availableVenues, setAvailableVenues] = useState<string[]>([])

  const form = useForm<UserPreferencesForm>({
    resolver: zodResolver(userPreferencesSchema),
    defaultValues: {
      preferredVenues: [],
      excludedVenues: [],
      weekdayTimes: [{ start: '18:00', end: '20:00' }],
      weekendTimes: [{ start: '09:00', end: '11:00' }],
      preferredDays: ['monday', 'tuesday', 'wednesday', 'thursday', 'friday'],
      maxPrice: 100,
      notificationSettings: {
        email: true,
        emailAddress: user?.email || '',
        instantAlerts: true,
        maxAlertsPerHour: 10,
        maxAlertsPerDay: 50,
        alertTimeWindowStart: '07:00',
        alertTimeWindowEnd: '22:00',
        unsubscribed: false,
      },
    },
  })

  // Load user preferences and available venues
  useEffect(() => {
    const loadData = async () => {
      setIsLoading(true)
      try {
        console.log('🔄 Settings: Loading user preferences...')
        
        // Load user preferences with detailed logging
        const prefs = await userApi.getUserPreferences()
        console.log('📋 Settings: Loaded user preferences:', prefs)
        setUserPreferences(prefs)
        
        // Load real venues from API
        console.log('🏟️ Settings: Loading available venues from API...')
        try {
          const { courtApi } = await import('@/services/courtApi')
          const venues = await courtApi.getVenues()
          console.log('🏟️ Settings: Loaded venues from API:', venues)
          const venueNames = venues.map(venue => venue.name)
          setAvailableVenues(venueNames)
          console.log('✅ Settings: Available venues set:', venueNames)
        } catch (venueError) {
          console.warn('⚠️ Settings: Failed to load venues from API, using fallback:', venueError)
          // Fallback to hardcoded venues if API fails
          const fallbackVenues = [
            'Victoria Park Tennis Centre',
            'Stratford Park Tennis Club', 
            'Ropemakers Field Tennis Courts',
            'Tennis Club Central',
            'Riverside Tennis Club',
            'Elite Tennis Academy',
            'Community Recreation Center'
          ]
          setAvailableVenues(fallbackVenues)
          console.log('🔄 Settings: Using fallback venues:', fallbackVenues)
        }
        
        // Update form with loaded preferences
        const formData = {
          preferredVenues: prefs.preferredVenues || [],
          excludedVenues: prefs.excludedVenues || [],
          weekdayTimes: prefs.weekdayTimes || prefs.times || [{ start: '18:00', end: '20:00' }],
          weekendTimes: prefs.weekendTimes || prefs.times || [{ start: '09:00', end: '11:00' }],
          preferredDays: prefs.preferredDays || ['monday', 'tuesday', 'wednesday', 'thursday', 'friday'],
          maxPrice: prefs.maxPrice || 100,
          notificationSettings: {
            email: prefs.notificationSettings?.email ?? true,
            emailAddress: prefs.notificationSettings?.emailAddress || user?.email || '',
            instantAlerts: prefs.notificationSettings?.instantAlerts ?? true,
            maxAlertsPerHour: prefs.notificationSettings?.maxAlertsPerHour || 10,
            maxAlertsPerDay: prefs.notificationSettings?.maxAlertsPerDay || 50,
            alertTimeWindowStart: prefs.notificationSettings?.alertTimeWindowStart || '07:00',
            alertTimeWindowEnd: prefs.notificationSettings?.alertTimeWindowEnd || '22:00',
            unsubscribed: prefs.notificationSettings?.unsubscribed ?? false,
          },
        }
        console.log('📝 Settings: Form data to be set:', formData)
        form.reset(formData)
        
      } catch (error) {
        console.error('❌ Settings: Failed to load settings:', error)
        addNotification({
          title: 'Error Loading Settings',
          message: 'Failed to load your preferences. Using defaults.',
          type: 'error',
        })
      } finally {
        setIsLoading(false)
      }
    }

    loadData()
  }, [form, setUserPreferences, addNotification, user?.email])

  const onSubmit = async (data: UserPreferencesForm) => {
    setIsSaving(true)
    try {
      console.log('💾 Settings: Saving form data:', data)
      
      const savedPreferences = await userApi.updateUserPreferences(data)
      console.log('✅ Settings: Successfully saved preferences:', savedPreferences)
      
      setUserPreferences(savedPreferences)
      addNotification({
        title: 'Settings Saved',
        message: 'Your preferences have been updated successfully.',
        type: 'success',
      })
    } catch (error) {
      console.error('❌ Settings: Failed to save preferences:', error)
      addNotification({
        title: 'Save Failed',
        message: error instanceof Error ? error.message : 'Failed to save your preferences.',
        type: 'error',
      })
    } finally {
      setIsSaving(false)
    }
  }

  const handleLogout = () => {
    clearAuthState()
    addNotification({
      title: 'Logged Out',
      message: 'You have been successfully logged out',
      type: 'info',
    })
    navigate('/login')
  }

  const addWeekdayTimeSlot = () => {
    const currentTimes = form.getValues('weekdayTimes') || []
    form.setValue('weekdayTimes', [...currentTimes, { start: '18:00', end: '20:00' }])
  }

  const removeWeekdayTimeSlot = (index: number) => {
    const currentTimes = form.getValues('weekdayTimes') || []
    if (currentTimes.length > 1) {
      form.setValue('weekdayTimes', currentTimes.filter((_, i) => i !== index))
    }
  }

  const addWeekendTimeSlot = () => {
    const currentTimes = form.getValues('weekendTimes') || []
    form.setValue('weekendTimes', [...currentTimes, { start: '09:00', end: '11:00' }])
  }

  const removeWeekendTimeSlot = (index: number) => {
    const currentTimes = form.getValues('weekendTimes') || []
    if (currentTimes.length > 1) {
      form.setValue('weekendTimes', currentTimes.filter((_, i) => i !== index))
    }
  }

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="w-8 h-8 animate-spin mx-auto mb-4" />
          <p className="text-gray-600">Loading settings...</p>
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
                <Settings2 className="w-8 h-8 text-blue-600 dark:text-blue-400" />
                <div>
                  <h1 className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white">
                    Settings
                  </h1>
                  <p className="text-gray-600 dark:text-gray-400">
                    Manage your tennis court monitoring preferences
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
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
            
            {/* Preferred Venues */}
            <Card className="border-0 shadow-lg bg-white/70 dark:bg-gray-800/70 backdrop-blur-sm">
              <CardHeader>
                <div className="flex items-center space-x-3">
                  <div className="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
                    <MapPin className="w-5 h-5 text-blue-600 dark:text-blue-400" />
                  </div>
                  <div>
                    <CardTitle className="text-xl text-gray-900 dark:text-white">
                      Venue Preferences
                    </CardTitle>
                    <CardDescription className="text-gray-600 dark:text-gray-400">
                      Select which tennis venues you'd like to monitor
                    </CardDescription>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="space-y-6">
                <FormField
                  control={form.control}
                  name="preferredVenues"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Preferred Venues</FormLabel>
                      <FormControl>
                        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                          {availableVenues.map((venue) => (
                            <div key={venue} className="flex items-center space-x-2">
                              <Checkbox
                                id={venue}
                                checked={field.value?.includes(venue) || false}
                                onCheckedChange={(checked) => {
                                  if (checked) {
                                    field.onChange([...(field.value || []), venue])
                                  } else {
                                    field.onChange((field.value || []).filter((v) => v !== venue))
                                  }
                                }}
                              />
                              <label htmlFor={venue} className="text-sm font-medium cursor-pointer">
                                {venue}
                              </label>
                            </div>
                          ))}
                        </div>
                      </FormControl>
                      <FormDescription>
                        Select the venues you want to monitor for available courts
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </CardContent>
            </Card>

            {/* Add New Venue/Court Section */}
            <Card className="border-0 shadow-lg bg-white/70 dark:bg-gray-800/70 backdrop-blur-sm">
              <CardHeader>
                <div className="flex items-center space-x-3">
                  <div className="p-2 bg-purple-100 dark:bg-purple-900 rounded-lg">
                    <MapPin className="w-5 h-5 text-purple-600 dark:text-purple-400" />
                  </div>
                  <div>
                    <CardTitle className="text-xl text-gray-900 dark:text-white">
                      Add New Venue to Scraping List
                    </CardTitle>
                    <CardDescription className="text-gray-600 dark:text-gray-400">
                      Request a new tennis venue to be added to our monitoring system
                    </CardDescription>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="space-y-6">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <label className="text-sm font-medium">Venue Name</label>
                    <Input placeholder="e.g., City Tennis Club" />
                  </div>
                  <div className="space-y-2">
                    <label className="text-sm font-medium">Venue Website URL</label>
                    <Input placeholder="https://example.com/book-courts" />
                  </div>
                  <div className="space-y-2">
                    <label className="text-sm font-medium">Booking Provider</label>
                    <select className="w-full p-2 border rounded-md">
                      <option value="">Select Provider</option>
                      <option value="clubspark">ClubSpark</option>
                      <option value="courtside">Courtside</option>
                      <option value="other">Other/Unknown</option>
                    </select>
                  </div>
                  <div className="space-y-2">
                    <label className="text-sm font-medium">Location (City)</label>
                    <Input placeholder="e.g., London, Birmingham" />
                  </div>
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">Additional Notes (Optional)</label>
                  <textarea 
                    className="w-full p-2 border rounded-md resize-none"
                    rows={3}
                    placeholder="Any special instructions or details about this venue..."
                  />
                </div>
                <Button 
                  type="button" 
                  variant="outline"
                  onClick={() => {
                    addNotification({
                      title: 'Venue Request Submitted',
                      message: 'Your venue request has been submitted for review. We\'ll add it to our scraping system within 24-48 hours.',
                      type: 'success',
                    })
                  }}
                >
                  Submit Venue Request
                </Button>
              </CardContent>
            </Card>

            {/* Time Preferences */}
            <Card className="border-0 shadow-lg bg-white/70 dark:bg-gray-800/70 backdrop-blur-sm">
              <CardHeader>
                <div className="flex items-center space-x-3">
                  <div className="p-2 bg-green-100 dark:bg-green-900 rounded-lg">
                    <Clock className="w-5 h-5 text-green-600 dark:text-green-400" />
                  </div>
                  <div>
                    <CardTitle className="text-xl text-gray-900 dark:text-white">
                      Time Preferences
                    </CardTitle>
                    <CardDescription className="text-gray-600 dark:text-gray-400">
                      Set your preferred playing times and days
                    </CardDescription>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="space-y-6">
                {/* Weekday Time Slots */}
                <FormField
                  control={form.control}
                  name="weekdayTimes"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Weekday Time Slots (Monday - Friday)</FormLabel>
                      <FormControl>
                        <div className="space-y-3">
                          {field.value?.map((timeSlot, index) => (
                            <div key={index} className="flex items-center space-x-3">
                              <Input
                                type="time"
                                value={timeSlot.start || ''}
                                onChange={(e) => {
                                  const newTimes = [...(field.value || [])]
                                  newTimes[index] = { 
                                    start: e.target.value,
                                    end: newTimes[index]?.end || '20:00'
                                  }
                                  field.onChange(newTimes)
                                }}
                                className="w-32"
                              />
                              <span className="text-gray-500">to</span>
                              <Input
                                type="time"
                                value={timeSlot.end || ''}
                                onChange={(e) => {
                                  const newTimes = [...(field.value || [])]
                                  newTimes[index] = { 
                                    start: newTimes[index]?.start || '18:00',
                                    end: e.target.value
                                  }
                                  field.onChange(newTimes)
                                }}
                                className="w-32"
                              />
                              {(field.value?.length || 0) > 1 && (
                                <Button
                                  type="button"
                                  variant="outline"
                                  size="sm"
                                  onClick={() => removeWeekdayTimeSlot(index)}
                                >
                                  Remove
                                </Button>
                              )}
                            </div>
                          ))}
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={addWeekdayTimeSlot}
                          >
                            Add Weekday Time Slot
                          </Button>
                        </div>
                      </FormControl>
                      <FormDescription>
                        Add your preferred playing time slots for weekdays
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                {/* Weekend Time Slots */}
                <FormField
                  control={form.control}
                  name="weekendTimes"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Weekend Time Slots (Saturday - Sunday)</FormLabel>
                      <FormControl>
                        <div className="space-y-3">
                          {field.value?.map((timeSlot, index) => (
                            <div key={index} className="flex items-center space-x-3">
                              <Input
                                type="time"
                                value={timeSlot.start || ''}
                                onChange={(e) => {
                                  const newTimes = [...(field.value || [])]
                                  newTimes[index] = { 
                                    start: e.target.value,
                                    end: newTimes[index]?.end || '11:00'
                                  }
                                  field.onChange(newTimes)
                                }}
                                className="w-32"
                              />
                              <span className="text-gray-500">to</span>
                              <Input
                                type="time"
                                value={timeSlot.end || ''}
                                onChange={(e) => {
                                  const newTimes = [...(field.value || [])]
                                  newTimes[index] = { 
                                    start: newTimes[index]?.start || '09:00',
                                    end: e.target.value
                                  }
                                  field.onChange(newTimes)
                                }}
                                className="w-32"
                              />
                              {(field.value?.length || 0) > 1 && (
                                <Button
                                  type="button"
                                  variant="outline"
                                  size="sm"
                                  onClick={() => removeWeekendTimeSlot(index)}
                                >
                                  Remove
                                </Button>
                              )}
                            </div>
                          ))}
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={addWeekendTimeSlot}
                          >
                            Add Weekend Time Slot
                          </Button>
                        </div>
                      </FormControl>
                      <FormDescription>
                        Add your preferred playing time slots for weekends
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                {/* <FormField
                  control={form.control}
                  name="preferredDays"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Preferred Days</FormLabel>
                      <FormControl>
                        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                          {daysOfWeek.map((day) => (
                            <div key={day.value} className="flex items-center space-x-2">
                              <Checkbox
                                id={day.value}
                                checked={field.value?.includes(day.value) || false}
                                onCheckedChange={(checked) => {
                                  if (checked) {
                                    field.onChange([...(field.value || []), day.value])
                                  } else {
                                    field.onChange((field.value || []).filter((d) => d !== day.value))
                                  }
                                }}
                              />
                              <label htmlFor={day.value} className="text-sm font-medium cursor-pointer">
                                {day.label}
                              </label>
                            </div>
                          ))}
                        </div>
                      </FormControl>
                      <FormDescription>
                        Select the days you prefer to play
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                /> */}
              </CardContent>
            </Card>

            {/* Price Preferences */}
            <Card className="border-0 shadow-lg bg-white/70 dark:bg-gray-800/70 backdrop-blur-sm">
              <CardHeader>
                <div className="flex items-center space-x-3">
                  <div className="p-2 bg-yellow-100 dark:bg-yellow-900 rounded-lg">
                    <DollarSign className="w-5 h-5 text-yellow-600 dark:text-yellow-400" />
                  </div>
                  <div>
                    <CardTitle className="text-xl text-gray-900 dark:text-white">
                      Price Preferences
                    </CardTitle>
                    <CardDescription className="text-gray-600 dark:text-gray-400">
                      Set your maximum budget for court bookings
                    </CardDescription>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <FormField
                  control={form.control}
                  name="maxPrice"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Maximum Price per Hour ($)</FormLabel>
                      <FormControl>
                        <Input
                          type="number"
                          min="1"
                          max="1000"
                          {...field}
                          onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
                          className="w-32"
                        />
                      </FormControl>
                      <FormDescription>
                        Only show courts within your budget
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </CardContent>
            </Card>

            {/* Notification Settings */}
            <Card className="border-0 shadow-lg bg-white/70 dark:bg-gray-800/70 backdrop-blur-sm">
              <CardHeader>
                <div className="flex items-center space-x-3">
                  <div className="p-2 bg-purple-100 dark:bg-purple-900 rounded-lg">
                    <Bell className="w-5 h-5 text-purple-600 dark:text-purple-400" />
                  </div>
                  <div>
                    <CardTitle className="text-xl text-gray-900 dark:text-white">
                      Notification Settings
                    </CardTitle>
                    <CardDescription className="text-gray-600 dark:text-gray-400">
                      Configure how you want to be notified about available courts
                    </CardDescription>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="space-y-6">
                <FormField
                  control={form.control}
                  name="notificationSettings.email"
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
                  name="notificationSettings.emailAddress"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Email Address</FormLabel>
                      <FormControl>
                        <Input
                          type="email"
                          placeholder="your.email@example.com"
                          {...field}
                        />
                      </FormControl>
                      <FormDescription>
                        Email address for notifications
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="notificationSettings.instantAlerts"
                  render={({ field }) => (
                    <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                      <div className="space-y-0.5">
                        <FormLabel className="text-base">Instant Alerts</FormLabel>
                        <FormDescription>
                          Get immediate notifications for urgent availability
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

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <FormField
                    control={form.control}
                    name="notificationSettings.maxAlertsPerHour"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Max Alerts per Hour</FormLabel>
                        <FormControl>
                          <Input
                            type="number"
                            min="1"
                            max="100"
                            {...field}
                            onChange={(e) => field.onChange(parseInt(e.target.value) || 1)}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="notificationSettings.maxAlertsPerDay"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Max Alerts per Day</FormLabel>
                        <FormControl>
                          <Input
                            type="number"
                            min="1"
                            max="1000"
                            {...field}
                            onChange={(e) => field.onChange(parseInt(e.target.value) || 1)}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <FormField
                    control={form.control}
                    name="notificationSettings.alertTimeWindowStart"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Alert Window Start</FormLabel>
                        <FormControl>
                          <Input
                            type="time"
                            {...field}
                          />
                        </FormControl>
                        <FormDescription>
                          Start time for receiving alerts
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="notificationSettings.alertTimeWindowEnd"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Alert Window End</FormLabel>
                        <FormControl>
                          <Input
                            type="time"
                            {...field}
                          />
                        </FormControl>
                        <FormDescription>
                          End time for receiving alerts
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>
              </CardContent>
            </Card>

            {/* Save Button */}
            <div className="flex justify-end space-x-4">
              <Button
                type="submit"
                disabled={isSaving}
                className="hover:scale-105 transition-transform duration-200 min-w-[120px]"
              >
                {isSaving ? (
                  <div className="flex items-center space-x-2">
                    <Loader2 className="w-4 h-4 animate-spin" />
                    <span>Saving...</span>
                  </div>
                ) : (
                  <div className="flex items-center space-x-2">
                    <Save className="w-4 h-4" />
                    <span>Save Preferences</span>
                  </div>
                )}
              </Button>
            </div>

          </form>
        </Form>
      </main>
    </div>
  )
}

export default Settings 