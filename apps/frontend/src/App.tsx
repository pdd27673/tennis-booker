import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { QueryClientProvider } from '@tanstack/react-query'

import Landing from '@/pages/Landing'
import Demo from '@/pages/Demo'
import Login from '@/pages/Login'
import Register from '@/pages/Register'
import Dashboard from '@/pages/Dashboard'
import Settings from '@/pages/Settings'
import Notifications from '@/pages/Notifications'
import { ProtectedRoute } from '@/components/ProtectedRoute'
import { ErrorBoundary } from '@/components/ErrorBoundary'
import { queryClient } from '@/lib/queryClient'

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <Router>
        <div className="min-h-screen bg-background">
          <Routes>
            <Route path="/" element={<Landing />} />
            <Route path="/demo" element={<Demo />} />
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            <Route 
              path="/dashboard" 
              element={
                <ErrorBoundary>
                  <ProtectedRoute>
                    <Dashboard />
                  </ProtectedRoute>
                </ErrorBoundary>
              } 
            />
            <Route 
              path="/settings" 
              element={
                <ErrorBoundary>
                  <ProtectedRoute>
                    <Settings />
                  </ProtectedRoute>
                </ErrorBoundary>
              } 
            />
            <Route 
              path="/notifications" 
              element={
                <ErrorBoundary>
                  <ProtectedRoute>
                    <Notifications />
                  </ProtectedRoute>
                </ErrorBoundary>
              } 
            />

            <Route path="*" element={<Landing />} />
          </Routes>
        </div>
      </Router>

    </QueryClientProvider>
  )
}

export default App
