# 🔧 Configuration Management Guide

This document defines the unified configuration structure and naming conventions for the Tennis Booker application across all services (Go backend, Python scraper, React frontend).

## 🏗️ Configuration Architecture

### **Hybrid Approach: Environment Variables + Configuration Files**

We use a **hybrid approach** that combines:
- **Environment Variables**: For runtime configuration and environment-specific overrides
- **Configuration Files**: For structured defaults and complex configurations
- **Vault Integration**: For secure secret management (separate from this configuration system)

### **Configuration vs. Secrets Separation**

| **Configuration (This System)** | **Secrets (Vault System)** |
|----------------------------------|----------------------------|
| ✅ API endpoints and URLs        | 🔐 Database passwords      |
| ✅ Feature flags                 | 🔐 API keys and tokens     |
| ✅ Logging levels               | 🔐 Email credentials       |
| ✅ Timeouts and intervals       | 🔐 JWT secrets             |
| ✅ Port numbers                 | 🔐 Platform credentials    |
| ✅ Non-sensitive defaults       | 🔐 Encryption keys         |

## 📋 Environment Detection

### **Environment Variable: `APP_ENV`**

All services use the `APP_ENV` environment variable to determine the current environment:

```bash
# Development (default)
APP_ENV=development

# Production
APP_ENV=production

# Testing
APP_ENV=test

# Staging
APP_ENV=staging
```

**Fallback Behavior**: If `APP_ENV` is not set, defaults to `development`.

### **React Frontend: `NODE_ENV`**

React/Vite uses the standard `NODE_ENV` for build-time configuration:
- `NODE_ENV=development` - Development builds
- `NODE_ENV=production` - Production builds

## 🗂️ Directory Structure

```
tennis-booker/
├── config/                          # Configuration files
│   ├── default.json                 # Base configuration (all environments)
│   ├── development.json             # Development overrides
│   ├── production.json              # Production overrides
│   ├── test.json                    # Test environment overrides
│   └── staging.json                 # Staging overrides
├── apps/
│   ├── backend/
│   │   └── config/                  # Go service-specific configs
│   │       ├── default.json
│   │       ├── development.json
│   │       └── production.json
│   ├── scraper/
│   │   └── config/                  # Python service-specific configs
│   │       ├── default.json
│   │       ├── development.json
│   │       └── production.json
│   └── frontend/
│       ├── .env.development         # Vite development config
│       ├── .env.production          # Vite production config
│       └── .env.example             # Template for local setup
└── .env.example                     # Global environment template
```

## 🏷️ Naming Conventions

### **Environment Variables**

**Format**: `{SERVICE}_{CATEGORY}_{SETTING}`

**Case**: `UPPER_SNAKE_CASE`

**Examples**:
```bash
# Global settings (no service prefix)
APP_ENV=production
LOG_LEVEL=info
API_BASE_URL=https://api.tennisbooker.com

# Backend service
BACKEND_API_PORT=8080
BACKEND_DB_POOL_SIZE=10
BACKEND_JWT_EXPIRY=24h

# Scraper service  
SCRAPER_INTERVAL=300
SCRAPER_TIMEOUT=30
SCRAPER_MAX_RETRIES=3

# Notification service
NOTIFICATION_PORT=8081
NOTIFICATION_EMAIL_RATE_LIMIT=10
NOTIFICATION_BATCH_SIZE=50

# Feature flags
FEATURE_ADVANCED_FILTERING=true
FEATURE_SMS_NOTIFICATIONS=false
FEATURE_ANALYTICS=true
```

### **Configuration File Keys**

**Format**: `camelCase` for JSON files

**Structure**: Nested objects for organization

**Example** (`config/default.json`):
```json
{
  "app": {
    "name": "Tennis Booker",
    "version": "1.0.0",
    "environment": "development"
  },
  "api": {
    "port": 8080,
    "timeout": "30s",
    "rateLimit": {
      "enabled": true,
      "requestsPerMinute": 100
    }
  },
  "scraper": {
    "interval": 300,
    "timeout": 30,
    "maxRetries": 3,
    "daysAhead": 8,
    "platforms": {
      "clubspark": {
        "enabled": true,
        "baseUrl": "https://clubspark.lta.org.uk"
      },
      "courtsides": {
        "enabled": true,
        "baseUrl": "https://courtsides.com"
      }
    }
  },
  "notification": {
    "port": 8081,
    "emailRateLimit": 10,
    "batchSize": 50,
    "retryAttempts": 3
  },
  "logging": {
    "level": "info",
    "format": "json",
    "enableConsole": true,
    "enableFile": false
  },
  "features": {
    "advancedFiltering": true,
    "smsNotifications": false,
    "analytics": true,
    "realTimeUpdates": false
  }
}
```

## 🔄 Configuration Loading Priority

**Priority Order** (highest to lowest):

1. **Environment Variables** (highest priority)
2. **Environment-specific config file** (e.g., `production.json`)
3. **Default config file** (`default.json`)
4. **Application defaults** (hardcoded fallbacks)

**Example**:
```bash
# If these exist:
# config/default.json: { "api": { "port": 8080 } }
# config/production.json: { "api": { "port": 9000 } }
# Environment: API_PORT=3000

# Final result: port = 3000 (env var wins)
```

## 🛠️ Implementation by Service Type

### **Go Services (Backend/Notification)**

**Library**: `github.com/spf13/viper`

**Configuration Loading**:
```go
// config/config.go
package config

import (
    "fmt"
    "os"
    "strings"
    
    "github.com/spf13/viper"
)

type Config struct {
    App struct {
        Name        string `mapstructure:"name"`
        Version     string `mapstructure:"version"`
        Environment string `mapstructure:"environment"`
    } `mapstructure:"app"`
    
    API struct {
        Port    int    `mapstructure:"port"`
        Timeout string `mapstructure:"timeout"`
    } `mapstructure:"api"`
    
    Logging struct {
        Level  string `mapstructure:"level"`
        Format string `mapstructure:"format"`
    } `mapstructure:"logging"`
    
    Features map[string]bool `mapstructure:"features"`
}

func Load() (*Config, error) {
    viper.SetConfigName("default")
    viper.SetConfigType("json")
    viper.AddConfigPath("./config")
    viper.AddConfigPath("../config")
    viper.AddConfigPath("../../config")
    
    // Read default config
    if err := viper.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("failed to read default config: %w", err)
    }
    
    // Get environment
    env := os.Getenv("APP_ENV")
    if env == "" {
        env = "development"
    }
    
    // Merge environment-specific config
    viper.SetConfigName(env)
    if err := viper.MergeInConfig(); err != nil {
        // Environment-specific config is optional
        fmt.Printf("No %s config found, using defaults\n", env)
    }
    
    // Enable environment variable overrides
    viper.AutomaticEnv()
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    viper.SetEnvPrefix("") // No prefix for global vars
    
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    
    return &config, nil
}
```

### **Python Services (Scraper)**

**Library**: `python-dotenv` + `json` + `os.environ`

**Configuration Loading**:
```python
# config/config.py
import json
import os
from pathlib import Path
from typing import Dict, Any, Optional
from dotenv import load_dotenv

class Config:
    def __init__(self):
        # Load .env file for local development
        load_dotenv()
        
        self.environment = os.getenv('APP_ENV', 'development')
        self._config = self._load_config()
    
    def _load_config(self) -> Dict[str, Any]:
        """Load configuration from files and environment variables"""
        config_dir = Path(__file__).parent.parent / 'config'
        
        # Load default config
        default_path = config_dir / 'default.json'
        config = self._load_json_file(default_path)
        
        # Merge environment-specific config
        env_path = config_dir / f'{self.environment}.json'
        env_config = self._load_json_file(env_path)
        if env_config:
            config = self._deep_merge(config, env_config)
        
        # Apply environment variable overrides
        config = self._apply_env_overrides(config)
        
        return config
    
    def _load_json_file(self, path: Path) -> Optional[Dict[str, Any]]:
        """Load JSON configuration file"""
        try:
            with open(path, 'r') as f:
                return json.load(f)
        except FileNotFoundError:
            return None
        except json.JSONDecodeError as e:
            raise ValueError(f"Invalid JSON in {path}: {e}")
    
    def _deep_merge(self, base: Dict, override: Dict) -> Dict:
        """Deep merge two dictionaries"""
        result = base.copy()
        for key, value in override.items():
            if key in result and isinstance(result[key], dict) and isinstance(value, dict):
                result[key] = self._deep_merge(result[key], value)
            else:
                result[key] = value
        return result
    
    def _apply_env_overrides(self, config: Dict[str, Any]) -> Dict[str, Any]:
        """Apply environment variable overrides"""
        # Map common environment variables to config paths
        env_mappings = {
            'SCRAPER_INTERVAL': ['scraper', 'interval'],
            'SCRAPER_TIMEOUT': ['scraper', 'timeout'],
            'SCRAPER_MAX_RETRIES': ['scraper', 'maxRetries'],
            'LOG_LEVEL': ['logging', 'level'],
            'FEATURE_ADVANCED_FILTERING': ['features', 'advancedFiltering'],
        }
        
        for env_var, config_path in env_mappings.items():
            value = os.getenv(env_var)
            if value is not None:
                # Convert string values to appropriate types
                if value.lower() in ('true', 'false'):
                    value = value.lower() == 'true'
                elif value.isdigit():
                    value = int(value)
                
                # Set nested value
                current = config
                for key in config_path[:-1]:
                    current = current.setdefault(key, {})
                current[config_path[-1]] = value
        
        return config
    
    def get(self, key: str, default: Any = None) -> Any:
        """Get configuration value using dot notation"""
        keys = key.split('.')
        value = self._config
        
        for k in keys:
            if isinstance(value, dict) and k in value:
                value = value[k]
            else:
                return default
        
        return value

# Global config instance
config = Config()
```

### **React Frontend (Vite)**

**Environment Files**:

`.env.development`:
```bash
# Development environment
VITE_API_BASE_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080/ws
VITE_FEATURE_ADVANCED_FILTERING=true
VITE_FEATURE_REAL_TIME_UPDATES=true
VITE_LOG_LEVEL=debug
```

`.env.production`:
```bash
# Production environment
VITE_API_BASE_URL=https://api.tennisbooker.com
VITE_WS_URL=wss://api.tennisbooker.com/ws
VITE_FEATURE_ADVANCED_FILTERING=true
VITE_FEATURE_REAL_TIME_UPDATES=false
VITE_LOG_LEVEL=error
```

**Usage in React**:
```typescript
// config/config.ts
interface AppConfig {
  apiBaseUrl: string;
  wsUrl: string;
  features: {
    advancedFiltering: boolean;
    realTimeUpdates: boolean;
  };
  logging: {
    level: string;
  };
}

export const config: AppConfig = {
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
  wsUrl: import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws',
  features: {
    advancedFiltering: import.meta.env.VITE_FEATURE_ADVANCED_FILTERING === 'true',
    realTimeUpdates: import.meta.env.VITE_FEATURE_REAL_TIME_UPDATES === 'true',
  },
  logging: {
    level: import.meta.env.VITE_LOG_LEVEL || 'info',
  },
};
```

## 📝 Common Configuration Categories

### **Application Settings**
- `app.name` - Application name
- `app.version` - Application version
- `app.environment` - Current environment

### **API Configuration**
- `api.port` - Server port
- `api.timeout` - Request timeout
- `api.baseUrl` - API base URL
- `api.rateLimit.*` - Rate limiting settings

### **Database Configuration**
- `database.poolSize` - Connection pool size
- `database.timeout` - Query timeout
- `database.retryAttempts` - Retry attempts

### **Scraper Configuration**
- `scraper.interval` - Scraping interval (seconds)
- `scraper.timeout` - Request timeout (seconds)
- `scraper.maxRetries` - Maximum retry attempts
- `scraper.daysAhead` - Days to scrape ahead
- `scraper.platforms.*` - Platform-specific settings

### **Notification Configuration**
- `notification.port` - Service port
- `notification.emailRateLimit` - Email rate limit
- `notification.batchSize` - Batch processing size
- `notification.retryAttempts` - Retry attempts

### **Logging Configuration**
- `logging.level` - Log level (debug, info, warn, error)
- `logging.format` - Log format (json, text)
- `logging.enableConsole` - Console logging
- `logging.enableFile` - File logging

### **Feature Flags**
- `features.advancedFiltering` - Enable advanced filtering
- `features.smsNotifications` - Enable SMS notifications
- `features.analytics` - Enable analytics
- `features.realTimeUpdates` - Enable real-time updates

## 🧪 Testing Configuration

### **Test Environment Setup**

Create `config/test.json`:
```json
{
  "api": {
    "port": 0
  },
  "logging": {
    "level": "error",
    "enableConsole": false
  },
  "scraper": {
    "interval": 1,
    "timeout": 5
  },
  "features": {
    "advancedFiltering": true,
    "smsNotifications": false,
    "analytics": false
  }
}
```

### **Environment Variable Testing**

```bash
# Test with environment overrides
APP_ENV=test \
LOG_LEVEL=debug \
SCRAPER_INTERVAL=10 \
npm test
```

## 🔍 Validation and Best Practices

### **Configuration Validation**

1. **Required Fields**: Validate that required configuration fields are present
2. **Type Checking**: Ensure values are of the correct type
3. **Range Validation**: Check that numeric values are within acceptable ranges
4. **URL Validation**: Validate URL formats for endpoints

### **Best Practices**

1. **Default Values**: Always provide sensible defaults
2. **Documentation**: Document all configuration options
3. **Environment Parity**: Keep development and production configs similar
4. **Secrets Separation**: Never put secrets in configuration files
5. **Validation**: Validate configuration on application startup
6. **Immutability**: Treat configuration as immutable after loading

## 🚀 Migration Guide

### **From Current State**

1. **Create Configuration Files**: Add `config/` directory with JSON files
2. **Add Viper Dependency**: `go get github.com/spf13/viper`
3. **Update Go Services**: Implement configuration loading
4. **Update Python Services**: Implement configuration loading
5. **Create React Environment Files**: Add `.env.*` files
6. **Update Documentation**: Document new configuration system

### **Backward Compatibility**

- Environment variables continue to work as overrides
- Existing `.env` files remain functional
- Gradual migration path for each service

---

This configuration system provides a robust, scalable foundation for managing application settings across all services while maintaining clear separation from secret management. 