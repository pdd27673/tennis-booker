import { useState } from 'react'
import { appName, appVersion, environment, apiUrl, features, isFeatureEnabled } from './config/config'

function App() {
  const [activeTab, setActiveTab] = useState('dashboard')

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
              {['dashboard', 'alerts', 'monitoring', 'settings'].map((tab) => (
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
              <div className="flex items-center space-x-2">
                <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
                <span className="text-sm text-gray-600">System Online</span>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="container-app section">
        {activeTab === 'dashboard' && (
          <div className="space-y-8">
            {/* Hero Section */}
            <div className="text-center space-y-6">
              <h1 className="text-5xl lg:text-6xl font-bold text-gray-900">
                Tennis Court
                <span className="text-gradient block">Monitoring Dashboard</span>
              </h1>
              <p className="text-xl text-gray-600 max-w-2xl mx-auto">
                Real-time monitoring and alerting system for tennis court availability. 
                Get notified when your favorite courts become available.
              </p>
              <div className="flex flex-col sm:flex-row gap-4 justify-center">
                <button className="btn btn-primary text-lg px-8 py-3">
                  View Alerts
                </button>
                <button className="btn btn-outline text-lg px-8 py-3">
                  Configure Monitoring
                </button>
              </div>
            </div>

            {/* Status Cards */}
            <div className="grid md:grid-cols-4 gap-6 mt-16">
              {[
                {
                  title: 'Active Monitors',
                  value: '12',
                  description: 'Courts being monitored',
                  icon: '👁️',
                  status: 'success'
                },
                {
                  title: 'Alerts Today',
                  value: '8',
                  description: 'Availability notifications sent',
                  icon: '🔔',
                  status: 'info'
                },
                {
                  title: 'System Health',
                  value: '99.9%',
                  description: 'Uptime this month',
                  icon: '💚',
                  status: 'success'
                },
                {
                  title: 'Last Check',
                  value: '2m ago',
                  description: 'Most recent scan',
                  icon: '⏱️',
                  status: 'info'
                }
              ].map((stat, index) => (
                <div key={index} className="card-hover">
                  <div className="text-center space-y-4">
                    <div className="text-4xl">{stat.icon}</div>
                    <div className="space-y-2">
                      <div className="text-3xl font-bold text-gray-900">{stat.value}</div>
                      <h3 className="text-lg font-semibold text-gray-900">{stat.title}</h3>
                      <p className="text-sm text-gray-600">{stat.description}</p>
                    </div>
                    <div className="flex justify-center">
                      <span className={`badge ${stat.status === 'success' ? 'badge-success' : 'badge-primary'}`}>
                        Active
                      </span>
                    </div>
                  </div>
                </div>
              ))}
            </div>

            {/* Features Grid */}
            <div className="grid md:grid-cols-3 gap-6 mt-16">
              {[
                {
                  title: 'Real-time Monitoring',
                  description: 'Continuous scanning of court availability across multiple platforms',
                  icon: '📊',
                  enabled: true
                },
                {
                  title: 'Smart Alerts',
                  description: 'Get notified instantly when courts become available',
                  icon: '🔔',
                  enabled: isFeatureEnabled('notificationsEnabled')
                },
                {
                  title: 'Health Monitoring',
                  description: 'System health checks and performance monitoring',
                  icon: '🏥',
                  enabled: isFeatureEnabled('advancedSearchEnabled')
                }
              ].map((feature, index) => (
                <div key={index} className="card-hover">
                  <div className="text-center space-y-4">
                    <div className="text-4xl">{feature.icon}</div>
                    <h3 className="text-xl font-semibold text-gray-900">{feature.title}</h3>
                    <p className="text-gray-600">{feature.description}</p>
                    <div className="flex justify-center">
                      <span className={`badge ${feature.enabled ? 'badge-success' : 'badge-warning'}`}>
                        {feature.enabled ? 'Active' : 'Configuring'}
                      </span>
                    </div>
                  </div>
                </div>
              ))}
            </div>

            {/* Configuration Information */}
            {isFeatureEnabled('debugMode') && (
              <div className="card mt-12 bg-gray-50 border-gray-300">
                <h2 className="text-2xl font-semibold text-gray-900 mb-6">System Configuration</h2>
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
                    <strong>Debug Mode Active:</strong> System configuration details are visible. 
                    This section will be hidden in production.
                  </p>
                </div>
              </div>
            )}
          </div>
        )}

        {activeTab === 'alerts' && (
          <div className="text-center space-y-6">
            <h2 className="text-3xl font-bold text-gray-900">Alert Management</h2>
            <div className="card max-w-2xl mx-auto">
              <div className="space-y-4">
                <div className="text-6xl">🔔</div>
                <h3 className="text-xl font-semibold">No Active Alerts</h3>
                <p className="text-gray-600">All monitored courts are currently unavailable. You'll be notified when availability changes.</p>
                <button className="btn btn-primary">
                  Configure Alert Preferences
                </button>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'monitoring' && (
          <div className="text-center space-y-6">
            <h2 className="text-3xl font-bold text-gray-900">Court Monitoring</h2>
            <div className="card max-w-2xl mx-auto">
              <div className="space-y-4">
                <div className="text-6xl">📊</div>
                <h3 className="text-xl font-semibold">Monitoring 12 Courts</h3>
                <p className="text-gray-600">System is actively scanning ClubSpark and Courtsides platforms for availability changes.</p>
                <div className="grid grid-cols-2 gap-4 mt-6">
                  <div className="p-4 bg-green-50 rounded-lg">
                    <div className="text-green-600 font-semibold">ClubSpark</div>
                    <div className="text-sm text-green-700">8 courts monitored</div>
                  </div>
                  <div className="p-4 bg-blue-50 rounded-lg">
                    <div className="text-blue-600 font-semibold">Courtsides</div>
                    <div className="text-sm text-blue-700">4 courts monitored</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'settings' && (
          <div className="text-center space-y-6">
            <h2 className="text-3xl font-bold text-gray-900">System Settings</h2>
            <div className="card max-w-md mx-auto">
              <div className="space-y-4">
                <div className="text-6xl">⚙️</div>
                <h3 className="text-xl font-semibold">Configuration</h3>
                <p className="text-gray-600">Manage monitoring preferences, notification settings, and system configuration.</p>
                <div className="space-y-3">
                  <button className="btn btn-outline w-full">Notification Preferences</button>
                  <button className="btn btn-outline w-full">Monitoring Settings</button>
                  <button className="btn btn-outline w-full">System Health</button>
                </div>
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
            <p className="text-sm">Intelligent tennis court availability monitoring and alerting system.</p>
            <div className="flex justify-center space-x-6 text-sm">
              <a href="#" className="hover:text-white transition-colors">System Status</a>
              <a href="#" className="hover:text-white transition-colors">Documentation</a>
              <a href="#" className="hover:text-white transition-colors">Support</a>
              <a href="#" className="hover:text-white transition-colors">API</a>
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}

export default App
