import React, { useState, useEffect, useRef } from 'react'
import { systemApi, type ScrapingLog } from '@/services/systemApi'

interface ScrapingLogsTerminalProps {
  className?: string
}

const ScrapingLogsTerminal: React.FC<ScrapingLogsTerminalProps> = ({ className = '' }) => {
  const [logs, setLogs] = useState<ScrapingLog[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [autoRefresh, setAutoRefresh] = useState(true)
  const [filter, setFilter] = useState<'all' | 'success' | 'error'>('all')
  const [isUserScrolledUp, setIsUserScrolledUp] = useState(false)
  const terminalRef = useRef<HTMLDivElement>(null)
  const intervalRef = useRef<NodeJS.Timeout | null>(null)

  const fetchLogs = async () => {
    try {
      setError(null)
      const fetchedLogs = await systemApi.getScrapingLogs(100)
      setLogs(fetchedLogs)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch logs')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchLogs()
  }, [])

  useEffect(() => {
    if (autoRefresh) {
      intervalRef.current = setInterval(fetchLogs, 10000) // Refresh every 10 seconds
    } else {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
        intervalRef.current = null
      }
    }

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
      }
    }
  }, [autoRefresh])

  // Handle scroll detection to determine if user has scrolled up
  const handleScroll = () => {
    if (terminalRef.current) {
      const { scrollTop, scrollHeight, clientHeight } = terminalRef.current
      const isAtBottom = scrollTop + clientHeight >= scrollHeight - 10 // 10px threshold
      setIsUserScrolledUp(!isAtBottom)
    }
  }

  // Auto-scroll to bottom when new logs arrive, but only if user hasn't scrolled up
  useEffect(() => {
    if (terminalRef.current && !isUserScrolledUp) {
      terminalRef.current.scrollTop = terminalRef.current.scrollHeight
    }
  }, [logs, isUserScrolledUp])

  const filteredLogs = logs.filter(log => {
    if (filter === 'success') return log.success
    if (filter === 'error') return !log.success
    return true
  })

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString('en-GB', {
      day: '2-digit',
      month: '2-digit',
      year: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })
  }

  const formatDuration = (ms: number) => {
    if (ms < 1000) return `${ms}ms`
    return `${(ms / 1000).toFixed(1)}s`
  }

  const getStatusIcon = (success: boolean) => {
    return success ? '✅' : '❌'
  }

  const getStatusColor = (success: boolean) => {
    return success ? 'text-green-400' : 'text-red-400'
  }

  return (
    <div className={`bg-gray-900 rounded-lg border border-gray-700 ${className}`}>
      {/* Terminal Header */}
      <div className="flex items-center justify-between px-4 py-3 bg-gray-800 rounded-t-lg border-b border-gray-700">
        <div className="flex items-center space-x-3">
          <div className="flex space-x-1">
            <div className="w-3 h-3 bg-red-500 rounded-full"></div>
            <div className="w-3 h-3 bg-yellow-500 rounded-full"></div>
            <div className="w-3 h-3 bg-green-500 rounded-full"></div>
          </div>
          <span className="text-gray-300 font-mono text-sm">tennis-scraper@logs:~$</span>
        </div>
        
        <div className="flex items-center space-x-4">
          {/* Filter Controls */}
          <div className="flex items-center space-x-2">
            <span className="text-gray-400 text-xs">Filter:</span>
            <select
              value={filter}
              onChange={(e) => setFilter(e.target.value as 'all' | 'success' | 'error')}
              className="bg-gray-700 text-gray-300 text-xs px-2 py-1 rounded border border-gray-600 focus:outline-none focus:border-blue-500"
            >
              <option value="all">All</option>
              <option value="success">Success</option>
              <option value="error">Errors</option>
            </select>
          </div>

          {/* Auto-refresh Toggle */}
          <div className="flex items-center space-x-2">
            <span className="text-gray-400 text-xs">Auto-refresh:</span>
            <button
              onClick={() => setAutoRefresh(!autoRefresh)}
              className={`text-xs px-2 py-1 rounded ${
                autoRefresh 
                  ? 'bg-green-600 text-white' 
                  : 'bg-gray-600 text-gray-300'
              }`}
            >
              {autoRefresh ? 'ON' : 'OFF'}
            </button>
          </div>

          {/* Manual Refresh */}
          <button
            onClick={fetchLogs}
            disabled={loading}
            className="text-gray-400 hover:text-white text-xs px-2 py-1 rounded bg-gray-700 hover:bg-gray-600 disabled:opacity-50"
          >
            {loading ? '⟳' : '↻'}
          </button>

          {/* Scroll to Bottom Button (only show when user has scrolled up) */}
          {isUserScrolledUp && (
            <button
              onClick={() => {
                if (terminalRef.current) {
                  terminalRef.current.scrollTop = terminalRef.current.scrollHeight
                  setIsUserScrolledUp(false)
                }
              }}
              className="text-blue-400 hover:text-blue-300 text-xs px-2 py-1 rounded bg-blue-900/30 hover:bg-blue-900/50 border border-blue-600/30"
            >
              ↓ Bottom
            </button>
          )}
        </div>
      </div>

      {/* Terminal Content */}
      <div 
        ref={terminalRef}
        onScroll={handleScroll}
        className="h-96 overflow-y-auto p-4 font-mono text-sm bg-gray-900 text-gray-300"
      >
        {loading && logs.length === 0 ? (
          <div className="text-yellow-400">Loading scraping logs...</div>
        ) : error ? (
          <div className="text-red-400">Error: {error}</div>
        ) : filteredLogs.length === 0 ? (
          <div className="text-gray-500">No logs found matching current filter.</div>
        ) : (
          <div className="space-y-1">
            {filteredLogs.map((log) => (
              <div key={log.id} className="border-l-2 border-gray-700 pl-3 py-1">
                <div className="flex items-start space-x-2">
                  <span className="text-gray-500 text-xs min-w-[120px]">
                    {formatTimestamp(log.scrapeTimestamp)}
                  </span>
                  <span className={`${getStatusColor(log.success)} min-w-[20px]`}>
                    {getStatusIcon(log.success)}
                  </span>
                  <span className="text-blue-400 min-w-[80px] truncate">
                    {log.platform}
                  </span>
                  <span className="text-gray-300 flex-1 truncate">
                    {log.venueName}
                  </span>
                  <span className="text-gray-500 text-xs min-w-[60px]">
                    {formatDuration(log.scrapeDurationMs)}
                  </span>
                  {log.success ? (
                    <span className="text-green-400 text-xs min-w-[80px]">
                      {log.slotsFound} slots
                    </span>
                  ) : (
                    <span className="text-red-400 text-xs min-w-[80px]">
                      Failed
                    </span>
                  )}
                </div>
                
                {/* Show errors if any */}
                {!log.success && log.errors && log.errors.length > 0 && (
                  <div className="ml-[142px] mt-1">
                    {log.errors.map((error, index) => (
                      <div key={index} className="text-red-300 text-xs opacity-80">
                        └─ {error}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Terminal Footer */}
      <div className="px-4 py-2 bg-gray-800 rounded-b-lg border-t border-gray-700 flex items-center justify-between">
        <div className="text-gray-400 text-xs">
          Showing {filteredLogs.length} of {logs.length} logs
        </div>
        <div className="text-gray-400 text-xs">
          {autoRefresh && 'Auto-refreshing every 10s'}
        </div>
      </div>
    </div>
  )
}

export default ScrapingLogsTerminal 