import axios, { AxiosError } from 'axios'
import config from '@/config/config'
import { tokenStorage } from '@/lib/tokenStorage'

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

// Create axios instance with default config
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Add auth token to requests
apiClient.interceptors.request.use((config) => {
  const token = tokenStorage.getAccessToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Types for system API responses (matching actual backend structure)
export interface SystemStatusResponse {
  status: string
  scrapingStatus: string
  lastUpdate: string
  activeJobs: number
  queuedJobs: number
  completedJobs: number
  erroredJobs: number
  systemHealth: string
  message: string
}

export interface SystemControlResponse {
  success: boolean
  message: string
  status: string
}

export interface HealthResponse {
  status: string
  timestamp: string
  version?: string
}

export interface ScrapingLog {
  id: string
  venueId: string
  venueName: string
  provider: string
  platform: string
  scrapeTimestamp: string
  success: boolean
  slotsFound: number
  scrapeDurationMs: number
  errors: string[] | null
  createdAt: string
}

// Transform backend system status to match frontend expectations
const transformSystemStatus = (backendStatus: SystemStatusResponse) => {
  const statusMap: { [key: string]: 'RUNNING' | 'PAUSED' | 'ERROR' } = {
    'running': 'RUNNING',
    'paused': 'PAUSED',
    'restarting': 'RUNNING',
    'error': 'ERROR'
  }

  return {
    systemStatus: statusMap[backendStatus.status] || 'ERROR',
    lastUpdate: new Date(backendStatus.lastUpdate),
    isScrapingActive: backendStatus.scrapingStatus === 'active',
    systemInfo: {
      monitoredClubs: 12, // This could be fetched from venues API
      lastScan: new Date(backendStatus.lastUpdate),
      nextScan: new Date(Date.now() + 30 * 60 * 1000), // 30 minutes from now
      averageResponseTime: 1.2, // This could be enhanced with real metrics
      successRate: backendStatus.erroredJobs > 0 ? 
        100 - (backendStatus.erroredJobs / Math.max(backendStatus.completedJobs + backendStatus.erroredJobs, 1)) * 100 : 
        100,
      courtsFoundToday: backendStatus.completedJobs,
      activeJobs: backendStatus.activeJobs,
      queuedJobs: backendStatus.queuedJobs,
      completedJobs: backendStatus.completedJobs,
      erroredJobs: backendStatus.erroredJobs,
      systemHealth: backendStatus.systemHealth,
      message: backendStatus.message,
    },
  }
}

// Handle API errors consistently
const handleSystemError = (error: AxiosError) => {
  if (error.response) {
    const errorData = error.response.data as any
    throw new Error(errorData.message || errorData.error || 'System operation failed')
  } else if (error.request) {
    throw new Error('Network error. Please check your connection.')
  } else {
    throw new Error('An unexpected error occurred.')
  }
}

export interface SystemStatus {
  status: string
  scrapingStatus: string
  lastUpdate: string
  lastScrapeTime?: string
  activeJobs: number
  queuedJobs: number
  completedJobs: number
  erroredJobs: number
  systemHealth: string
  message: string
}

export interface SystemMetrics {
  uptime: string
  lastScrapeTime: string
  notificationsSent: number
  scraperHealth: string
}

export const systemApi = {
  // Get system health status (public endpoint, no auth required)
  async getHealth(): Promise<HealthResponse> {
    try {
      const response = await axios.get(`${API_BASE_URL}/api/health`)
      return response.data
    } catch (error) {
      handleSystemError(error as AxiosError)
      throw error
    }
  },

  // Get system status
  async getSystemStatus(): Promise<SystemStatus> {
    try {
      const response = await apiClient.get('/api/system/status')
      return response.data
    } catch (error) {
      console.error('Failed to fetch system status:', error)
      throw error
    }
  },

  // Get system metrics (derived from status and other sources)
  async getSystemMetrics(): Promise<SystemMetrics> {
    try {
      const status = await this.getSystemStatus()
      
      // Calculate uptime based on last update time
      const formatUptime = (lastUpdateStr: string): string => {
        try {
          const lastUpdate = new Date(lastUpdateStr)
          const now = new Date()
          const diffMs = now.getTime() - lastUpdate.getTime()
          const diffMins = Math.floor(diffMs / (1000 * 60))
          const diffHours = Math.floor(diffMins / 60)
          const diffDays = Math.floor(diffHours / 24)
          
          if (diffDays > 0) return `${diffDays} day${diffDays > 1 ? 's' : ''}`
          if (diffHours > 0) return `${diffHours} hour${diffHours > 1 ? 's' : ''}`
          if (diffMins > 0) return `${diffMins} minute${diffMins > 1 ? 's' : ''}`
          return 'Just started'
        } catch (error) {
          return 'Unknown'
        }
      }
      
      // Format last scrape time
      const formatLastScrape = (lastScrapeStr?: string): string => {
        if (!lastScrapeStr) return 'Never'
        try {
          const lastScrape = new Date(lastScrapeStr)
          const now = new Date()
          const diffMs = now.getTime() - lastScrape.getTime()
          const diffMins = Math.floor(diffMs / (1000 * 60))
          
          if (diffMins < 1) return 'Just now'
          if (diffMins < 60) return `${diffMins} minute${diffMins > 1 ? 's' : ''} ago`
          
          const diffHours = Math.floor(diffMins / 60)
          if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`
          
          return lastScrape.toLocaleDateString()
        } catch (error) {
          return 'Never'
        }
      }
      
      // Determine scraper health based on status and recent activity
      const getScraperHealth = (): string => {
        if (status.scrapingStatus === 'active') return 'Excellent'
        if (status.scrapingStatus === 'paused') return 'Paused'
        if (status.scrapingStatus === 'error') return 'Error'
        if (status.systemHealth === 'healthy') return 'Good'
        return 'Unknown'
      }
      
      return {
        uptime: formatUptime(status.lastUpdate),
        lastScrapeTime: formatLastScrape(status.lastScrapeTime),
        notificationsSent: status.completedJobs || 0,
        scraperHealth: getScraperHealth()
      }
    } catch (error) {
      console.error('Failed to fetch system metrics:', error)
      return {
        uptime: 'Unknown',
        lastScrapeTime: 'Never',
        notificationsSent: 0,
        scraperHealth: 'Unknown'
      }
    }
  },

  // Get scraping logs
  async getScrapingLogs(limit: number = 50): Promise<ScrapingLog[]> {
    try {
      const response = await apiClient.get(`/api/system/logs?limit=${limit}`)
      return response.data || []
    } catch (error) {
      console.error('Failed to fetch scraping logs:', error)
      return []
    }
  },

  // System control actions
  async pauseScraping(): Promise<{ status: string; message: string }> {
    try {
      const response = await apiClient.post('/api/system/pause')
      return response.data
    } catch (error) {
      console.error('Failed to pause scraping:', error)
      throw error
    }
  },

  async resumeScraping(): Promise<{ status: string; message: string }> {
    try {
      const response = await apiClient.post('/api/system/resume')
      return response.data
    } catch (error) {
      console.error('Failed to resume scraping:', error)
      throw error
    }
  },

  async restartSystem(): Promise<{ status: string; message: string }> {
    try {
      const response = await apiClient.post('/api/system/restart')
      return response.data
    } catch (error) {
      console.error('Failed to restart system:', error)
      throw error
    }
  }
}

export default systemApi 