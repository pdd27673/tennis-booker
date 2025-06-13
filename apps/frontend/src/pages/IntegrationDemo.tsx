import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { BackgroundGradientAnimation } from '@/components/ui/background-gradient-animation'
import { TextGenerateEffect } from '@/components/ui/text-generate-effect'
import { useAppStore } from '@/stores/appStore'
import { usePosts, useCreatePost, type Post } from '@/hooks/useTestApi'

export default function IntegrationDemo() {
  const { 
    theme, 
    setTheme, 
    isAuthenticated, 
    login, 
    logout, 
    userProfile,
    notifications,
    addNotification,
    clearNotifications 
  } = useAppStore()
  
  const { data: posts, isLoading, error, refetch } = usePosts()
  const createPostMutation = useCreatePost()
  const [newPostTitle, setNewPostTitle] = useState('')

  const handleCreatePost = () => {
    if (newPostTitle.trim()) {
      createPostMutation.mutate({
        title: newPostTitle,
        body: 'This is a test post created from the Integration Demo',
        userId: 1,
      })
      setNewPostTitle('')
      addNotification({
        title: 'Post Creation',
        message: 'Post creation initiated',
        type: 'info',
      })
    }
  }

  const handleLogin = () => {
    login({
      id: '1',
      name: 'Demo User',
      email: 'demo@example.com',
      avatar: 'https://github.com/shadcn.png',
    })
    addNotification({
      title: 'Login',
      message: 'Successfully logged in!',
      type: 'success',
    })
  }

  const handleLogout = () => {
    logout()
    addNotification({
      title: 'Logout',
      message: 'Successfully logged out!',
      type: 'info',
    })
  }

  const toggleTheme = () => {
    const newTheme = theme === 'light' ? 'dark' : 'light'
    setTheme(newTheme)
    addNotification({
      title: 'Theme Change',
      message: `Theme changed to ${newTheme}`,
      type: 'success',
    })
  }

  return (
    <div className={`min-h-screen transition-colors duration-300 ${
      theme === 'dark' ? 'bg-gray-900 text-white' : 'bg-gray-50 text-gray-900'
    }`}>
      {/* Background Animation */}
      <div className="fixed inset-0 z-0">
        <BackgroundGradientAnimation>
          <div className="absolute z-50 inset-0 flex items-center justify-center text-white font-bold px-4 pointer-events-none text-3xl text-center md:text-4xl lg:text-7xl">
            <p className="bg-clip-text text-transparent drop-shadow-2xl bg-gradient-to-b from-white/80 to-white/20">
              Integration Demo
            </p>
          </div>
        </BackgroundGradientAnimation>
      </div>

      {/* Content */}
      <div className="relative z-10 container mx-auto px-4 py-8">
        <div className="mb-8">
          <TextGenerateEffect 
            words="Tennis Court Monitoring - Integration Demo"
            className="text-2xl md:text-4xl font-bold text-center mb-4"
          />
          <p className="text-center text-lg opacity-80">
            Demonstrating Zustand + React Query + ShadCN UI + Aceternity UI
          </p>
        </div>

        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {/* Zustand State Management Demo */}
          <Card className={`${theme === 'dark' ? 'bg-gray-800 border-gray-700' : 'bg-white'}`}>
            <CardHeader>
              <CardTitle>Zustand State Management</CardTitle>
              <CardDescription>Global state with persistence</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Theme Toggle */}
              <div className="flex items-center justify-between">
                <Label htmlFor="theme-toggle">Dark Mode</Label>
                <Switch
                  id="theme-toggle"
                  checked={theme === 'dark'}
                  onCheckedChange={toggleTheme}
                />
              </div>

              {/* Authentication */}
              <div className="space-y-2">
                <Label>Authentication Status</Label>
                <div className="flex items-center justify-between">
                  <span className={`text-sm ${isAuthenticated ? 'text-green-600' : 'text-red-600'}`}>
                    {isAuthenticated ? '✅ Logged In' : '❌ Logged Out'}
                  </span>
                  <Button
                    size="sm"
                    onClick={isAuthenticated ? handleLogout : handleLogin}
                    variant={isAuthenticated ? 'destructive' : 'default'}
                  >
                    {isAuthenticated ? 'Logout' : 'Login'}
                  </Button>
                </div>
              </div>

              {/* User Profile */}
              {isAuthenticated && userProfile && (
                <div className="p-3 bg-gray-100 dark:bg-gray-700 rounded">
                  <div className="flex items-center space-x-3">
                    <img
                      src={userProfile.avatar}
                      alt={userProfile.name}
                      className="w-8 h-8 rounded-full"
                    />
                    <div>
                      <p className="font-medium text-sm">{userProfile.name}</p>
                      <p className="text-xs opacity-70">{userProfile.email}</p>
                    </div>
                  </div>
                </div>
              )}

              {/* Current State Display */}
              <div className="text-xs space-y-1 p-3 bg-gray-100 dark:bg-gray-700 rounded">
                <div>Theme: <span className="font-mono">{theme}</span></div>
                <div>Auth: <span className="font-mono">{isAuthenticated.toString()}</span></div>
                <div>Notifications: <span className="font-mono">{notifications.length}</span></div>
              </div>
            </CardContent>
          </Card>

          {/* React Query Demo */}
          <Card className={`${theme === 'dark' ? 'bg-gray-800 border-gray-700' : 'bg-white'}`}>
            <CardHeader>
              <CardTitle>React Query (Server State)</CardTitle>
              <CardDescription>Data fetching and caching</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Fetch Controls */}
              <div className="flex space-x-2">
                <Button
                  size="sm"
                  onClick={() => refetch()}
                  disabled={isLoading}
                  className="flex-1"
                >
                  {isLoading ? 'Loading...' : 'Refetch Posts'}
                </Button>
              </div>

              {/* Create Post */}
              <div className="space-y-2">
                <Label>Create New Post</Label>
                <div className="flex space-x-2">
                  <Input
                    placeholder="Post title"
                    value={newPostTitle}
                    onChange={(e) => setNewPostTitle(e.target.value)}
                    className="flex-1"
                  />
                  <Button
                    size="sm"
                    onClick={handleCreatePost}
                    disabled={createPostMutation.isPending || !newPostTitle.trim()}
                  >
                    {createPostMutation.isPending ? '...' : 'Create'}
                  </Button>
                </div>
              </div>

              {/* Status Display */}
              <div className="text-xs space-y-1 p-3 bg-gray-100 dark:bg-gray-700 rounded">
                <div>Status: <span className={`font-mono ${isLoading ? 'text-yellow-600' : 'text-green-600'}`}>
                  {isLoading ? 'Loading' : 'Ready'}
                </span></div>
                <div>Posts: <span className="font-mono">{posts?.length || 0}</span></div>
                <div>Mutation: <span className={`font-mono ${createPostMutation.isPending ? 'text-yellow-600' : 'text-gray-600'}`}>
                  {createPostMutation.isPending ? 'Pending' : 'Idle'}
                </span></div>
              </div>

              {/* Error Display */}
              {error && (
                <div className="p-2 bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200 rounded text-xs">
                  Error: {error.message}
                </div>
              )}

              {/* Success Display */}
              {createPostMutation.isSuccess && (
                <div className="p-2 bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200 rounded text-xs">
                  Post created successfully!
                </div>
              )}
            </CardContent>
          </Card>

          {/* Notifications Demo */}
          <Card className={`${theme === 'dark' ? 'bg-gray-800 border-gray-700' : 'bg-white'}`}>
            <CardHeader>
              <CardTitle>Notifications System</CardTitle>
              <CardDescription>Real-time state updates</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex space-x-2">
                <Button
                  size="sm"
                  onClick={() => addNotification({
                    title: 'Test Notification',
                    message: 'Test notification',
                    type: 'info',
                  })}
                  className="flex-1"
                >
                  Add Notification
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={clearNotifications}
                  disabled={notifications.length === 0}
                >
                  Clear All
                </Button>
              </div>

              <div className="space-y-2 max-h-48 overflow-y-auto">
                {notifications.length === 0 ? (
                  <p className="text-sm text-gray-500 text-center py-4">No notifications</p>
                ) : (
                  notifications.slice(-5).reverse().map((notification) => (
                    <div
                      key={notification.id}
                      className={`p-2 rounded text-xs ${
                        notification.type === 'success' ? 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200' :
                        notification.type === 'error' ? 'bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200' :
                        'bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200'
                      }`}
                    >
                      <div className="font-medium">{notification.message}</div>
                      <div className="opacity-70">
                        {notification.timestamp.toLocaleTimeString()}
                      </div>
                    </div>
                  ))
                )}
              </div>
            </CardContent>
          </Card>

          {/* Posts Display */}
          <Card className={`md:col-span-2 lg:col-span-3 ${theme === 'dark' ? 'bg-gray-800 border-gray-700' : 'bg-white'}`}>
            <CardHeader>
              <CardTitle>Fetched Posts (React Query + ShadCN UI)</CardTitle>
              <CardDescription>Data from JSONPlaceholder API</CardDescription>
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <div className="flex items-center justify-center py-8">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
                  <span className="ml-2">Loading posts...</span>
                </div>
              ) : (
                <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                  {posts?.map((post: Post) => (
                    <Card key={post.id} className={`${theme === 'dark' ? 'bg-gray-700 border-gray-600' : 'bg-gray-50'}`}>
                      <CardHeader className="pb-3">
                        <CardTitle className="text-sm">{post.title}</CardTitle>
                      </CardHeader>
                      <CardContent className="pt-0">
                        <p className="text-xs opacity-70 line-clamp-3">
                          {post.body}
                        </p>
                        <div className="mt-2 text-xs opacity-50">
                          Post #{post.id} • User {post.userId}
                        </div>
                      </CardContent>
                    </Card>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Integration Status Summary */}
        <Card className={`mt-6 ${theme === 'dark' ? 'bg-gray-800 border-gray-700' : 'bg-white'}`}>
          <CardHeader>
            <CardTitle>Integration Status</CardTitle>
            <CardDescription>All systems working together</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid md:grid-cols-4 gap-4 text-sm">
              <div className="text-center">
                <div className="text-2xl mb-2">🎨</div>
                <div className="font-medium">ShadCN UI</div>
                <div className="text-green-600">✅ Active</div>
              </div>
              <div className="text-center">
                <div className="text-2xl mb-2">✨</div>
                <div className="font-medium">Aceternity UI</div>
                <div className="text-green-600">✅ Active</div>
              </div>
              <div className="text-center">
                <div className="text-2xl mb-2">🐻</div>
                <div className="font-medium">Zustand</div>
                <div className="text-green-600">✅ Active</div>
              </div>
              <div className="text-center">
                <div className="text-2xl mb-2">🔄</div>
                <div className="font-medium">React Query</div>
                <div className="text-green-600">✅ Active</div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
} 