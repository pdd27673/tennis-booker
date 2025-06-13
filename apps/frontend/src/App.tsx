import { useState } from 'react'
import { appName, appVersion, environment, apiUrl, features, isFeatureEnabled } from './config/config'

function App() {
  const [activeTab, setActiveTab] = useState('home')

  return (
    <div className="min-h-screen bg-gradient-to-br from-primary-50 to-secondary-50">
      {/* Header */}
      <header className="bg-white shadow-lg border-b border-gray-200">
        <div className="container-app">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center space-x-4">
              <div className="w-8 h-8 bg-gradient-primary rounded-lg flex items-center justify-center">
                <span className="text-white font-bold text-sm">🎾</span>
              </div>
              <h1 className="text-xl font-bold text-gradient">{appName}</h1>
              <span className="badge badge-primary">v{appVersion}</span>
            </div>
            
            <nav className="hidden md:flex space-x-1">
              {['home', 'courts', 'bookings', 'profile'].map((tab) => (
                <button
                  key={tab}
                  onClick={() => setActiveTab(tab)}
                  className={`px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200 capitalize ${
                    activeTab === tab
                      ? 'bg-primary-100 text-primary-700'
                      : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
                  }`}
                >
                  {tab}
                </button>
              ))}
            </nav>
            
            <div className="flex items-center space-x-3">
              <span className={`badge ${environment === 'production' ? 'badge-success' : 'badge-warning'}`}>
                {environment}
              </span>
              <button className="btn btn-primary">
                Sign In
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="container-app section">
        {activeTab === 'home' && (
          <div className="space-y-8">
            {/* Hero Section */}
            <div className="text-center space-y-6">
                          <h1 className="text-5xl lg:text-6xl font-bold text-gray-900">
              Book Your Perfect
              <span className="text-gradient block">Tennis Court</span>
            </h1>
            <p className="text-xl text-gray-600 max-w-2xl mx-auto">
              Find and reserve tennis courts at the best venues near you. 
              Easy booking, instant confirmation, and competitive prices.
            </p>
              <div className="flex flex-col sm:flex-row gap-4 justify-center">
                <button className="btn btn-primary text-lg px-8 py-3">
                  Find Courts
                </button>
                <button className="btn btn-outline text-lg px-8 py-3">
                  Learn More
                </button>
              </div>
            </div>

            {/* Features Grid */}
            <div className="grid md:grid-cols-3 gap-6 mt-16">
              {[
                {
                  title: 'Easy Booking',
                  description: 'Book courts in just a few clicks with our intuitive interface',
                  icon: '📅',
                  enabled: isFeatureEnabled('advancedSearchEnabled')
                },
                {
                  title: 'Real-time Availability',
                  description: 'See live court availability and book instantly',
                  icon: '⚡',
                  enabled: true
                },
                {
                  title: 'Smart Notifications',
                  description: 'Get notified about booking confirmations and reminders',
                  icon: '🔔',
                  enabled: isFeatureEnabled('notificationsEnabled')
                }
              ].map((feature, index) => (
                <div key={index} className="card-hover animate-fade-in">
                  <div className="text-center space-y-4">
                    <div className="text-4xl">{feature.icon}</div>
                    <h3 className="text-xl font-semibold text-gray-900">{feature.title}</h3>
                    <p className="text-gray-600">{feature.description}</p>
                    <div className="flex justify-center">
                      <span className={`badge ${feature.enabled ? 'badge-success' : 'badge-warning'}`}>
                        {feature.enabled ? 'Available' : 'Coming Soon'}
                      </span>
                    </div>
                  </div>
                </div>
              ))}
            </div>

            {/* Configuration Information */}
                         {isFeatureEnabled('debugMode') && (
               <div className="card mt-12 bg-gray-50 border-gray-300">
                 <h2 className="text-2xl font-semibold text-gray-900 mb-6">Configuration Information</h2>
                <div className="grid md:grid-cols-2 gap-6">
                  <div className="space-y-3">
                                         <div className="flex justify-between">
                       <span className="font-medium text-gray-700">App Name:</span>
                       <span className="text-gray-900">{appName}</span>
                     </div>
                     <div className="flex justify-between">
                       <span className="font-medium text-gray-700">Version:</span>
                       <span className="text-gray-900">{appVersion}</span>
                     </div>
                     <div className="flex justify-between">
                       <span className="font-medium text-gray-700">Environment:</span>
                       <span className={`badge ${environment === 'production' ? 'badge-success' : 'badge-warning'}`}>
                         {environment}
                       </span>
                     </div>
                     <div className="flex justify-between">
                       <span className="font-medium text-gray-700">API URL:</span>
                       <span className="text-gray-900 font-mono text-sm">{apiUrl}</span>
                     </div>
                  </div>
                  
                                     <div className="space-y-3">
                     <h3 className="font-semibold text-gray-900">Feature Flags:</h3>
                     <div className="space-y-2">
                       {Object.entries(features).map(([feature, enabled]) => (
                         <div key={feature} className="flex justify-between items-center">
                           <span className="text-sm text-gray-600 capitalize">
                             {feature.replace(/([A-Z])/g, ' $1').toLowerCase()}
                           </span>
                           <span className={`badge ${enabled ? 'badge-success' : 'badge-error'}`}>
                             {enabled ? '✅' : '❌'}
                           </span>
                         </div>
                       ))}
                     </div>
                   </div>
                </div>
                
                <div className="mt-6 p-4 bg-primary-50 rounded-lg border border-primary-200">
                  <p className="text-sm text-primary-700">
                    <strong>Debug Mode Active:</strong> Configuration details are visible. 
                    This section will be hidden in production.
                  </p>
                </div>
              </div>
            )}
          </div>
        )}

                 {activeTab === 'courts' && (
           <div className="text-center space-y-6">
             <h2 className="text-3xl font-bold text-gray-900">Find Tennis Courts</h2>
            <div className="card max-w-md mx-auto">
              <div className="space-y-4">
                <div>
                  <label className="label">Location</label>
                  <input type="text" className="input" placeholder="Enter city or postcode" />
                </div>
                <div>
                  <label className="label">Date</label>
                  <input type="date" className="input" />
                </div>
                <div>
                  <label className="label">Time</label>
                  <select className="input">
                    <option>Morning (6AM - 12PM)</option>
                    <option>Afternoon (12PM - 6PM)</option>
                    <option>Evening (6PM - 10PM)</option>
                  </select>
                </div>
                <button className="btn btn-primary w-full">Search Courts</button>
              </div>
            </div>
          </div>
        )}

                 {activeTab === 'bookings' && (
           <div className="text-center space-y-6">
             <h2 className="text-3xl font-bold text-gray-900">My Bookings</h2>
             <div className="card max-w-2xl mx-auto">
               <p className="text-gray-600">No bookings found. Start by searching for courts!</p>
              <button 
                onClick={() => setActiveTab('courts')} 
                className="btn btn-primary mt-4"
              >
                Find Courts
              </button>
            </div>
          </div>
        )}

                 {activeTab === 'profile' && (
           <div className="text-center space-y-6">
             <h2 className="text-3xl font-bold text-gray-900">Profile</h2>
             <div className="card max-w-md mx-auto">
               <div className="space-y-4">
                 <div className="w-20 h-20 bg-gradient-primary rounded-full mx-auto flex items-center justify-center">
                   <span className="text-white text-2xl">👤</span>
                 </div>
                 <h3 className="text-xl font-semibold">Guest User</h3>
                 <p className="text-gray-600">Sign in to access your profile and booking history</p>
                <button className="btn btn-primary w-full">Sign In</button>
              </div>
            </div>
          </div>
        )}
      </main>

      {/* Footer */}
      <footer className="bg-gray-900 text-gray-300 mt-16">
        <div className="container-app py-8">
          <div className="text-center space-y-4">
            <div className="flex items-center justify-center space-x-2">
              <span className="text-2xl">🎾</span>
              <span className="font-bold text-white">{appName}</span>
            </div>
            <p className="text-sm">Making tennis court booking simple and accessible for everyone.</p>
            <div className="flex justify-center space-x-6 text-sm">
              <a href="#" className="hover:text-white transition-colors">About</a>
              <a href="#" className="hover:text-white transition-colors">Contact</a>
              <a href="#" className="hover:text-white transition-colors">Privacy</a>
              <a href="#" className="hover:text-white transition-colors">Terms</a>
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}

export default App
