import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { useNavigate } from 'react-router-dom'
import { 
  ArrowLeft,
  Play, 
  Pause,
  MapPin,
  Clock,
  Bell,
  Activity,
  CheckCircle,
  AlertCircle,
  Calendar,
  RefreshCw
} from 'lucide-react'

const Demo = () => {
  const navigate = useNavigate()
  const [isPlaying, setIsPlaying] = useState(false)
  const [currentStep, setCurrentStep] = useState(0)
  const [notifications, setNotifications] = useState<Array<{
    id: number
    court: string
    venue: string
    time: string
    status: 'new' | 'viewed'
  }>>([])

  const demoSteps = [
    {
      title: "Court Monitoring Active",
      description: "CourtScout is monitoring your preferred venues in real-time",
      action: "monitoring"
    },
    {
      title: "Court Availability Detected",
      description: "A court matching your preferences becomes available",
      action: "detection"
    },
    {
      title: "Instant Notification Sent",
      description: "You receive an immediate alert with booking details",
      action: "notification"
    },
    {
      title: "Quick Booking Action",
      description: "Click to book directly from the notification",
      action: "booking"
    }
  ]

  const mockCourts = [
    {
      court: "Court 1",
      venue: "Victoria Park Tennis Centre",
      time: "Today 6:00 PM - 7:00 PM",
      price: "£25",
      status: "available" as const
    },
    {
      court: "Court 3", 
      venue: "Stratford Park Tennis Club",
      time: "Tomorrow 11:00 AM - 12:00 PM", 
      price: "£20",
      status: "available" as const
    },
    {
      court: "Court 2",
      venue: "Ropemakers Field",
      time: "Today 8:00 PM - 9:00 PM",
      price: "£30", 
      status: "available" as const
    }
  ]

  useEffect(() => {
    let interval: ReturnType<typeof setInterval>
    
    if (isPlaying) {
      interval = setInterval(() => {
        setCurrentStep((prev) => {
          const next = (prev + 1) % demoSteps.length
          
          // Add notification when we reach the notification step
          if (next === 2) {
            const randomCourt = mockCourts[Math.floor(Math.random() * mockCourts.length)]
            const newNotification = {
              id: Date.now(),
              court: randomCourt?.court || 'Court 1',
              venue: randomCourt?.venue || 'Victoria Park Tennis Centre',
              time: randomCourt?.time || 'Today 6:00 PM - 7:00 PM',
              status: 'new' as const
            }
            setNotifications(prev => [newNotification, ...prev.slice(0, 2)])
          }
          
          return next
        })
      }, 3000)
    }

    return () => clearInterval(interval)
  }, [isPlaying])

  const handlePlayPause = () => {
    setIsPlaying(!isPlaying)
  }

  const resetDemo = () => {
    setIsPlaying(false)
    setCurrentStep(0)
    setNotifications([])
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-white to-slate-100 dark:from-slate-900 dark:via-slate-800 dark:to-slate-900">
      
      {/* Header */}
      <header className="bg-white/80 dark:bg-slate-900/80 backdrop-blur-md border-b border-slate-200/50 dark:border-slate-700/50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center space-x-4">
              <Button
                variant="outline"
                size="sm"
                onClick={() => navigate('/')}
                className="hover:scale-105 transition-transform duration-200"
              >
                <ArrowLeft className="w-4 h-4 mr-2" />
                Back to Home
              </Button>
              <div className="flex items-center space-x-3">
                <div className="text-2xl">🎾</div>
                <span className="text-xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
                  CourtScout Demo
                </span>
              </div>
            </div>
            <Button onClick={() => navigate('/register')} className="bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700">
              Get Started Free
            </Button>
          </div>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        
        {/* Demo Introduction */}
        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold text-gray-900 dark:text-white mb-4">
            See CourtScout in Action
          </h1>
          <p className="text-xl text-gray-600 dark:text-gray-400 mb-8 max-w-3xl mx-auto">
            Watch how CourtScout monitors London tennis courts and sends you instant notifications 
            when your preferred venues have availability.
          </p>
          
          <div className="flex items-center justify-center gap-4">
            <Button
              onClick={handlePlayPause}
              className="bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700"
              size="lg"
            >
              {isPlaying ? (
                <>
                  <Pause className="w-5 h-5 mr-2" />
                  Pause Demo
                </>
              ) : (
                <>
                  <Play className="w-5 h-5 mr-2" />
                  Start Demo
                </>
              )}
            </Button>
            <Button
              onClick={resetDemo}
              variant="outline"
              size="lg"
            >
              <RefreshCw className="w-5 h-5 mr-2" />
              Reset
            </Button>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          
          {/* Demo Steps */}
          <div className="space-y-6">
            <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">
              How It Works
            </h2>
            
            {demoSteps.map((step, index) => (
              <Card 
                key={index} 
                className={`transition-all duration-300 ${
                  currentStep === index 
                    ? 'ring-2 ring-blue-500 shadow-lg scale-105' 
                    : 'shadow-sm hover:shadow-md'
                }`}
              >
                <CardHeader className="pb-3">
                  <div className="flex items-center space-x-3">
                    <div className={`w-8 h-8 rounded-full flex items-center justify-center font-semibold text-sm ${
                      currentStep === index 
                        ? 'bg-blue-600 text-white' 
                        : currentStep > index 
                        ? 'bg-green-600 text-white'
                        : 'bg-gray-200 text-gray-600'
                    }`}>
                      {currentStep > index ? <CheckCircle className="w-4 h-4" /> : index + 1}
                    </div>
                    <CardTitle className="text-lg">{step.title}</CardTitle>
                    {currentStep === index && isPlaying && (
                      <div className="w-2 h-2 bg-blue-500 rounded-full animate-pulse" />
                    )}
                  </div>
                </CardHeader>
                <CardContent>
                  <CardDescription className="text-base">
                    {step.description}
                  </CardDescription>
                </CardContent>
              </Card>
            ))}
          </div>

          {/* Live Demo Panel */}
          <div className="space-y-6">
            <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">
              Live Demo
            </h2>

            {/* Monitoring Status */}
            <Card className="border-0 shadow-lg">
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle className="flex items-center gap-2">
                    <Activity className="w-5 h-5 text-blue-600" />
                    System Status
                  </CardTitle>
                  <Badge variant={isPlaying ? "default" : "secondary"} className={
                    isPlaying ? "bg-green-100 text-green-800" : "bg-gray-100 text-gray-800"
                  }>
                    {isPlaying ? "Monitoring Active" : "Demo Paused"}
                  </Badge>
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span className="text-gray-500">Courts Monitored:</span>
                    <div className="font-semibold">3 venues</div>
                  </div>
                  <div>
                    <span className="text-gray-500">Next Check:</span>
                    <div className="font-semibold">{isPlaying ? 'Live' : 'Paused'}</div>
                  </div>
                </div>
                
                {currentStep >= 1 && (
                  <div className="p-3 bg-blue-50 dark:bg-blue-900/20 rounded-lg border border-blue-200 dark:border-blue-800">
                    <div className="flex items-center gap-2 text-blue-700 dark:text-blue-400">
                      <AlertCircle className="w-4 h-4" />
                      <span className="text-sm font-medium">
                        New availability detected at {mockCourts[0]?.venue || 'Victoria Park Tennis Centre'}
                      </span>
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Available Courts */}
            <Card className="border-0 shadow-lg">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Calendar className="w-5 h-5 text-green-600" />
                  Available Courts
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                {mockCourts.map((court, index) => (
                  <div 
                    key={index}
                    className={`p-3 rounded-lg border transition-all duration-300 ${
                      currentStep >= 1 && index === 0
                        ? 'bg-green-50 border-green-200 dark:bg-green-900/20 dark:border-green-800 animate-pulse'
                        : 'bg-gray-50 border-gray-200 dark:bg-gray-800 dark:border-gray-700'
                    }`}
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-1">
                          <h4 className="font-medium text-gray-900 dark:text-white">
                            {court.court}
                          </h4>
                          <Badge variant="outline" className="text-xs bg-green-100 text-green-700 border-green-200">
                            Available
                          </Badge>
                        </div>
                        <div className="space-y-1 text-sm text-gray-600 dark:text-gray-400">
                          <div className="flex items-center gap-1">
                            <MapPin className="w-3 h-3" />
                            {court.venue}
                          </div>
                          <div className="flex items-center gap-1">
                            <Clock className="w-3 h-3" />
                            {court.time}
                          </div>
                        </div>
                      </div>
                      <div className="text-right">
                        <div className="font-semibold text-gray-900 dark:text-white">
                          {court.price}
                        </div>
                        <div className="text-xs text-gray-500">per hour</div>
                      </div>
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>

            {/* Notifications */}
            <Card className="border-0 shadow-lg">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Bell className="w-5 h-5 text-purple-600" />
                  Live Notifications
                  {notifications.length > 0 && (
                    <Badge variant="default" className="bg-purple-100 text-purple-800">
                      {notifications.length}
                    </Badge>
                  )}
                </CardTitle>
              </CardHeader>
              <CardContent>
                {notifications.length === 0 ? (
                  <div className="text-center py-6 text-gray-500 dark:text-gray-400">
                    <Bell className="w-8 h-8 mx-auto mb-2 opacity-50" />
                    <p className="text-sm">No notifications yet</p>
                    <p className="text-xs">Start the demo to see live alerts</p>
                  </div>
                ) : (
                  <div className="space-y-3">
                    {notifications.map((notification) => (
                      <div 
                        key={notification.id}
                        className="p-3 bg-purple-50 dark:bg-purple-900/20 rounded-lg border border-purple-200 dark:border-purple-800 animate-slide-down"
                      >
                        <div className="flex items-start justify-between">
                          <div className="flex-1">
                            <div className="flex items-center gap-2 mb-1">
                              <span className="text-sm font-medium text-purple-900 dark:text-purple-100">
                                Court Available!
                              </span>
                              <Badge variant="outline" className="text-xs bg-purple-100 text-purple-700 border-purple-300">
                                New
                              </Badge>
                            </div>
                            <p className="text-sm text-purple-800 dark:text-purple-200">
                              {notification.court} at {notification.venue}
                            </p>
                            <p className="text-xs text-purple-600 dark:text-purple-300">
                              {notification.time}
                            </p>
                          </div>
                          <Button size="sm" className="bg-purple-600 hover:bg-purple-700 text-white">
                            Book Now
                          </Button>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </div>

        {/* Call to Action */}
        <div className="mt-16 text-center">
          <Card className="max-w-2xl mx-auto bg-gradient-to-r from-blue-50 to-purple-50 dark:from-blue-900/20 dark:to-purple-900/20 border-0 shadow-xl">
            <CardContent className="p-8">
              <h3 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">
                Ready to Try CourtScout?
              </h3>
              <p className="text-gray-600 dark:text-gray-400 mb-6">
                Stop manually checking tennis court websites. Let CourtScout monitor London venues 
                and notify you when courts become available.
              </p>
              <div className="flex flex-col sm:flex-row gap-4 justify-center">
                <Button 
                  size="lg"
                  onClick={() => navigate('/register')}
                  className="bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700"
                >
                  Get Started Free
                </Button>
                <Button 
                  variant="outline" 
                  size="lg"
                  onClick={() => navigate('/login')}
                >
                  Sign In
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}

export default Demo