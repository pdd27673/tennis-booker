import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useAuth } from '@/hooks/useAuth'
import { tokenStorage } from '@/lib/tokenStorage'
import { useNavigate } from 'react-router-dom'

export function AuthStatus() {
  const { 
    isAuthenticated, 
    user, 
    accessToken, 
    refreshToken,
    logout,
    refreshUserInfo
  } = useAuth()
  const navigate = useNavigate()

  const handleLogout = () => {
    logout()
  }

  const handleCheckStorage = () => {
    const storageTokens = {
      accessToken: tokenStorage.getAccessToken(),
      refreshToken: tokenStorage.getRefreshToken(),
    }
    console.log('Current tokens in storage:', storageTokens)
  }

  const handleGoToDashboard = () => {
    navigate('/dashboard')
  }

  const handleRefreshUserInfo = async () => {
    const result = await refreshUserInfo()
    if (result.success) {
      console.log('User information refreshed successfully')
    } else {
      console.error('Failed to refresh user info:', result.error)
    }
  }

  return (
    <Card className="w-full max-w-2xl">
      <CardHeader>
        <CardTitle>Authentication Status</CardTitle>
        <CardDescription>Current authentication state and token information</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Authentication Status */}
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="text-sm font-medium">Status:</label>
            <p className={`text-sm ${isAuthenticated ? 'text-green-600' : 'text-red-600'}`}>
              {isAuthenticated ? '✅ Authenticated' : '❌ Not Authenticated'}
            </p>
          </div>
          <div>
            <label className="text-sm font-medium">User:</label>
            <p className="text-sm">
              {user ? user.name : 'None'}
            </p>
          </div>
        </div>

        {/* User Information */}
        {user && (
          <div className="p-3 bg-gray-50 dark:bg-gray-800 rounded">
            <h4 className="font-medium mb-2">User Profile</h4>
            <div className="text-sm space-y-1">
              <p><strong>ID:</strong> {user.id}</p>
              <p><strong>Name:</strong> {user.name}</p>
              <p><strong>Email:</strong> {user.email}</p>
            </div>
          </div>
        )}

        {/* Token Information */}
        <div className="space-y-2">
          <h4 className="font-medium">Token Status</h4>
          <div className="grid grid-cols-1 gap-2 text-sm">
            <div className="p-2 bg-gray-50 dark:bg-gray-800 rounded">
              <strong>Access Token (Memory):</strong>
              <p className="font-mono text-xs break-all">
                {accessToken ? `${accessToken.substring(0, 40)}...` : 'None'}
              </p>
            </div>
            <div className="p-2 bg-gray-50 dark:bg-gray-800 rounded">
              <strong>Refresh Token (localStorage):</strong>
              <p className="font-mono text-xs break-all">
                {refreshToken ? `${refreshToken.substring(0, 40)}...` : 'None'}
              </p>
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="flex flex-wrap gap-2">
          {isAuthenticated && (
            <>
              <Button onClick={handleGoToDashboard} variant="default" size="sm">
                Go to Dashboard
              </Button>
              <Button onClick={handleRefreshUserInfo} variant="outline" size="sm">
                Refresh User Info
              </Button>
              <Button onClick={handleLogout} variant="destructive" size="sm">
                Logout
              </Button>
            </>
          )}
          <Button onClick={handleCheckStorage} variant="outline" size="sm">
            Check Storage (Console)
          </Button>
        </div>

        {/* Storage Info */}
        <div className="text-xs text-gray-600 dark:text-gray-400 p-3 bg-blue-50 dark:bg-blue-900/20 rounded">
          <p><strong>Storage Strategy:</strong></p>
          <p>• Access Token: Stored in memory (lost on page refresh)</p>
          <p>• Refresh Token: Stored in localStorage (persists across sessions)</p>
          <p>• User Profile: Stored in Zustand store (not persisted)</p>
        </div>
      </CardContent>
    </Card>
  )
} 