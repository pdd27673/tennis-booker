import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { 
  Activity, 
  Pause, 
  AlertTriangle, 
  CheckCircle, 
  Clock,
  Zap
} from 'lucide-react'

export type SystemStatus = 'IDLE' | 'RUNNING' | 'PAUSED' | 'ERROR'

interface SystemStatusDisplayProps {
  status: SystemStatus
  lastUpdate?: Date
  className?: string
}

const statusConfig = {
  IDLE: {
    icon: Clock,
    label: 'Idle',
    description: 'System is ready but not actively monitoring',
    badgeVariant: 'secondary' as const,
    bgColor: 'bg-gray-50 dark:bg-gray-800/50',
    iconColor: 'text-gray-500 dark:text-gray-400',
  },
  RUNNING: {
    icon: Activity,
    label: 'Running',
    description: 'Actively monitoring tennis courts',
    badgeVariant: 'default' as const,
    bgColor: 'bg-green-50 dark:bg-green-900/20',
    iconColor: 'text-green-600 dark:text-green-400',
  },
  PAUSED: {
    icon: Pause,
    label: 'Paused',
    description: 'Monitoring temporarily suspended',
    badgeVariant: 'secondary' as const,
    bgColor: 'bg-yellow-50 dark:bg-yellow-900/20',
    iconColor: 'text-yellow-600 dark:text-yellow-400',
  },
  ERROR: {
    icon: AlertTriangle,
    label: 'Error',
    description: 'System encountered an issue',
    badgeVariant: 'destructive' as const,
    bgColor: 'bg-red-50 dark:bg-red-900/20',
    iconColor: 'text-red-600 dark:text-red-400',
  },
}

export default function SystemStatusDisplay({ 
  status, 
  lastUpdate, 
  className = '' 
}: SystemStatusDisplayProps) {
  const config = statusConfig[status]
  const IconComponent = config.icon

  const formatLastUpdate = (date: Date) => {
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMinutes = Math.floor(diffMs / (1000 * 60))
    
    if (diffMinutes < 1) return 'Just now'
    if (diffMinutes < 60) return `${diffMinutes}m ago`
    
    const diffHours = Math.floor(diffMinutes / 60)
    if (diffHours < 24) return `${diffHours}h ago`
    
    const diffDays = Math.floor(diffHours / 24)
    return `${diffDays}d ago`
  }

  return (
    <Card className={`border-0 shadow-sm ${config.bgColor} ${className}`}>
      <CardContent className="p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className={`p-2 rounded-lg ${config.bgColor}`}>
              <IconComponent className={`w-5 h-5 ${config.iconColor}`} />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <h3 className="font-semibold text-gray-900 dark:text-white">
                  System Status
                </h3>
                <Badge variant={config.badgeVariant} className="text-xs">
                  {config.label}
                </Badge>
              </div>
              <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                {config.description}
              </p>
            </div>
          </div>
          
          {/* Status Indicator */}
          <div className="flex items-center space-x-2">
            {status === 'RUNNING' && (
              <div className="flex items-center space-x-1">
                <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
                <Zap className="w-4 h-4 text-green-500" />
              </div>
            )}
            {status === 'ERROR' && (
              <AlertTriangle className="w-5 h-5 text-red-500" />
            )}
            {status === 'PAUSED' && (
              <Pause className="w-5 h-5 text-yellow-500" />
            )}
            {status === 'IDLE' && (
              <CheckCircle className="w-5 h-5 text-gray-500" />
            )}
          </div>
        </div>
        
        {/* Last Update */}
        {lastUpdate && (
          <div className="mt-3 pt-3 border-t border-gray-200 dark:border-gray-700">
            <p className="text-xs text-gray-500 dark:text-gray-400">
              Last updated: {formatLastUpdate(lastUpdate)}
            </p>
          </div>
        )}
      </CardContent>
    </Card>
  )
} 