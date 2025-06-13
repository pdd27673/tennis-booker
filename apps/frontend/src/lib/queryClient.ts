import { QueryClient } from '@tanstack/react-query'

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // Time in milliseconds that unused/inactive cache data remains in memory
      gcTime: 1000 * 60 * 5, // 5 minutes
      // Time in milliseconds after data is considered stale
      staleTime: 1000 * 60 * 1, // 1 minute
      // Number of times to retry failed requests
      retry: 3,
      // Retry delay function
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
      // Refetch on window focus
      refetchOnWindowFocus: false,
      // Refetch on reconnect
      refetchOnReconnect: true,
    },
    mutations: {
      // Number of times to retry failed mutations
      retry: 1,
      // Retry delay for mutations
      retryDelay: 1000,
    },
  },
}) 