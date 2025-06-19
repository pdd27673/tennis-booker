import axios, { AxiosError } from 'axios'
import config from '@/config/config'
import { tokenStorage } from '@/lib/tokenStorage'

// Create axios instance for court API calls
const courtApiClient = axios.create({
  baseURL: config.apiUrl,
  timeout: config.apiTimeout,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Add request interceptor to include auth token
courtApiClient.interceptors.request.use(
  (config) => {
    const token = tokenStorage.getAccessToken()

    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Types for API responses
export interface Venue {
  id: string
  name: string
  provider: string
  url: string
  location: {
    address: string
    city: string
    post_code: string
    latitude?: number
    longitude?: number
  }
  courts: Court[]
  booking_window?: number
  created_at: string
  updated_at: string
  last_scraped_at?: string
  scraping_interval?: number
  is_active: boolean
}

export interface Court {
  id: string
  name: string
  surface?: string
  indoor: boolean
  floodlights?: boolean
  court_type?: string
  tags?: string[]
}

export interface CourtSlot {
  id: string
  venueId: string
  venueName: string
  courtId: string
  courtName: string
  date: string
  startTime: string
  endTime: string
  duration?: number
  price: number
  currency: string
  available: boolean
  bookingUrl: string
  platform: string
  createdAt: string
  updatedAt: string
}

export interface CourtSlotFilter {
  venueId?: string
  date?: string
  startTime?: string
  endTime?: string
  provider?: string
  minPrice?: number
  maxPrice?: number
  limit?: number
}

// Transform court slot to match frontend CourtCard props
export const transformCourtSlotToCardProps = (slot: CourtSlot) => {
  const formatTimeSlot = () => {
    // Handle invalid or missing date
    if (!slot.date) {
      return 'Date TBD'
    }
    
    let date: Date
    try {
      date = new Date(slot.date)
      // Check if date is valid
      if (isNaN(date.getTime())) {
        return 'Invalid Date'
      }
    } catch (error) {
      return 'Invalid Date'
    }
    
    const today = new Date()
    const tomorrow = new Date(today)
    tomorrow.setDate(tomorrow.getDate() + 1)
    
    let datePrefix = ''
    if (date.toDateString() === today.toDateString()) {
      datePrefix = 'Today '
    } else if (date.toDateString() === tomorrow.toDateString()) {
      datePrefix = 'Tomorrow '
    } else {
      datePrefix = date.toLocaleDateString() + ' '
    }
    
    // Convert 24h time to 12h format
    const formatTime = (time: string | undefined | null) => {
      if (!time || typeof time !== 'string') {
        return '12:00 PM' // Default fallback
      }
      
      const timeParts = time.split(':')
      if (timeParts.length < 2) {
        return '12:00 PM' // Default fallback for invalid format
      }
      
      const [hours, minutes] = timeParts
      const hour = parseInt(hours || '0', 10)
      const ampm = hour >= 12 ? 'PM' : 'AM'
      const hour12 = hour % 12 || 12
      return `${hour12}:${minutes || '00'} ${ampm}`
    }
    
    return `${datePrefix}${formatTime(slot.startTime)} - ${formatTime(slot.endTime)}`
  }

  const getAvailabilityStatus = () => {
    if (slot.available) {
      return 'available' as const
    }
    return 'booked' as const
  }

  return {
    courtName: slot.courtName || 'Unknown Court',
    courtType: 'Tennis Court', // Default type, could be enhanced based on court data
    venue: slot.venueName || 'Unknown Venue',
    availabilityStatus: getAvailabilityStatus(),
    timeSlot: formatTimeSlot(),
    price: `${slot.currency === 'GBP' ? '£' : '$'}${slot.price || 0}/hour`,
    maxPlayers: 4, // Default value, could be enhanced based on court type
    bookingLink: slot.bookingUrl || '#',
  }
}

// Handle API errors consistently
const handleCourtError = (error: AxiosError) => {
  if (error.response) {
    const errorData = error.response.data as any
    throw new Error(errorData.message || errorData.error || 'API request failed')
  } else if (error.request) {
    throw new Error('Network error. Please check your connection.')
  } else {
    throw new Error('An unexpected error occurred.')
  }
}

export const courtApi = {
  async getVenues(): Promise<Venue[]> {
    try {
      const response = await courtApiClient.get('/api/venues')
      return response.data || []
    } catch (error) {

      handleCourtError(error as AxiosError)
      return []
    }
  },

  async getCourtSlots(filter?: CourtSlotFilter): Promise<CourtSlot[]> {
    try {
      const params = new URLSearchParams()
      
      if (filter) {
        if (filter.venueId) params.append('venueId', filter.venueId)
        if (filter.date) params.append('date', filter.date)
        if (filter.startTime) params.append('startTime', filter.startTime)
        if (filter.endTime) params.append('endTime', filter.endTime)
        if (filter.provider) params.append('provider', filter.provider)
        if (filter.minPrice !== undefined) params.append('minPrice', filter.minPrice.toString())
        if (filter.maxPrice !== undefined) params.append('maxPrice', filter.maxPrice.toString())
        if (filter.limit) params.append('limit', filter.limit.toString())
      }
      
      const url = `/api/courts${params.toString() ? `?${params.toString()}` : ''}`
      const response = await courtApiClient.get(url)
      return response.data || []
    } catch (error) {
      handleCourtError(error as AxiosError)
      return []
    }
  },

  async getAvailableSlots(limit: number = 100): Promise<CourtSlot[]> {
    return this.getCourtSlots({ limit })
  },

  async getTodaySlots(): Promise<CourtSlot[]> {
    const today = new Date().toISOString().split('T')[0] // YYYY-MM-DD format
    return this.getCourtSlots({ date: today })
  },

  async getSlotsByVenue(venueId: string, limit: number = 50): Promise<CourtSlot[]> {
    return this.getCourtSlots({ venueId, limit })
  },

  // Get court slots transformed for CourtCard components
  async getCourtSlotsForCards(filter?: CourtSlotFilter) {
    const slots = await this.getCourtSlots(filter)
    return slots.map(transformCourtSlotToCardProps)
  },

  // Get dashboard statistics
  async getDashboardStats() {
    try {
      const response = await courtApiClient.get('/api/dashboard/stats')
      
      const backendStats = response.data
      
      // Transform backend stats to match frontend expectations
      const stats = {
        totalVenues: backendStats.totalVenues || 0,
        totalCourts: backendStats.totalCourtSlots || 0, // Backend uses totalCourtSlots
        activeCourts: Math.max(0, (backendStats.totalCourtSlots || 0) - (backendStats.availableSlots || 0)), // Calculate active courts
        availableSlots: backendStats.availableSlots || 0,
        systemStatus: 'RUNNING' as const,
        lastUpdate: new Date(),
      }


      return stats
    } catch (error) {

      return {
        totalVenues: 0,
        totalCourts: 0,
        activeCourts: 0,
        availableSlots: 0,
        systemStatus: 'ERROR' as const,
        lastUpdate: new Date(),
      }
    }
  },
}

export default courtApi 