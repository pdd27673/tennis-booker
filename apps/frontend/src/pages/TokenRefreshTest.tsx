import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useAuth } from '@/hooks/useAuth'
import { 
  testProtectedEndpoint, 
  testExpiredTokenEndpoint,
  apiClient
} from '@/services/tokenRefreshService'
import { mockAuthApi } from '@/services/mockAuthApi'
import { tokenStorage } from '@/lib/tokenStorage'

export default function TokenRefreshTest() {
  const { isAuthenticated, user, accessToken } = useAuth()
  const [testResults, setTestResults] = useState<Array<{
    test: string
    result: string
    success: boolean
    timestamp: Date
  }>>([])
  const [isLoading, setIsLoading] = useState(false)

  const addTestResult = (test: string, result: string, success: boolean) => {
    setTestResults(prev => [...prev, {
      test,
      result,
      success,
      timestamp: new Date()
    }])
  }

  const clearResults = () => {
    setTestResults([])
  }

  // Test 1: Call protected endpoint with valid token
  const testValidToken = async () => {
    setIsLoading(true)
    try {
      const response = await testProtectedEndpoint()
      addTestResult(
        'Valid Token Test',
        response.success ? `Success: ${JSON.stringify(response.data)}` : `Failed: ${response.error}`,
        response.success
      )
    } catch (error) {
      addTestResult(
        'Valid Token Test',
        `Error: ${error instanceof Error ? error.message : 'Unknown error'}`,
        false
      )
    } finally {
      setIsLoading(false)
    }
  }

  // Test 2: Call endpoint that always returns 401 (simulates expired token)
  const testExpiredToken = async () => {
    setIsLoading(true)
    try {
      const response = await testExpiredTokenEndpoint()
      addTestResult(
        'Expired Token Test',
        response.success ? `Unexpected success: ${JSON.stringify(response.data)}` : `Expected failure: ${response.error}`,
        !response.success // We expect this to fail
      )
    } catch (error) {
      addTestResult(
        'Expired Token Test',
        `Error: ${error instanceof Error ? error.message : 'Unknown error'}`,
        false
      )
    } finally {
      setIsLoading(false)
    }
  }

  // Test 3: Simulate token refresh by calling the refresh endpoint directly
  const testTokenRefresh = async () => {
    setIsLoading(true)
    try {
      const refreshToken = tokenStorage.getRefreshToken()
      if (!refreshToken) {
        addTestResult('Token Refresh Test', 'No refresh token available', false)
        return
      }

      const response = await mockAuthApi.refreshTokenEndpoint(refreshToken)
      addTestResult(
        'Token Refresh Test',
        response.success 
          ? `Success: New tokens generated` 
          : `Failed: ${response.error}`,
        response.success
      )
    } catch (error) {
      addTestResult(
        'Token Refresh Test',
        `Error: ${error instanceof Error ? error.message : 'Unknown error'}`,
        false
      )
    } finally {
      setIsLoading(false)
    }
  }

  // Test 4: Test axios interceptor with a mock 401 response
  const testAxiosInterceptor = async () => {
    setIsLoading(true)
    try {
      // This will trigger a 401 response and should automatically refresh the token
      const response = await apiClient.get('/mock/protected-data')
      addTestResult(
        'Axios Interceptor Test',
        `Success: ${JSON.stringify(response.data)}`,
        true
      )
    } catch (error) {
      addTestResult(
        'Axios Interceptor Test',
        `Error: ${error instanceof Error ? error.message : 'Request failed'}`,
        false
      )
    } finally {
      setIsLoading(false)
    }
  }

  // Test 5: Manually expire current token and test refresh
  const testManualTokenExpiry = async () => {
    setIsLoading(true)
    try {
      // Set an obviously expired token
      const expiredToken = `access_${user?.id}_1000000000` // Very old timestamp
      tokenStorage.setAccessToken(expiredToken)
      
      // Try to call protected endpoint - should trigger refresh
      const response = await testProtectedEndpoint()
      addTestResult(
        'Manual Token Expiry Test',
        response.success 
          ? `Success: Token was refreshed automatically` 
          : `Failed: ${response.error}`,
        response.success
      )
    } catch (error) {
      addTestResult(
        'Manual Token Expiry Test',
        `Error: ${error instanceof Error ? error.message : 'Unknown error'}`,
        false
      )
    } finally {
      setIsLoading(false)
    }
  }

  if (!isAuthenticated) {
    return (
      <div className="container mx-auto p-6">
        <Card>
          <CardHeader>
            <CardTitle>Token Refresh Test</CardTitle>
            <CardDescription>
              Please log in to test token refresh functionality
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-gray-600">You need to be authenticated to run these tests.</p>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Token Refresh Test Suite</CardTitle>
          <CardDescription>
            Test various scenarios for automatic token refresh functionality
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Current Auth Status */}
          <div className="p-4 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg">
            <h3 className="font-semibold text-blue-900 dark:text-blue-100 mb-2">
              Current Authentication Status
            </h3>
            <div className="text-sm text-blue-700 dark:text-blue-300 space-y-1">
              <p><strong>User:</strong> {user?.name} ({user?.email})</p>
              <p><strong>Access Token:</strong> {accessToken ? `${accessToken.substring(0, 30)}...` : 'None'}</p>
              <p><strong>Authenticated:</strong> {isAuthenticated ? 'Yes' : 'No'}</p>
            </div>
          </div>

          {/* Test Buttons */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <Button 
              onClick={testValidToken} 
              disabled={isLoading}
              variant="default"
            >
              Test Valid Token
            </Button>
            
            <Button 
              onClick={testExpiredToken} 
              disabled={isLoading}
              variant="outline"
            >
              Test Expired Token
            </Button>
            
            <Button 
              onClick={testTokenRefresh} 
              disabled={isLoading}
              variant="outline"
            >
              Test Token Refresh
            </Button>
            
            <Button 
              onClick={testAxiosInterceptor} 
              disabled={isLoading}
              variant="outline"
            >
              Test Axios Interceptor
            </Button>
            
            <Button 
              onClick={testManualTokenExpiry} 
              disabled={isLoading}
              variant="destructive"
            >
              Test Manual Expiry
            </Button>
            
            <Button 
              onClick={clearResults} 
              disabled={isLoading}
              variant="secondary"
            >
              Clear Results
            </Button>
          </div>

          {/* Test Results */}
          {testResults.length > 0 && (
            <div className="space-y-3">
              <h3 className="text-lg font-semibold">Test Results</h3>
              <div className="space-y-2 max-h-96 overflow-y-auto">
                {testResults.map((result, index) => (
                  <div 
                    key={index}
                    className={`p-3 rounded-lg border ${
                      result.success 
                        ? 'bg-green-50 border-green-200 dark:bg-green-900/20 dark:border-green-800' 
                        : 'bg-red-50 border-red-200 dark:bg-red-900/20 dark:border-red-800'
                    }`}
                  >
                    <div className="flex justify-between items-start mb-1">
                      <h4 className={`font-medium ${
                        result.success 
                          ? 'text-green-900 dark:text-green-100' 
                          : 'text-red-900 dark:text-red-100'
                      }`}>
                        {result.test}
                      </h4>
                      <span className={`text-xs ${
                        result.success 
                          ? 'text-green-600 dark:text-green-400' 
                          : 'text-red-600 dark:text-red-400'
                      }`}>
                        {result.success ? '✅' : '❌'}
                      </span>
                    </div>
                    <p className={`text-sm ${
                      result.success 
                        ? 'text-green-700 dark:text-green-300' 
                        : 'text-red-700 dark:text-red-300'
                    }`}>
                      {result.result}
                    </p>
                    <p className="text-xs text-gray-500 mt-1">
                      {result.timestamp.toLocaleTimeString()}
                    </p>
                  </div>
                ))}
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
} 