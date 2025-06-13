import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Clock, MapPin, Users } from 'lucide-react'

export interface CourtCardProps {
  courtName: string
  courtType: string
  venue: string
  availabilityStatus: 'available' | 'booked' | 'maintenance' | 'pending'
  timeSlot: string
  price?: string
  maxPlayers?: number
  bookingLink: string
  className?: string
}

const statusConfig = {
  available: {
    badge: 'bg-green-100 text-green-800 border-green-200',
    label: 'Available',
    buttonVariant: 'default' as const,
    buttonText: 'Book Now'
  },
  booked: {
    badge: 'bg-red-100 text-red-800 border-red-200',
    label: 'Booked',
    buttonVariant: 'secondary' as const,
    buttonText: 'View Details'
  },
  maintenance: {
    badge: 'bg-yellow-100 text-yellow-800 border-yellow-200',
    label: 'Maintenance',
    buttonVariant: 'outline' as const,
    buttonText: 'Notify Me'
  },
  pending: {
    badge: 'bg-blue-100 text-blue-800 border-blue-200',
    label: 'Pending',
    buttonVariant: 'outline' as const,
    buttonText: 'Join Waitlist'
  }
}

export function CourtCard({
  courtName,
  courtType,
  venue,
  availabilityStatus,
  timeSlot,
  price,
  maxPlayers,
  bookingLink,
  className
}: CourtCardProps) {
  const config = statusConfig[availabilityStatus]
  
  const handleBookingClick = () => {
    if (bookingLink.startsWith('http')) {
      window.open(bookingLink, '_blank', 'noopener,noreferrer')
    } else {
      // For placeholder links, show an alert
      alert(`Booking link: ${bookingLink}\nCourt: ${courtName}\nTime: ${timeSlot}`)
    }
  }

  return (
    <Card className={`hover:shadow-lg transition-shadow duration-200 ${className}`}>
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <CardTitle className="text-lg font-semibold text-gray-900 dark:text-gray-100">
              {courtName}
            </CardTitle>
            <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
              {courtType}
            </p>
          </div>
          <Badge 
            variant="outline" 
            className={config.badge}
          >
            {config.label}
          </Badge>
        </div>
      </CardHeader>

      <CardContent className="space-y-3">
        {/* Venue */}
        <div className="flex items-center space-x-2 text-sm text-gray-600 dark:text-gray-400">
          <MapPin className="w-4 h-4" />
          <span>{venue}</span>
        </div>

        {/* Time Slot */}
        <div className="flex items-center space-x-2 text-sm text-gray-600 dark:text-gray-400">
          <Clock className="w-4 h-4" />
          <span>{timeSlot}</span>
        </div>

        {/* Max Players (if provided) */}
        {maxPlayers && (
          <div className="flex items-center space-x-2 text-sm text-gray-600 dark:text-gray-400">
            <Users className="w-4 h-4" />
            <span>Up to {maxPlayers} players</span>
          </div>
        )}

        {/* Price (if provided) */}
        {price && (
          <div className="flex items-center justify-between pt-2 border-t border-gray-200 dark:border-gray-700">
            <span className="text-sm text-gray-600 dark:text-gray-400">Price:</span>
            <span className="text-lg font-semibold text-gray-900 dark:text-gray-100">
              {price}
            </span>
          </div>
        )}
      </CardContent>

      <CardFooter className="pt-3">
        <Button 
          variant={config.buttonVariant}
          className="w-full"
          onClick={handleBookingClick}
          disabled={availabilityStatus === 'maintenance'}
        >
          {config.buttonText}
        </Button>
      </CardFooter>
    </Card>
  )
}

export default CourtCard 