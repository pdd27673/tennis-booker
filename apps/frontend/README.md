# Tennis Booker Frontend

React 18 + TypeScript + Vite frontend for the Tennis Booker application.

## Configuration System

This frontend uses Vite's built-in environment variable support for configuration management. All configuration variables must be prefixed with `VITE_` to be accessible in the browser.

### Environment Files

Create environment-specific files in the project root:

- `.env.development` - Development environment settings
- `.env.production` - Production environment settings  
- `.env.local` - Local overrides (gitignored)

### Available Configuration Variables

#### Application Settings
- `VITE_APP_NAME` - Application name (default: "Tennis Booker")
- `VITE_APP_VERSION` - Application version (default: "0.1.0")
- `VITE_APP_ENVIRONMENT` - Environment name (default: Vite's MODE)

#### API Configuration
- `VITE_API_URL` - Backend API URL (default: "http://localhost:8080")
- `VITE_API_TIMEOUT` - API request timeout in milliseconds (default: 30000)

#### Feature Flags
- `VITE_FEATURE_ANALYTICS_ENABLED` - Enable analytics (default: true)
- `VITE_FEATURE_NOTIFICATIONS_ENABLED` - Enable notifications (default: true)
- `VITE_FEATURE_ADVANCED_SEARCH_ENABLED` - Enable advanced search (default: true)
- `VITE_FEATURE_DARK_MODE_ENABLED` - Enable dark mode (default: true)
- `VITE_MOCK_API_ENABLED` - Use mock API instead of real backend (default: false)
- `VITE_DEBUG_MODE` - Enable debug mode (default: true in development)

#### Logging
- `VITE_LOG_LEVEL` - Logging level: debug, info, warn, error (default: debug in dev, info in prod)

#### External Services
- `VITE_GOOGLE_ANALYTICS_ID` - Google Analytics tracking ID (optional)
- `VITE_SENTRY_DSN` - Sentry error tracking DSN (optional)

### Usage in Code

```typescript
import config, { 
  appName, 
  apiUrl, 
  features, 
  isFeatureEnabled 
} from './config/config';

// Use individual exports
console.log(`Welcome to ${appName}`);
console.log(`API URL: ${apiUrl}`);

// Check feature flags
if (isFeatureEnabled('analyticsEnabled')) {
  // Initialize analytics
}

// Use full config object
console.log('Full config:', config);
```

### Configuration Priority

1. Environment variables (highest priority)
2. Environment-specific files (`.env.development`, `.env.production`)
3. Default values in code (lowest priority)

### Validation

The configuration system includes validation for:
- API URL format
- Positive timeout values
- Valid log levels
- Required configuration presence

Invalid configuration will throw an error during application startup.

## Development

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

## Environment Setup

1. Copy `.env.example` to `.env.development`
2. Customize values for your local environment
3. For production, create `.env.production` with production values

The configuration will be displayed on the main page in development mode when debug mode is enabled.
