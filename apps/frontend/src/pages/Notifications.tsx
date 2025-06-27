import React, { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { userApi } from '@/services/userApi'
import type { UserNotification, NotificationResponse } from '@/services/userApi'
import { useAppStore } from '@/stores/appStore'

const Notifications: React.FC = () => {
  const [notifications, setNotifications] = useState<UserNotification[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [page, setPage] = useState(0)
  const [totalPages, setTotalPages] = useState(0)
  const [totalNotifications, setTotalNotifications] = useState(0)
  
  const { addNotification } = useAppStore()
  const navigate = useNavigate()

  const loadNotifications = async (pageNum: number = 0) => {
    try {
      setLoading(true)
      setError(null)
      
      const response: NotificationResponse = await userApi.getUserNotifications(pageNum, 20)
      
      setNotifications(response.notifications)
      setPage(response.pagination.page)
      setTotalPages(response.pagination.totalPages)
      setTotalNotifications(response.pagination.total)
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load notifications'
      setError(errorMessage)
      addNotification({
        type: 'error',
        title: 'Error',
        message: errorMessage,
      })
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadNotifications()
  }, [])

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleDateString('en-US', {
      weekday: 'short',
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    })
  }

  const formatTime = (timeString: string) => {
    const [hours = '0', minutes = '00'] = timeString.split(':')
    const hour24 = parseInt(hours, 10)
    const hour12 = hour24 === 0 ? 12 : hour24 > 12 ? hour24 - 12 : hour24
    const ampm = hour24 >= 12 ? 'PM' : 'AM'
    return `${hour12}:${minutes} ${ampm}`
  }

  const getNotificationIcon = (type: string) => {
    switch (type) {
      case 'availability':
        return '🎾'
      case 'booking':
        return '📅'
      case 'alert':
        return '⚠️'
      default:
        return '📩'
    }
  }

  const getStatusBadge = (emailSent: boolean, emailStatus: string) => {
    if (!emailSent) {
      return <span className="px-2 py-1 text-xs bg-yellow-100 text-yellow-800 rounded-full">Pending</span>
    }
    
    if (emailStatus === 'sent') {
      return <span className="px-2 py-1 text-xs bg-green-100 text-green-800 rounded-full">Sent</span>
    }
    
    return <span className="px-2 py-1 text-xs bg-red-100 text-red-800 rounded-full">Failed</span>
  }

  const handlePageChange = (newPage: number) => {
    if (newPage >= 0 && newPage < totalPages) {
      loadNotifications(newPage)
    }
  }

  if (loading && notifications.length === 0) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 p-6">
        <div className="max-w-4xl mx-auto">
          <div className="bg-white rounded-lg shadow-sm p-8">
            <div className="animate-pulse">
              <div className="h-8 bg-gray-200 rounded w-1/3 mb-6"></div>
              <div className="space-y-4">
                {[...Array(5)].map((_, i) => (
                  <div key={i} className="flex items-center space-x-4">
                    <div className="w-12 h-12 bg-gray-200 rounded-full"></div>
                    <div className="flex-1">
                      <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
                      <div className="h-3 bg-gray-200 rounded w-1/2"></div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 p-6">
      <div className="max-w-4xl mx-auto">
        <div className="bg-white rounded-lg shadow-sm">
          {/* Header */}
          <div className="px-6 py-4 border-b border-gray-200">
            <div className="flex items-center justify-between mb-4">
              <button
                onClick={() => navigate('/dashboard')}
                className="flex items-center text-gray-600 hover:text-gray-900 transition-colors"
              >
                <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                </svg>
                Back to Dashboard
              </button>
            </div>
            <div className="flex items-center justify-between">
              <div>
                <h1 className="text-2xl font-bold text-gray-900">Notifications</h1>
                <p className="text-gray-600 mt-1">
                  {totalNotifications} total notification{totalNotifications !== 1 ? 's' : ''}
                </p>
              </div>
              <button
                onClick={() => loadNotifications(page)}
                disabled={loading}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
              >
                {loading ? 'Refreshing...' : 'Refresh'}
              </button>
            </div>
          </div>

          {/* Error State */}
          {error && (
            <div className="px-6 py-4 bg-red-50 border-l-4 border-red-400">
              <p className="text-red-700">{error}</p>
              <button
                onClick={() => loadNotifications(page)}
                className="mt-2 text-red-600 hover:text-red-800 underline"
              >
                Try again
              </button>
            </div>
          )}

          {/* Notifications List */}
          <div className="divide-y divide-gray-200">
            {notifications.length === 0 && !loading ? (
              <div className="px-6 py-12 text-center">
                <div className="text-6xl mb-4">📩</div>
                <h3 className="text-lg font-medium text-gray-900 mb-2">No notifications yet</h3>
                <p className="text-gray-600">
                  You'll see court availability alerts and other notifications here.
                </p>
              </div>
            ) : (
              notifications.map((notification) => (
                <div key={notification.id} className="px-6 py-4 hover:bg-gray-50 transition-colors">
                  <div className="flex items-start space-x-4">
                    <div className="text-2xl">{getNotificationIcon(notification.type)}</div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between mb-2">
                        <h3 className="text-lg font-semibold text-gray-900">
                          {notification.venueName}
                        </h3>
                        {getStatusBadge(notification.emailSent, notification.emailStatus)}
                      </div>
                      <div className="space-y-1">
                        <p className="text-gray-700">
                          <span className="font-medium">{notification.courtName}</span>
                          {' • '}
                          <span>{formatDate(notification.date)}</span>
                          {' • '}
                          <span>{formatTime(notification.time)}</span>
                          {notification.price > 0 && (
                            <>
                              {' • '}
                              <span className="font-medium text-green-600">£{notification.price.toFixed(2)}</span>
                            </>
                          )}
                        </p>
                        <p className="text-sm text-gray-500">
                          {formatDate(notification.createdAt)} at{' '}
                          {new Date(notification.createdAt).toLocaleTimeString('en-US', {
                            hour: 'numeric',
                            minute: '2-digit',
                            hour12: true,
                          })}
                        </p>
                      </div>
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="px-6 py-4 border-t border-gray-200">
              <div className="flex items-center justify-between">
                <p className="text-sm text-gray-700">
                  Page {page + 1} of {totalPages}
                </p>
                <div className="flex space-x-2">
                  <button
                    onClick={() => handlePageChange(page - 1)}
                    disabled={page === 0 || loading}
                    className="px-3 py-1 text-sm bg-gray-100 text-gray-700 rounded hover:bg-gray-200 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                  >
                    Previous
                  </button>
                  <button
                    onClick={() => handlePageChange(page + 1)}
                    disabled={page >= totalPages - 1 || loading}
                    className="px-3 py-1 text-sm bg-gray-100 text-gray-700 rounded hover:bg-gray-200 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                  >
                    Next
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default Notifications