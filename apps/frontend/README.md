# Tennis Booker Frontend

React 18 + TypeScript + Vite + Tailwind CSS frontend for the Tennis Booker application.

## Quick Start

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

## Configuration

This frontend uses Vite's environment variable system. All variables must be prefixed with `VITE_` to be accessible in the browser.

### Environment Files

- `.env.development` - Development settings
- `.env.production` - Production settings  
- `.env.local` - Local overrides (gitignored)

### Key Configuration Variables

```bash
# Application
VITE_APP_NAME="Tennis Booker"
VITE_APP_VERSION="0.1.0"
VITE_APP_ENVIRONMENT="development"

# API
VITE_API_URL="http://localhost:8080"

# Feature Flags
VITE_FEATURE_ANALYTICS_ENABLED=true
VITE_FEATURE_NOTIFICATIONS_ENABLED=true
VITE_FEATURE_ADVANCED_SEARCH_ENABLED=true
VITE_DEBUG_MODE=true
```

### Usage in Code

```typescript
import { appName, apiUrl, isFeatureEnabled } from './config/config';

console.log(`Welcome to ${appName}`);
if (isFeatureEnabled('analyticsEnabled')) {
  // Initialize analytics
}
```

## Tech Stack

- **React 18** - UI framework
- **TypeScript** - Type safety
- **Vite** - Build tool and dev server
- **Tailwind CSS** - Utility-first CSS framework
- **ESLint** - Code linting

## Project Structure

```
src/
├── config/          # Configuration management
├── styles/          # Global styles and Tailwind setup
├── App.tsx          # Main application component
└── main.tsx         # Application entry point
```

The configuration system includes validation and will display debug information in development mode when `VITE_DEBUG_MODE=true`.
