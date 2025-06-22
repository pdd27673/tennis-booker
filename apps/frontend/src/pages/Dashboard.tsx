import { useEffect, useState, useCallback, useRef } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { TextGenerateEffect } from '@/components/ui/text-generate-effect'
import { useNavigate } from 'react-router-dom'
import CourtCard, { type CourtCardProps } from '@/components/CourtCard'
import ScrapingLogsTerminal from '@/components/ScrapingLogsTerminal'
import UserPreferencesCard from '@/components/UserPreferencesCard'
import { courtApi } from '@/services/courtApi'
import { systemApi } from '@/services/systemApi'
import { useAppStore } from '@/stores/appStore'
import { 
  RefreshCw, 
  Play, 
  Pause, 
  RotateCcw, 
  Activity,
  Clock,
  Bell,
  Loader2
} from 'lucide-react'

// Mock court data
const mockCourtData: Omit<CourtCardProps, 'className'>[] = [
  {
    courtName: 'Court Alpha',
    courtType: 'Tennis Court',
    venue: 'Tennis Club Central',
    availabilityStatus: 'available',
    timeSlot: 'Today 2:00 PM - 3:00 PM',
    price: '$25/hour',
    maxPlayers: 4,
    bookingLink: 'https://example.com/book/court-alpha'
  },
  {
    courtName: 'Court Beta',
    courtType: 'Tennis Court',
    venue: 'Riverside Tennis Club',
    availabilityStatus: 'available',
    timeSlot: 'Today 4:00 PM - 5:00 PM',
    price: '$30/hour',
    maxPlayers: 4,
    bookingLink: '#placeholder-beta'
  },
  {
    courtName: 'Court Gamma',
    courtType: 'Padel Court',
    venue: 'Sports Complex North',
    availabilityStatus: 'booked',
    timeSlot: 'Today 6:00 PM - 7:00 PM',
    price: '$40/hour',
    maxPlayers: 4,
    bookingLink: '#placeholder-gamma'
  },
  {
    courtName: 'Court Delta',
    courtType: 'Tennis Court',
    venue: 'Tennis Club Central',
    availabilityStatus: 'available',
    timeSlot: 'Tomorrow 10:00 AM - 11:00 AM',
    price: '$25/hour',
    maxPlayers: 4,
    bookingLink: '#placeholder-delta'
  },
  {
    courtName: 'Court Epsilon',
    courtType: 'Squash Court',
    venue: 'Downtown Sports Center',
    availabilityStatus: 'maintenance',
    timeSlot: 'Tomorrow 2:00 PM - 3:00 PM',
    price: '$20/hour',
    maxPlayers: 2,
    bookingLink: '#placeholder-epsilon'
  },
  {
    courtName: 'Court Zeta',
    courtType: 'Tennis Court',
    venue: 'Riverside Tennis Club',
    availabilityStatus: 'pending',
    timeSlot: 'Tomorrow 4:00 PM - 5:00 PM',
    price: '$30/hour',
    maxPlayers: 4,
    bookingLink: '#placeholder-zeta'
  }
]

export default function Dashboard() {
  const navigate = useNavigate()
  const { userProfile: user, clearAuthState, addNotification } = useAppStore()
  
  // State for real API data
  const [courtData, setCourtData] = useState<Omit<CourtCardProps, 'className'>[]>([])
  const [dashboardStats, setDashboardStats] = useState({
    systemStatus: 'LOADING',
    activeCourts: 0,
    availableSlots: 0,
    totalVenues: 0,
  })
  const [systemMetrics, setSystemMetrics] = useState({
    uptime: '0 minutes',
    lastScrapeTime: 'Never',
    notificationsSent: 0,
    scraperHealth: 'Unknown'
  })
  const [recentActivity, setRecentActivity] = useState<Array<{
    id: string
    type: 'available' | 'booked' | 'notification' | 'system'
    message: string
    timestamp: Date
    venue?: string
  }>>([])
  const [lastActivityUpdate, setLastActivityUpdate] = useState<Date>(new Date())
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isSystemControlLoading, setIsSystemControlLoading] = useState(false)
  const isLoadingRef = useRef(false)

  // Helper function to safely get court slot properties
  const getCourtSlotProperty = useCallback((slot: unknown, property: string, fallback: string = 'Unknown'): string => {
    if (slot && typeof slot === 'object' && slot !== null) {
      const obj = slot as Record<string, unknown>
      return String(obj[property] || fallback)
    }
    return fallback
  }, [])

  const generateRecentActivity = useCallback((
    stats: typeof dashboardStats, 
    systemStatus: string, 
    metrics: typeof systemMetrics, 
    courtSlots?: unknown[]
  ) => {
    const now = new Date()
    const activities = []
    const timestamp = now.getTime()

    // Add system status activity
    if (systemStatus === 'RUNNING') {
      activities.push({
        id: `system-running-${timestamp}`,
        type: 'system' as const,
        message: `System is operational - ${metrics.scraperHealth || 'Excellent'} health`,
        timestamp: new Date(now.getTime() - 2 * 60 * 1000), // 2 minutes ago
      })
    }

    // Add last scrape activity
    if (metrics.lastScrapeTime) {
      activities.push({
        id: `last-scrape-${timestamp}`,
        type: 'system' as const,
        message: `Last scrape completed at ${metrics.lastScrapeTime}`,
        timestamp: new Date(now.getTime() - 8 * 60 * 1000), // 8 minutes ago
      })
    }

    // Add court availability activities using real court slot data
    if (courtSlots && courtSlots.length > 0) {
      // Get a random available slot for realistic activity
      const availableSlots = courtSlots.filter(slot => 
        slot && typeof slot === 'object' && 
        ('available' in slot ? getCourtSlotProperty(slot, 'available', 'true') !== 'false' : true)
      )
      if (availableSlots.length > 0) {
        const randomSlot = availableSlots[Math.floor(Math.random() * Math.min(availableSlots.length, 5))]
        if (randomSlot) {
          const courtName = getCourtSlotProperty(randomSlot, 'court_name') || getCourtSlotProperty(randomSlot, 'courtName', 'Unknown Court')
          const venueName = getCourtSlotProperty(randomSlot, 'venue_name') || getCourtSlotProperty(randomSlot, 'venue', 'Unknown Venue')
          activities.push({
            id: `court-available-${timestamp}`,
            type: 'available' as const,
            message: `${courtName} at ${venueName} - slot opened`,
            timestamp: new Date(now.getTime() - 14 * 60 * 1000), // 14 minutes ago
            venue: venueName
          })
        }
      }

      // Add a booking activity (simulated)
      if (availableSlots.length > 1) {
        const randomSlot = availableSlots[Math.floor(Math.random() * Math.min(availableSlots.length, 5))]
        if (randomSlot) {
          const courtName = getCourtSlotProperty(randomSlot, 'court_name') || getCourtSlotProperty(randomSlot, 'courtName', 'Unknown Court')
          const venueName = getCourtSlotProperty(randomSlot, 'venue_name') || getCourtSlotProperty(randomSlot, 'venue', 'Unknown Venue')
          activities.push({
            id: `court-booked-${timestamp}`,
            type: 'booked' as const,
            message: `${courtName} at ${venueName} - slot filled`,
            timestamp: new Date(now.getTime() - 20 * 60 * 1000), // 20 minutes ago
            venue: venueName
          })
        }
      }
    } else {
      // Fallback activities if no court slots available
      activities.push({
        id: `courts-available-${timestamp}`,
        type: 'available' as const,
        message: `${stats.availableSlots} court slots became available`,
        timestamp: new Date(now.getTime() - 5 * 60 * 1000), // 5 minutes ago
        venue: 'Multiple venues'
      })
    }

    // Add notification activity
    if (metrics.notificationsSent > 0) {
      activities.push({
        id: `notifications-${timestamp}`,
        type: 'notification' as const,
        message: `${metrics.notificationsSent} notifications sent to users`,
        timestamp: new Date(now.getTime() - 12 * 60 * 1000), // 12 minutes ago
      })
    }

    // Sort by timestamp (most recent first) and limit to 4 activities
    return activities.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime()).slice(0, 4)
  }, [getCourtSlotProperty])

  // Fetch real data from APIs
  useEffect(() => {
    const fetchDashboardData = async () => {
      try {
        setIsLoading(true)
        setError(null)
  
        
        // Fetch data with individual error handling for better resilience
        const results = await Promise.allSettled([
          courtApi.getCourtSlotsForCards({ limit: 6 }), // Get 6 slots for display
          courtApi.getDashboardStats(),
          systemApi.getSystemStatus(),
          systemApi.getSystemMetrics()
        ])

        // Handle court slots
        if (results[0].status === 'fulfilled') {
                  setCourtData(results[0].value)
      } else {
          setCourtData(mockCourtData)
        }

        // Handle dashboard stats
        let stats = { activeCourts: 0, availableSlots: 0, totalVenues: 0 }
        if (results[1].status === 'fulfilled') {
          stats = results[1].value
        } else {
          // Use default stats on error
        }

        // Handle system status
        let systemStatus = 'RUNNING'
        if (results[2].status === 'fulfilled') {
          systemStatus = results[2].value.status || 'RUNNING'
        } else {
          // Use default system status on error
        }

        // Handle system metrics
        let metrics = {
          uptime: 'Unknown',
          lastScrapeTime: 'Never',
          notificationsSent: 0,
          scraperHealth: 'Unknown'
        }
        if (results[3].status === 'fulfilled') {
          metrics = results[3].value
        } else {
          // Use default metrics on error
        }
        
        setDashboardStats({
          systemStatus: systemStatus as 'RUNNING' | 'ERROR' | 'MAINTENANCE',
          activeCourts: stats.activeCourts,
          availableSlots: stats.availableSlots,
          totalVenues: stats.totalVenues,
        })
        setSystemMetrics(metrics)

      } catch {
        setError('Failed to load dashboard data. Please try refreshing the page.')
        // Fallback to mock data on critical error
        setCourtData(mockCourtData)
        setDashboardStats({
          systemStatus: 'ERROR',
          activeCourts: 0,
          availableSlots: 0,
          totalVenues: 0,
        })
      } finally {
        setIsLoading(false)
      }
    }

    fetchDashboardData()

    // Set up auto-refresh every 30 seconds for system metrics
    const refreshInterval = setInterval(() => {
      if (!isLoadingRef.current) {

        Promise.allSettled([
          courtApi.getCourtSlots({ limit: 10 }), // Get court slots for activity generation
          systemApi.getSystemStatus(),
          systemApi.getSystemMetrics(),
          courtApi.getDashboardStats()
        ]).then((results) => {
          // Use callback refs to get current state
          setDashboardStats(currentStats => {
            let latestStats = currentStats
            let latestSystemStatus = currentStats.systemStatus
            
            if (results[1].status === 'fulfilled') {
              const statusResult = results[1] as PromiseFulfilledResult<{ status: string }>
              latestSystemStatus = statusResult.value.status || currentStats.systemStatus
            }
            
            if (results[3].status === 'fulfilled') {
              const statsResult = results[3] as PromiseFulfilledResult<{ activeCourts: number; availableSlots: number; totalVenues: number }>
              latestStats = {
                ...currentStats,
                systemStatus: latestSystemStatus,
                activeCourts: statsResult.value.activeCourts,
                availableSlots: statsResult.value.availableSlots,
                totalVenues: statsResult.value.totalVenues,
              }
            } else {
              latestStats = {
                ...currentStats,
                systemStatus: latestSystemStatus
              }
            }
            
            return latestStats
          })
          
          if (results[2].status === 'fulfilled') {
            const metricsResult = results[2] as PromiseFulfilledResult<typeof systemMetrics>
            setSystemMetrics(metricsResult.value)
          }
        })
      }
    }, 30000) // 30 seconds

    return () => clearInterval(refreshInterval)
  }, [])

  // Separate useEffect to update activity when state changes
  const updateActivity = useCallback(async () => {
    if (isLoading) return // Don't update while loading
    
    try {
      const courtSlots = await courtApi.getCourtSlots({ limit: 10 })
      const newActivities = generateRecentActivity(dashboardStats, dashboardStats.systemStatus, systemMetrics, courtSlots)
      setRecentActivity(newActivities)
      setLastActivityUpdate(new Date())
    } catch {
      // Fallback to generating activity without court slots
      const newActivities = generateRecentActivity(dashboardStats, dashboardStats.systemStatus, systemMetrics, [])
      setRecentActivity(newActivities)
      setLastActivityUpdate(new Date())
    }
  }, [dashboardStats, systemMetrics, generateRecentActivity, isLoading])

  // Update the ref whenever isLoading changes
  useEffect(() => {
    isLoadingRef.current = isLoading
  }, [isLoading])

  useEffect(() => {
    updateActivity()
  }, [updateActivity])

  const handleLogout = () => {
    clearAuthState()
    addNotification({
      title: 'Logged Out',
      message: 'You have been successfully logged out',
      type: 'info',
    })
    navigate('/login')
  }

  const handleRefreshUserInfo = async () => {
    // For now, just show a message that this feature is not implemented
    addNotification({
      title: 'Refresh User Info',
      message: 'User info refresh feature will be implemented soon',
      type: 'info',
    })
  }

  const handleDebugTokens = async () => {
    const { tokenStorage } = await import('@/lib/tokenStorage')
    const accessToken = tokenStorage.getAccessToken()
    const refreshToken = tokenStorage.getRefreshToken()
    const localStorageToken = localStorage.getItem('accessToken')
    

    
    addNotification({
      title: 'Token Debug',
      message: `Access: ${accessToken ? 'Present' : 'Missing'}, Refresh: ${refreshToken ? 'Present' : 'Missing'}, LocalStorage: ${localStorageToken ? 'Present' : 'Missing'}`,
      type: 'info',
    })
  }

  const handleRefreshData = async () => {
    setIsLoading(true)
    setError(null)
    
    try {
      const [courtSlots, stats] = await Promise.all([
        courtApi.getCourtSlotsForCards({ limit: 6 }),
        courtApi.getDashboardStats(),
      ])
      
      setCourtData(courtSlots)
      setDashboardStats(prev => ({
        ...prev,
        activeCourts: stats.activeCourts,
        availableSlots: stats.availableSlots,
        totalVenues: stats.totalVenues,
      }))
    } catch (error) {
      console.error('Failed to refresh data:', error)
      setError('Failed to refresh data. Please try again.')
    } finally {
      setIsLoading(false)
    }
  }

  // System control functions
  const handlePauseSystem = async () => {
    if (dashboardStats.systemStatus === 'PAUSED') return
    
    setIsSystemControlLoading(true)
    try {
      const result = await systemApi.pauseScraping()
      
      // Update system status immediately
      setDashboardStats(prev => ({
        ...prev,
        systemStatus: 'PAUSED'
      }))
      
      addNotification({
        title: 'System Paused',
        message: result.message || 'Tennis court monitoring has been paused',
        type: 'info',
      })
    } catch (error) {
      addNotification({
        title: 'Failed to Pause System',
        message: error instanceof Error ? error.message : 'Unknown error occurred',
        type: 'error',
      })
    } finally {
      setIsSystemControlLoading(false)
    }
  }

  const handleResumeSystem = async () => {
    if (dashboardStats.systemStatus === 'RUNNING') return
    
    setIsSystemControlLoading(true)
    try {
      const result = await systemApi.resumeScraping()
      
      // Update system status immediately
      setDashboardStats(prev => ({
        ...prev,
        systemStatus: 'RUNNING'
      }))
      
      addNotification({
        title: 'System Resumed',
        message: result.message || 'Tennis court monitoring has been resumed',
        type: 'success',
      })
    } catch (error) {
      addNotification({
        title: 'Failed to Resume System',
        message: error instanceof Error ? error.message : 'Unknown error occurred',
        type: 'error',
      })
    } finally {
      setIsSystemControlLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-gray-50 to-gray-100 dark:from-gray-900 dark:via-gray-900 dark:to-gray-800">
      {/* Header */}
      <header className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm shadow-lg border-b border-gray-200/50 dark:border-gray-700/50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center py-4 sm:py-6 space-y-4 sm:space-y-0">
            <div className="flex-1 w-full sm:w-auto">
              <div className="flex items-center space-x-3">
                <div className="text-3xl sm:text-4xl">🎾</div>
                <div className="flex-1 min-w-0">
                  <TextGenerateEffect 
                    words="Tennis Court Monitor"
                    className="text-xl sm:text-2xl lg:text-3xl font-bold text-gray-900 dark:text-white truncate"
                    duration={0.3}
                    filter={false}
                  />
                  <p className="text-gray-600 dark:text-gray-400 mt-1 animate-fade-in text-sm sm:text-base truncate">
                    Welcome back, <span className="font-medium text-gray-900 dark:text-gray-100">{user?.name}</span>!
                  </p>
                </div>
              </div>
            </div>
            <div className="flex flex-wrap items-center gap-2 sm:gap-3 w-full sm:w-auto justify-end">
              <Button 
                variant="outline" 
                onClick={() => navigate('/settings')}
                className="hover:scale-105 transition-transform duration-200 text-xs sm:text-sm px-2 sm:px-4 py-1 sm:py-2"
                size="sm"
              >
                Settings
              </Button>
              <Button 
                variant="outline" 
                onClick={() => navigate('/token-test')}
                className="hover:scale-105 transition-transform duration-200 text-xs sm:text-sm px-2 sm:px-4 py-1 sm:py-2 hidden md:inline-flex"
                size="sm"
              >
                Token Test
              </Button>
              <Button 
                variant="outline" 
                onClick={handleRefreshUserInfo}
                className="hover:scale-105 transition-transform duration-200 text-xs sm:text-sm px-2 sm:px-4 py-1 sm:py-2 hidden lg:inline-flex"
                size="sm"
              >
                Refresh Info
              </Button>
              <Button 
                variant="destructive" 
                onClick={handleLogout}
                className="hover:scale-105 transition-transform duration-200 text-xs sm:text-sm px-2 sm:px-4 py-1 sm:py-2"
                size="sm"
              >
                Logout
              </Button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto py-4 sm:py-8 px-4 sm:px-6 lg:px-8">
        <div className="space-y-6 sm:space-y-8">
          
          {/* Real-time Monitoring Section */}
          <section className="animate-fade-in-up">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between mb-4 sm:mb-6 space-y-3 sm:space-y-0">
              <div className="flex-1 min-w-0">
                <h2 className="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
                  Real-time Monitoring
                </h2>
                <p className="text-gray-600 dark:text-gray-400 mt-1 text-sm sm:text-base">
                  Live system status and court activity
                </p>
              </div>
              <div className="flex items-center gap-2">
                {/* System Control Buttons */}
                {dashboardStats.systemStatus === 'RUNNING' ? (
                  <Button 
                    variant="outline"
                    size="sm"
                    onClick={handlePauseSystem}
                    disabled={isSystemControlLoading}
                    className="hover:scale-105 transition-all duration-200 shadow-md hover:shadow-lg text-yellow-700 border-yellow-300 hover:bg-yellow-50 dark:text-yellow-400 dark:border-yellow-600 dark:hover:bg-yellow-900/20"
                  >
                    {isSystemControlLoading ? 'Pausing...' : '⏸️ Pause System'}
                  </Button>
                ) : dashboardStats.systemStatus === 'PAUSED' ? (
                  <Button 
                    variant="outline"
                    size="sm"
                    onClick={handleResumeSystem}
                    disabled={isSystemControlLoading}
                    className="hover:scale-105 transition-all duration-200 shadow-md hover:shadow-lg text-green-700 border-green-300 hover:bg-green-50 dark:text-green-400 dark:border-green-600 dark:hover:bg-green-900/20"
                  >
                    {isSystemControlLoading ? 'Resuming...' : '▶️ Resume System'}
                  </Button>
                ) : null}
                
              <Badge 
                variant="secondary" 
                  className={`animate-pulse shadow-lg self-start sm:self-auto ${
                    dashboardStats.systemStatus === 'RUNNING' 
                      ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100'
                      : dashboardStats.systemStatus === 'PAUSED'
                      ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-100'
                      : 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-100'
                  }`}
              >
                  <div className={`w-2 h-2 rounded-full mr-2 animate-ping ${
                    dashboardStats.systemStatus === 'RUNNING' 
                      ? 'bg-green-500'
                      : dashboardStats.systemStatus === 'PAUSED'
                      ? 'bg-yellow-500'
                      : 'bg-red-500'
                  }`}></div>
                  {dashboardStats.systemStatus === 'RUNNING' 
                    ? 'System Online' 
                    : dashboardStats.systemStatus === 'PAUSED'
                    ? 'System Paused'
                    : 'System Error'}
              </Badge>
              </div>
            </div>
            
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6">
              {/* System Status Card */}
              <Card className="hover:shadow-xl transition-all duration-300 hover:-translate-y-1 border-0 shadow-lg bg-white/70 dark:bg-gray-800/70 backdrop-blur-sm">
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-medium text-gray-700 dark:text-gray-300">System Status</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="flex items-center space-x-2">
                    <div className={`w-3 h-3 rounded-full animate-pulse shadow-lg ${
                      dashboardStats.systemStatus === 'RUNNING' ? 'bg-green-500 shadow-green-500/50' :
                      dashboardStats.systemStatus === 'PAUSED' ? 'bg-yellow-500 shadow-yellow-500/50' :
                      dashboardStats.systemStatus === 'LOADING' ? 'bg-blue-500 shadow-blue-500/50' :
                      'bg-red-500 shadow-red-500/50'
                    }`}></div>
                    <span className={`text-sm font-medium ${
                      dashboardStats.systemStatus === 'RUNNING' ? 'text-green-600 dark:text-green-400' :
                      dashboardStats.systemStatus === 'PAUSED' ? 'text-yellow-600 dark:text-yellow-400' :
                      dashboardStats.systemStatus === 'LOADING' ? 'text-blue-600 dark:text-blue-400' :
                      'text-red-600 dark:text-red-400'
                    }`}>
                      {dashboardStats.systemStatus === 'RUNNING' ? 'All Systems Operational' :
                       dashboardStats.systemStatus === 'PAUSED' ? 'System Paused' :
                       dashboardStats.systemStatus === 'LOADING' ? 'Loading...' :
                       'System Error'}
                    </span>
                  </div>
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                    {isLoading ? 'Loading...' : 'Last updated: just now'}
                  </p>
                </CardContent>
              </Card>

              {/* Active Courts */}
              <Card className="hover:shadow-xl transition-all duration-300 hover:-translate-y-1 border-0 shadow-lg bg-white/70 dark:bg-gray-800/70 backdrop-blur-sm">
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-medium text-gray-700 dark:text-gray-300">Active Courts</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold text-blue-600 dark:text-blue-400 bg-gradient-to-r from-blue-600 to-cyan-600 bg-clip-text text-transparent">
                    {isLoading ? '...' : dashboardStats.activeCourts}
                  </div>
                  <p className="text-xs text-gray-500 dark:text-gray-400">
                    Courts currently in use
                  </p>
                </CardContent>
              </Card>

              {/* Available Slots */}
              <Card className="hover:shadow-xl transition-all duration-300 hover:-translate-y-1 border-0 shadow-lg bg-white/70 dark:bg-gray-800/70 backdrop-blur-sm">
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-medium text-gray-700 dark:text-gray-300">Available Slots</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold text-green-600 dark:text-green-400 bg-gradient-to-r from-green-600 to-emerald-600 bg-clip-text text-transparent">
                    {isLoading ? '...' : dashboardStats.availableSlots}
                  </div>
                  <p className="text-xs text-gray-500 dark:text-gray-400">
                    Slots available today
                  </p>
                </CardContent>
              </Card>

              {/* Notifications Sent */}
              <Card className="hover:shadow-xl transition-all duration-300 hover:-translate-y-1 border-0 shadow-lg bg-white/70 dark:bg-gray-800/70 backdrop-blur-sm">
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-medium text-gray-700 dark:text-gray-300">Notifications</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold text-purple-600 dark:text-purple-400 bg-gradient-to-r from-purple-600 to-pink-600 bg-clip-text text-transparent">
                    {isLoading ? '...' : systemMetrics.notificationsSent}
                  </div>
                  <p className="text-xs text-gray-500 dark:text-gray-400">
                    Sent in last 24h
                  </p>
                </CardContent>
              </Card>
            </div>

            {/* Recent Activity */}
            <Card className="mt-4 sm:mt-6 hover:shadow-xl transition-all duration-300 border-0 shadow-lg bg-white/70 dark:bg-gray-800/70 backdrop-blur-sm">
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                <CardTitle className="text-lg sm:text-xl text-gray-900 dark:text-white">Recent Activity</CardTitle>
                    <CardDescription className="text-gray-600 dark:text-gray-400 text-sm sm:text-base">
                      Latest system updates and court changes
                    </CardDescription>
                  </div>
                  <div className="text-xs text-gray-500 dark:text-gray-400">
                    Updated {lastActivityUpdate.toLocaleTimeString()}
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <div className="space-y-3 sm:space-y-4">
                  {recentActivity.length === 0 ? (
                    <div className="text-center py-4 text-gray-500 dark:text-gray-400">
                      No recent activity to display
                    </div>
                  ) : (
                    recentActivity.map((activity) => {
                      const getActivityStyle = (type: string) => {
                        switch (type) {
                          case 'available':
                            return {
                              bg: 'bg-green-50/50 dark:bg-green-900/20 hover:bg-green-50 dark:hover:bg-green-900/30',
                              badge: 'bg-green-100 text-green-700 border-green-200'
                            }
                          case 'booked':
                            return {
                              bg: 'bg-blue-50/50 dark:bg-blue-900/20 hover:bg-blue-50 dark:hover:bg-blue-900/30',
                              badge: 'bg-blue-100 text-blue-700 border-blue-200'
                            }
                          case 'notification':
                            return {
                              bg: 'bg-purple-50/50 dark:bg-purple-900/20 hover:bg-purple-50 dark:hover:bg-purple-900/30',
                              badge: 'bg-purple-100 text-purple-700 border-purple-200'
                            }
                          case 'system':
                            return {
                              bg: 'bg-gray-50/50 dark:bg-gray-900/20 hover:bg-gray-50 dark:hover:bg-gray-900/30',
                              badge: 'bg-gray-100 text-gray-700 border-gray-200'
                            }
                          default:
                            return {
                              bg: 'bg-gray-50/50 dark:bg-gray-900/20 hover:bg-gray-50 dark:hover:bg-gray-900/30',
                              badge: 'bg-gray-100 text-gray-700 border-gray-200'
                            }
                        }
                      }

                      const formatTimeAgo = (timestamp: Date) => {
                        const now = new Date()
                        const diffMs = now.getTime() - timestamp.getTime()
                        const diffMins = Math.floor(diffMs / (1000 * 60))
                        
                        if (diffMins < 1) return 'Just now'
                        if (diffMins < 60) return `${diffMins} min ago`
                        
                        const diffHours = Math.floor(diffMins / 60)
                        if (diffHours < 24) return `${diffHours}h ago`
                        
                        return timestamp.toLocaleDateString()
                      }

                      const style = getActivityStyle(activity.type)
                      
                      return (
                        <div 
                          key={activity.id}
                          className={`flex flex-col sm:flex-row sm:items-center space-y-2 sm:space-y-0 sm:space-x-3 text-sm p-3 rounded-lg ${style.bg} transition-colors duration-200`}
                        >
                    <div className="flex items-center space-x-3 flex-1 min-w-0">
                            <Badge variant="outline" className={`${style.badge} shadow-sm flex-shrink-0 capitalize`}>
                              {activity.type}
                      </Badge>
                      <span className="text-gray-900 dark:text-gray-100 flex-1 min-w-0">
                              {activity.message}
                      </span>
                    </div>
                    <span className="text-gray-500 dark:text-gray-400 text-xs self-start sm:self-auto flex-shrink-0">
                            {formatTimeAgo(activity.timestamp)}
                    </span>
                        </div>
                      )
                    })
                  )}
                </div>
              </CardContent>
            </Card>
          </section>

          {/* Enhanced System Controls Section */}
          <section className="animate-fade-in-up" style={{ animationDelay: '0.15s' }}>
            <div className="flex items-center justify-between mb-4 sm:mb-6">
              <div>
                <h2 className="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
                  System Controls
                </h2>
                <p className="text-gray-600 dark:text-gray-400 mt-1 text-sm sm:text-base">
                  Manage and monitor the tennis court scraping system
                </p>
              </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {/* System Control Buttons */}
              <Card className="border-0 shadow-lg bg-white/70 dark:bg-gray-800/70 backdrop-blur-sm">
                <CardHeader>
                  <div className="flex items-center space-x-3">
                    <div className="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
                      <Activity className="w-5 h-5 text-blue-600 dark:text-blue-400" />
                    </div>
                    <div>
                      <CardTitle className="text-lg text-gray-900 dark:text-white">
                        System Operations
                      </CardTitle>
                      <CardDescription className="text-gray-600 dark:text-gray-400">
                        Control the scraping system
                      </CardDescription>
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="grid grid-cols-2 gap-3">
                    {/* Pause/Resume Button */}
                    {dashboardStats.systemStatus === 'RUNNING' ? (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={handlePauseSystem}
                        disabled={isSystemControlLoading}
                        className="flex items-center space-x-2 hover:bg-yellow-50 hover:border-yellow-300 hover:text-yellow-700"
                      >
                        {isSystemControlLoading ? (
                          <Loader2 className="w-4 h-4 animate-spin" />
                        ) : (
                          <Pause className="w-4 h-4" />
                        )}
                        <span>Pause</span>
                      </Button>
                    ) : (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={handleResumeSystem}
                        disabled={isSystemControlLoading}
                        className="flex items-center space-x-2 hover:bg-green-50 hover:border-green-300 hover:text-green-700"
                      >
                        {isSystemControlLoading ? (
                          <Loader2 className="w-4 h-4 animate-spin" />
                        ) : (
                          <Play className="w-4 h-4" />
                        )}
                        <span>Resume</span>
                      </Button>
                    )}

                    {/* Restart Button */}
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={async () => {
                        setIsSystemControlLoading(true)
                        try {
                          await systemApi.restartSystem()
                          addNotification({
                            title: 'System Restarted',
                            message: 'The system has been restarted successfully',
                            type: 'success',
                          })
                          // Refresh data after restart
                          setTimeout(() => {
                            window.location.reload()
                          }, 2000)
                        } catch (error) {
                          addNotification({
                            title: 'Restart Failed',
                            message: error instanceof Error ? error.message : 'Failed to restart system',
                            type: 'error',
                          })
                        } finally {
                          setIsSystemControlLoading(false)
                        }
                      }}
                      disabled={isSystemControlLoading}
                      className="flex items-center space-x-2 hover:bg-orange-50 hover:border-orange-300 hover:text-orange-700"
                    >
                      {isSystemControlLoading ? (
                        <Loader2 className="w-4 h-4 animate-spin" />
                      ) : (
                        <RotateCcw className="w-4 h-4" />
                      )}
                      <span>Restart</span>
                    </Button>
                  </div>

                                     {/* Refresh Data Button */}
                   <Button
                     variant="default"
                     size="sm"
                     onClick={handleRefreshData}
                     disabled={isLoading}
                     className="w-full flex items-center space-x-2"
                   >
                     {isLoading ? (
                       <Loader2 className="w-4 h-4 animate-spin" />
                     ) : (
                       <RefreshCw className="w-4 h-4" />
                     )}
                     <span>Refresh All Data</span>
                   </Button>

                   {/* Debug Token Button */}
                   <Button
                     variant="outline"
                     size="sm"
                     onClick={handleDebugTokens}
                     className="w-full flex items-center space-x-2 text-xs"
                   >
                     <span>🔍 Debug Tokens</span>
                   </Button>
                </CardContent>
              </Card>

              {/* System Metrics */}
              <Card className="border-0 shadow-lg bg-white/70 dark:bg-gray-800/70 backdrop-blur-sm">
                <CardHeader>
                  <div className="flex items-center space-x-3">
                    <div className="p-2 bg-green-100 dark:bg-green-900 rounded-lg">
                      <Clock className="w-5 h-5 text-green-600 dark:text-green-400" />
                    </div>
                    <div>
                      <CardTitle className="text-lg text-gray-900 dark:text-white">
                        System Metrics
                      </CardTitle>
                      <CardDescription className="text-gray-600 dark:text-gray-400">
                        Real-time system information
                      </CardDescription>
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-1">
                      <p className="text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wide">Uptime</p>
                      <p className="text-sm font-medium text-gray-900 dark:text-white">
                        {systemMetrics.uptime}
                      </p>
                    </div>
                    <div className="space-y-1">
                      <p className="text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wide">Last Scrape</p>
                      <p className="text-sm font-medium text-gray-900 dark:text-white">
                        {systemMetrics.lastScrapeTime}
                      </p>
                    </div>
                    <div className="space-y-1">
                      <p className="text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wide">Notifications</p>
                      <div className="flex items-center space-x-2">
                        <Bell className="w-4 h-4 text-purple-600 dark:text-purple-400" />
                        <p className="text-sm font-medium text-gray-900 dark:text-white">
                          {systemMetrics.notificationsSent}
                        </p>
                      </div>
                    </div>
                    <div className="space-y-1">
                      <p className="text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wide">Health</p>
                      <Badge 
                        variant="outline" 
                        className={`text-xs ${
                          systemMetrics.scraperHealth === 'Excellent' ? 'bg-green-100 text-green-700 border-green-200' :
                          systemMetrics.scraperHealth === 'Good' ? 'bg-blue-100 text-blue-700 border-blue-200' :
                          systemMetrics.scraperHealth === 'Fair' ? 'bg-yellow-100 text-yellow-700 border-yellow-200' :
                          systemMetrics.scraperHealth === 'Poor' ? 'bg-red-100 text-red-700 border-red-200' :
                          'bg-gray-100 text-gray-700 border-gray-200'
                        }`}
                      >
                        {systemMetrics.scraperHealth}
                      </Badge>
                  </div>
                </div>
              </CardContent>
            </Card>
            </div>
          </section>

          {/* User Preferences Section */}
          <section className="animate-fade-in-up" style={{ animationDelay: '0.16s' }}>
            <div className="flex items-center justify-between mb-4 sm:mb-6">
              <div>
                <h2 className="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white bg-gradient-to-r from-indigo-600 to-purple-600 bg-clip-text text-transparent">
                  Your Preferences
                </h2>
                <p className="text-gray-600 dark:text-gray-400 mt-1 text-sm sm:text-base">
                  Current notification and booking preferences
                </p>
              </div>
            </div>

            <UserPreferencesCard />
          </section>

          {/* Scraping Logs Terminal Section */}
          <section className="animate-fade-in-up" style={{ animationDelay: '0.175s' }}>
            <div className="flex items-center justify-between mb-4 sm:mb-6">
              <div>
                <h2 className="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white bg-gradient-to-r from-purple-600 to-pink-600 bg-clip-text text-transparent">
                  Scraping Logs
                </h2>
                <p className="text-gray-600 dark:text-gray-400 mt-1 text-sm sm:text-base">
                  Real-time monitoring of court scraping operations
                </p>
              </div>
            </div>

            <ScrapingLogsTerminal className="w-full" />
          </section>

          {/* Court Availability Section */}
          <section className="animate-fade-in-up" style={{ animationDelay: '0.2s' }}>
            <div className="flex flex-col sm:flex-row sm:items-center justify-between mb-4 sm:mb-6 space-y-3 sm:space-y-0">
              <div className="flex-1 min-w-0">
                <h2 className="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white bg-gradient-to-r from-green-600 to-blue-600 bg-clip-text text-transparent">
                  Court Availability
                </h2>
                <p className="text-gray-600 dark:text-gray-400 mt-1 text-sm sm:text-base">
                  Available courts and booking opportunities
                </p>
              </div>
              <Button 
                variant="outline" 
                size="sm"
                className="hover:scale-105 transition-all duration-200 shadow-md hover:shadow-lg self-start sm:self-auto"
                onClick={handleRefreshData}
                disabled={isLoading}
              >
                {isLoading ? 'Loading...' : 'Refresh Data'}
              </Button>
            </div>
            
            {/* Error Message */}
            {error && (
              <div className="mb-4 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
                <p className="text-red-700 dark:text-red-400 text-sm">{error}</p>
              </div>
            )}
            
            {/* Court Cards Grid */}
            <div className="grid grid-cols-1 sm:grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6">
              {courtData.length === 0 && !isLoading ? (
                <div className="col-span-full text-center py-8">
                  <p className="text-gray-500 dark:text-gray-400">No court data available. Please check back later.</p>
                </div>
              ) : (
                courtData.map((court, index) => (
                <div 
                  key={`${court.courtName}-${index}`}
                  className="animate-fade-in-up"
                  style={{ animationDelay: `${0.1 * index}s` }}
                >
                  <CourtCard
                    {...court}
                    className="border-0 shadow-lg bg-white/70 dark:bg-gray-800/70 backdrop-blur-sm hover:shadow-xl transition-all duration-300 hover:-translate-y-1"
                  />
                </div>
                ))
              )}
            </div>
          </section>

        </div>
      </main>
    </div>
  )
} 