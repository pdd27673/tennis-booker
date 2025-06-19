/**
 * Configuration module for React frontend
 * Handles environment-specific configuration using Vite's import.meta.env
 */

export interface AppConfig {
  // Application Settings
  appName: string;
  appVersion: string;
  environment: string;
  
  // API Configuration
  apiUrl: string;
  apiTimeout: number;
  
  // Feature Flags
  features: {
    analyticsEnabled: boolean;
    notificationsEnabled: boolean;
    advancedSearchEnabled: boolean;
    darkModeEnabled: boolean;
    mockApiEnabled: boolean;
    debugMode: boolean;
  };
  
  // Logging
  logLevel: string;
  
  // External Services
  googleAnalyticsId?: string;
  sentryDsn?: string;
}

/**
 * Parse a string value to boolean
 * Handles various string representations of boolean values
 */
function parseBoolean(value: string | undefined, defaultValue: boolean = false): boolean {
  if (!value) return defaultValue;
  return value.toLowerCase() === 'true' || value === '1';
}

/**
 * Parse a string value to number
 * Returns default value if parsing fails
 */
function parseNumber(value: string | undefined, defaultValue: number): number {
  if (!value) return defaultValue;
  const parsed = parseInt(value, 10);
  return isNaN(parsed) ? defaultValue : parsed;
}

/**
 * Get configuration from environment variables
 * All Vite environment variables are prefixed with VITE_
 */
function getConfig(): AppConfig {
  const env = import.meta.env;
  
  return {
    // Application Settings
    appName: env.VITE_APP_NAME || 'Tennis Booker',
    appVersion: env.VITE_APP_VERSION || '0.1.0',
    environment: env.VITE_APP_ENVIRONMENT || env.MODE || 'development',
    
    // API Configuration
    apiUrl: env.VITE_API_URL || 'http://localhost:8080',
    apiTimeout: parseNumber(env.VITE_API_TIMEOUT, 30000),
    
    // Feature Flags
    features: {
      analyticsEnabled: parseBoolean(env.VITE_FEATURE_ANALYTICS_ENABLED, true),
      notificationsEnabled: parseBoolean(env.VITE_FEATURE_NOTIFICATIONS_ENABLED, true),
      advancedSearchEnabled: parseBoolean(env.VITE_FEATURE_ADVANCED_SEARCH_ENABLED, true),
      darkModeEnabled: parseBoolean(env.VITE_FEATURE_DARK_MODE_ENABLED, true),
      mockApiEnabled: parseBoolean(env.VITE_MOCK_API_ENABLED, false),
      debugMode: parseBoolean(env.VITE_DEBUG_MODE, env.MODE === 'development'),
    },
    
    // Logging
    logLevel: env.VITE_LOG_LEVEL || (env.MODE === 'development' ? 'debug' : 'info'),
    
    // External Services
    googleAnalyticsId: env.VITE_GOOGLE_ANALYTICS_ID,
    sentryDsn: env.VITE_SENTRY_DSN,
  };
}

/**
 * Validate configuration values
 * Throws an error if required configuration is missing or invalid
 */
function validateConfig(config: AppConfig): void {
  const errors: string[] = [];
  
  // Validate API URL
  try {
    new URL(config.apiUrl);
  } catch {
    errors.push(`Invalid API URL: ${config.apiUrl}`);
  }
  
  // Validate API timeout
  if (config.apiTimeout <= 0) {
    errors.push(`API timeout must be positive: ${config.apiTimeout}`);
  }
  
  // Validate log level
  const validLogLevels = ['debug', 'info', 'warn', 'error'];
  if (!validLogLevels.includes(config.logLevel.toLowerCase())) {
    errors.push(`Invalid log level: ${config.logLevel}. Must be one of: ${validLogLevels.join(', ')}`);
  }
  
  if (errors.length > 0) {
    throw new Error(`Configuration validation failed:\n${errors.join('\n')}`);
  }
}

// Create and validate configuration
const config = getConfig();
validateConfig(config);

// Export the configuration
export default config;

// Export individual configuration sections for convenience
export const { appName, appVersion, environment } = config;
export const { apiUrl, apiTimeout } = config;
export const { features } = config;
export const { logLevel } = config;

// Helper functions for common configuration checks
export const isProduction = () => environment === 'production';
export const isDevelopment = () => environment === 'development';
export const isFeatureEnabled = (feature: keyof typeof features) => features[feature];

// Configuration loaded successfully 