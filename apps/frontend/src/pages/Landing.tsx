// React import removed as it's not needed with modern React
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { BackgroundGradientAnimation } from '@/components/ui/background-gradient-animation'
import { TextGenerateEffect } from '@/components/ui/text-generate-effect'
import { useNavigate } from 'react-router-dom'
import { 
  Target, 
  Zap, 
  Clock, 
  Bell, 
  Activity, 
  Users,
  CheckCircle,
  ArrowRight,
  Play,
  TrendingUp,
  Shield,
  Smartphone,
  Globe
} from 'lucide-react'

const Landing = () => {
  const navigate = useNavigate()

  const features = [
    {
      icon: Target,
      title: "London Court Monitoring",
      description: "Automated system monitors tennis venues across London for availability updates",
      color: "text-blue-500"
    },
    {
      icon: Zap,
      title: "Quick Notifications", 
      description: "Get notified when your preferred courts become available",
      color: "text-yellow-500"
    },
    {
      icon: Clock,
      title: "Perfect Timing",
      description: "Set custom time preferences and never miss prime playing hours again",
      color: "text-green-500"
    },
    {
      icon: Bell,
      title: "Custom Notifications",
      description: "Filter notifications to only get alerts for courts that match your preferences",
      color: "text-purple-500"
    }
  ]

  const stats = [
    { label: "Courts Monitored", value: "50+", icon: Activity },
    { label: "Daily Alerts Sent", value: "500+", icon: Bell },
    { label: "Happy Players", value: "1.2K+", icon: Users },
    { label: "Success Rate", value: "94%", icon: TrendingUp }
  ]


  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-white to-slate-100 dark:from-slate-900 dark:via-slate-800 dark:to-slate-900">
      
      {/* Navigation */}
      <nav className="relative z-50 bg-white/80 dark:bg-slate-900/80 backdrop-blur-md border-b border-slate-200/50 dark:border-slate-700/50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center space-x-3">
              <div className="text-2xl">🎾</div>
              <span className="text-xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
                CourtScout
              </span>
            </div>
            <div className="flex items-center space-x-4">
              <Button variant="ghost" onClick={() => navigate('/login')}>
                Sign In
              </Button>
              <Button onClick={() => navigate('/register')} className="bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700">
                Get Started
              </Button>
            </div>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <section className="relative overflow-hidden">
        <BackgroundGradientAnimation>
          <div className="absolute inset-0 z-10 flex items-center justify-center">
            <div className="max-w-4xl mx-auto text-center px-4 sm:px-6 lg:px-8 py-20">
              <Badge variant="secondary" className="mb-6 bg-white/20 backdrop-blur-sm text-white border-white/30">
                🎾 London Tennis Court Monitoring
              </Badge>
              
              <TextGenerateEffect 
                words="Never Miss Your Perfect Court Again"
                className="text-4xl sm:text-5xl lg:text-6xl font-bold text-white mb-6 leading-tight"
                duration={0.5}
              />
              
              <p className="text-xl text-white/90 mb-8 leading-relaxed max-w-2xl mx-auto">
                CourtScout monitors London tennis courts and alerts you when your preferred venues have availability. 
                Stop manually checking booking sites - let our system do the work for you.
              </p>
              
              <div className="flex flex-col sm:flex-row gap-4 justify-center items-center mb-12">
                <Button 
                  size="lg" 
                  onClick={() => navigate('/register')}
                  className="bg-white text-blue-600 hover:bg-white/90 shadow-lg hover:shadow-xl transition-all duration-300 group"
                >
                  Get Started Free
                  <ArrowRight className="w-4 h-4 ml-2 group-hover:translate-x-1 transition-transform" />
                </Button>
                <Button 
                  variant="outline" 
                  size="lg"
                  onClick={() => navigate('/demo')}
                  className="border-white text-white hover:bg-white/20 backdrop-blur-sm bg-black/20"
                >
                  <Play className="w-4 h-4 mr-2" />
                  View Demo
                </Button>
              </div>

              {/* Quick Stats */}
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-6 max-w-2xl mx-auto">
                {stats.map((stat, index) => (
                  <div key={index} className="text-center">
                    <div className="text-2xl sm:text-3xl font-bold text-white">{stat.value}</div>
                    <div className="text-sm text-white/80">{stat.label}</div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </BackgroundGradientAnimation>
      </section>

      {/* Features Section */}
      <section className="py-20 bg-white dark:bg-slate-900">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-3xl sm:text-4xl font-bold text-gray-900 dark:text-white mb-4">
              How CourtScout Works
            </h2>
            <p className="text-xl text-gray-600 dark:text-gray-400 max-w-2xl mx-auto">
              Our intelligent monitoring system takes the hassle out of finding available tennis courts
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
            {features.map((feature, index) => (
              <Card key={index} className="border-0 shadow-lg hover:shadow-xl transition-all duration-300 hover:-translate-y-2 bg-gradient-to-b from-white to-gray-50/50 dark:from-gray-800 dark:to-gray-900/50">
                <CardHeader className="text-center pb-4">
                  <div className={`w-12 h-12 mx-auto mb-4 rounded-lg bg-gradient-to-br from-blue-500/10 to-purple-500/10 flex items-center justify-center`}>
                    <feature.icon className={`w-6 h-6 ${feature.color}`} />
                  </div>
                  <CardTitle className="text-lg font-semibold text-gray-900 dark:text-white">
                    {feature.title}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <CardDescription className="text-gray-600 dark:text-gray-400 text-center leading-relaxed">
                    {feature.description}
                  </CardDescription>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>


      {/* Security & Trust Section */}
      <section className="py-20 bg-white dark:bg-slate-900">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-3xl sm:text-4xl font-bold text-gray-900 dark:text-white mb-4">
              Built for Modern Tennis Players
            </h2>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            <div className="text-center">
              <div className="w-16 h-16 mx-auto mb-6 rounded-lg bg-gradient-to-br from-green-500/10 to-emerald-500/10 flex items-center justify-center">
                <Shield className="w-8 h-8 text-green-500" />
              </div>
              <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-3">
                Secure & Private
              </h3>
              <p className="text-gray-600 dark:text-gray-400">
                Your data is encrypted and never shared. We only monitor public booking availability.
              </p>
            </div>

            <div className="text-center">
              <div className="w-16 h-16 mx-auto mb-6 rounded-lg bg-gradient-to-br from-blue-500/10 to-cyan-500/10 flex items-center justify-center">
                <Smartphone className="w-8 h-8 text-blue-500" />
              </div>
              <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-3">
                Mobile Optimized
              </h3>
              <p className="text-gray-600 dark:text-gray-400">
                Perfect experience on any device. Get alerts on your phone, tablet, or desktop.
              </p>
            </div>

            <div className="text-center">
              <div className="w-16 h-16 mx-auto mb-6 rounded-lg bg-gradient-to-br from-purple-500/10 to-pink-500/10 flex items-center justify-center">
                <Globe className="w-8 h-8 text-purple-500" />
              </div>
              <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-3">
                Always Available
              </h3>
              <p className="text-gray-600 dark:text-gray-400">
                99.9% uptime guarantee. Our monitoring never sleeps so you never miss opportunities.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-20 bg-gradient-to-r from-blue-600 to-purple-600">
        <div className="max-w-4xl mx-auto text-center px-4 sm:px-6 lg:px-8">
          <h2 className="text-3xl sm:text-4xl font-bold text-white mb-6">
            Ready to Never Miss a Court Again?
          </h2>
          <p className="text-xl text-white/90 mb-8 leading-relaxed">
            Join tennis players using CourtScout to find available courts in London.
            Get started today - it's completely free!
          </p>
          
          <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
            <Button 
              size="lg" 
              onClick={() => navigate('/register')}
              className="bg-white text-blue-600 hover:bg-white/90 shadow-lg hover:shadow-xl transition-all duration-300 group"
            >
              Get Started Free
              <CheckCircle className="w-4 h-4 ml-2 group-hover:scale-110 transition-transform" />
            </Button>
            <Button 
              variant="outline" 
              size="lg"
              onClick={() => navigate('/login')}
              className="border-white text-white hover:bg-white/20 backdrop-blur-sm bg-black/20"
            >
              Already have an account?
            </Button>
          </div>
          
          <p className="text-white/70 text-sm mt-6">
            Always free • No setup fees • No credit card required
          </p>
        </div>
      </section>

      {/* Footer */}
      <footer className="bg-slate-900 text-white py-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
            <div className="col-span-1 md:col-span-2">
              <div className="flex items-center space-x-3 mb-4">
                <div className="text-2xl">🎾</div>
                <span className="text-xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
                  CourtScout
                </span>
              </div>
              <p className="text-gray-400 max-w-md">
                The smartest way to find and book tennis courts in London. 
                Never miss your perfect playing opportunity again.
              </p>
            </div>
            
            <div>
              <h4 className="font-semibold mb-4">Product</h4>
              <ul className="space-y-2 text-gray-400">
                <li><button onClick={() => navigate('/demo')} className="hover:text-white transition-colors text-left">Demo</button></li>
                <li><button onClick={() => navigate('/register')} className="hover:text-white transition-colors text-left">Get Started</button></li>
                <li><a href="/api-docs.json" target="_blank" className="hover:text-white transition-colors">API Docs</a></li>
              </ul>
            </div>
            
            <div>
              <h4 className="font-semibold mb-4">Support</h4>
              <ul className="space-y-2 text-gray-400">
                <li><button onClick={() => navigate('/login')} className="hover:text-white transition-colors text-left">Contact</button></li>
                <li><a href="http://localhost:8080/api/health" target="_blank" className="hover:text-white transition-colors">Status</a></li>
              </ul>
            </div>
          </div>
          
          <div className="border-t border-gray-800 mt-8 pt-8 text-center text-gray-400">
            <p>&copy; 2025 CourtScout. All rights reserved.</p>
          </div>
        </div>
      </footer>
    </div>
  )
}

export default Landing