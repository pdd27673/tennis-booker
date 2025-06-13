import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { TextGenerateEffect } from '@/components/ui/text-generate-effect'
import { useAuth } from '@/hooks/useAuth'
import { useNavigate } from 'react-router-dom'
import CourtCard, { type CourtCardProps } from '@/components/CourtCard'

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
  const { user, logout, refreshUserInfo } = useAuth()
  const navigate = useNavigate()

  const handleLogout = () => {
    logout()
  }

  const handleRefreshUserInfo = async () => {
    const result = await refreshUserInfo()
    if (result.success) {
      console.log('User information refreshed successfully')
    } else {
      console.error('Failed to refresh user info:', result.error)
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
              <Badge 
                variant="secondary" 
                className="bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100 animate-pulse shadow-lg self-start sm:self-auto"
              >
                <div className="w-2 h-2 bg-green-500 rounded-full mr-2 animate-ping"></div>
                System Online
              </Badge>
            </div>
            
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6">
              {/* System Status Card */}
              <Card className="hover:shadow-xl transition-all duration-300 hover:-translate-y-1 border-0 shadow-lg bg-white/70 dark:bg-gray-800/70 backdrop-blur-sm">
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-medium text-gray-700 dark:text-gray-300">System Status</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="flex items-center space-x-2">
                    <div className="w-3 h-3 bg-green-500 rounded-full animate-pulse shadow-lg shadow-green-500/50"></div>
                    <span className="text-sm font-medium text-green-600 dark:text-green-400">
                      All Systems Operational
                    </span>
                  </div>
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                    Last updated: 2 minutes ago
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
                    4/6
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
                    12
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
                    8
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
                <CardTitle className="text-lg sm:text-xl text-gray-900 dark:text-white">Recent Activity</CardTitle>
                <CardDescription className="text-gray-600 dark:text-gray-400 text-sm sm:text-base">Latest system updates and court changes</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-3 sm:space-y-4">
                  <div className="flex flex-col sm:flex-row sm:items-center space-y-2 sm:space-y-0 sm:space-x-3 text-sm p-3 rounded-lg bg-green-50/50 dark:bg-green-900/20 hover:bg-green-50 dark:hover:bg-green-900/30 transition-colors duration-200">
                    <div className="flex items-center space-x-3 flex-1 min-w-0">
                      <Badge variant="outline" className="bg-green-100 text-green-700 border-green-200 shadow-sm flex-shrink-0">
                        Available
                      </Badge>
                      <span className="text-gray-900 dark:text-gray-100 flex-1 min-w-0">
                        Court 3 at Tennis Club Central - 2:00 PM slot opened
                      </span>
                    </div>
                    <span className="text-gray-500 dark:text-gray-400 text-xs self-start sm:self-auto flex-shrink-0">
                      5 min ago
                    </span>
                  </div>
                  <div className="flex flex-col sm:flex-row sm:items-center space-y-2 sm:space-y-0 sm:space-x-3 text-sm p-3 rounded-lg bg-blue-50/50 dark:bg-blue-900/20 hover:bg-blue-50 dark:hover:bg-blue-900/30 transition-colors duration-200">
                    <div className="flex items-center space-x-3 flex-1 min-w-0">
                      <Badge variant="outline" className="bg-blue-100 text-blue-700 border-blue-200 shadow-sm flex-shrink-0">
                        Booked
                      </Badge>
                      <span className="text-gray-900 dark:text-gray-100 flex-1 min-w-0">
                        Court 1 at Riverside Tennis - 4:00 PM slot filled
                      </span>
                    </div>
                    <span className="text-gray-500 dark:text-gray-400 text-xs self-start sm:self-auto flex-shrink-0">
                      12 min ago
                    </span>
                  </div>
                  <div className="flex flex-col sm:flex-row sm:items-center space-y-2 sm:space-y-0 sm:space-x-3 text-sm p-3 rounded-lg bg-purple-50/50 dark:bg-purple-900/20 hover:bg-purple-50 dark:hover:bg-purple-900/30 transition-colors duration-200">
                    <div className="flex items-center space-x-3 flex-1 min-w-0">
                      <Badge variant="outline" className="bg-purple-100 text-purple-700 border-purple-200 shadow-sm flex-shrink-0">
                        Notification
                      </Badge>
                      <span className="text-gray-900 dark:text-gray-100 flex-1 min-w-0">
                        Alert sent for preferred court availability
                      </span>
                    </div>
                    <span className="text-gray-500 dark:text-gray-400 text-xs self-start sm:self-auto flex-shrink-0">
                      18 min ago
                    </span>
                  </div>
                </div>
              </CardContent>
            </Card>
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
              >
                Refresh Data
              </Button>
            </div>
            
            {/* Court Cards Grid */}
            <div className="grid grid-cols-1 sm:grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6">
              {mockCourtData.map((court, index) => (
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
              ))}
            </div>
          </section>

        </div>
      </main>
    </div>
  )
} 