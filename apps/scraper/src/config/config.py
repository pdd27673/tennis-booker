"""
Configuration management for the Tennis Booker scraper service.

This module loads configuration from JSON files and environment variables,
following the unified configuration structure defined in the project.
"""

import json
import os
from pathlib import Path
from typing import Dict, Any, Optional, Union
from dotenv import load_dotenv


class Config:
    """Configuration manager for the scraper service"""
    
    def __init__(self):
        # Load .env file for local development
        load_dotenv()
        
        self.environment = os.getenv('APP_ENV', 'development')
        self._config = self._load_config()
    
    def _load_config(self) -> Dict[str, Any]:
        """Load configuration from files and environment variables"""
        # Find config directory (search multiple locations)
        config_dir = self._find_config_directory()
        
        if not config_dir:
            raise FileNotFoundError("Could not find config directory")
        
        # Load default config
        default_path = config_dir / 'default.json'
        config = self._load_json_file(default_path)
        
        if not config:
            raise FileNotFoundError(f"Could not load default config from {default_path}")
        
        # Merge environment-specific config
        env_path = config_dir / f'{self.environment}.json'
        env_config = self._load_json_file(env_path)
        if env_config:
            config = self._deep_merge(config, env_config)
        
        # Apply environment variable overrides
        config = self._apply_env_overrides(config)
        
        return config
    
    def _find_config_directory(self) -> Optional[Path]:
        """Find the config directory by searching multiple locations"""
        current_dir = Path(__file__).parent
        
        # Search paths relative to this file
        search_paths = [
            current_dir / 'config',  # Same directory
            current_dir.parent / 'config',  # Parent directory (src/config)
            current_dir.parent.parent / 'config',  # apps/scraper/config
            current_dir.parent.parent.parent / 'config',  # apps/config
            current_dir.parent.parent.parent.parent / 'config',  # project root config
        ]
        
        for path in search_paths:
            if path.exists() and (path / 'default.json').exists():
                return path
        
        return None
    
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
            # App settings
            'APP_ENV': ['app', 'environment'],
            
            # Scraper settings
            'SCRAPER_INTERVAL': ['scraper', 'interval'],
            'SCRAPER_TIMEOUT': ['scraper', 'timeout'],
            'SCRAPER_MAX_RETRIES': ['scraper', 'maxRetries'],
            'SCRAPER_DAYS_AHEAD': ['scraper', 'daysAhead'],
            
            # Logging settings
            'LOG_LEVEL': ['logging', 'level'],
            'LOG_FORMAT': ['logging', 'format'],
            'LOG_ENABLE_CONSOLE': ['logging', 'enableConsole'],
            'LOG_ENABLE_FILE': ['logging', 'enableFile'],
            
            # Feature flags
            'FEATURE_ADVANCED_FILTERING': ['features', 'advancedFiltering'],
            'FEATURE_SMS_NOTIFICATIONS': ['features', 'smsNotifications'],
            'FEATURE_ANALYTICS': ['features', 'analytics'],
            'FEATURE_REAL_TIME_UPDATES': ['features', 'realTimeUpdates'],
            
            # Platform settings
            'CLUBSPARK_ENABLED': ['scraper', 'platforms', 'clubspark', 'enabled'],
            'CLUBSPARK_BASE_URL': ['scraper', 'platforms', 'clubspark', 'baseUrl'],
            'COURTSIDES_ENABLED': ['scraper', 'platforms', 'courtsides', 'enabled'],
            'COURTSIDES_BASE_URL': ['scraper', 'platforms', 'courtsides', 'baseUrl'],
        }
        
        for env_var, config_path in env_mappings.items():
            value = os.getenv(env_var)
            if value is not None:
                # Convert string values to appropriate types
                converted_value = self._convert_env_value(value)
                
                # Set nested value
                current = config
                for key in config_path[:-1]:
                    current = current.setdefault(key, {})
                current[config_path[-1]] = converted_value
        
        return config
    
    def _convert_env_value(self, value: str) -> Union[str, int, float, bool]:
        """Convert environment variable string to appropriate type"""
        # Boolean conversion
        if value.lower() in ('true', 'false'):
            return value.lower() == 'true'
        
        # Integer conversion
        if value.isdigit():
            return int(value)
        
        # Float conversion
        try:
            if '.' in value:
                return float(value)
        except ValueError:
            pass
        
        # Return as string
        return value
    
    def _parse_duration(self, duration_str: str) -> int:
        """Parse duration string to seconds (e.g., '5m' -> 300, '30s' -> 30)"""
        if not isinstance(duration_str, str):
            return int(duration_str)
        
        duration_str = duration_str.strip().lower()
        
        # If it's just a number, return as is
        if duration_str.isdigit():
            return int(duration_str)
        
        # Parse duration with units
        import re
        match = re.match(r'^(\d+(?:\.\d+)?)\s*([smhd]?)$', duration_str)
        if not match:
            # If we can't parse it, try to convert to int
            try:
                return int(float(duration_str))
            except ValueError:
                raise ValueError(f"Invalid duration format: {duration_str}")
        
        value, unit = match.groups()
        value = float(value)
        
        # Convert to seconds
        if unit == 's' or unit == '':
            return int(value)
        elif unit == 'm':
            return int(value * 60)
        elif unit == 'h':
            return int(value * 3600)
        elif unit == 'd':
            return int(value * 86400)
        else:
            raise ValueError(f"Unknown duration unit: {unit}")
    
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
    
    def get_app_name(self) -> str:
        """Get application name"""
        return self.get('app.name', 'Tennis Booker Scraper')
    
    def get_app_version(self) -> str:
        """Get application version"""
        return self.get('app.version', '1.0.0')
    
    def get_environment(self) -> str:
        """Get current environment"""
        return self.get('app.environment', 'development')
    
    def get_scraper_interval(self) -> int:
        """Get scraper interval in seconds"""
        value = self.get('scraper.interval', 300)
        if isinstance(value, str):
            # Handle duration strings like "5m", "300s"
            return self._parse_duration(value)
        return int(value)
    
    def get_scraper_timeout(self) -> int:
        """Get scraper timeout in seconds"""
        value = self.get('scraper.timeout', 30)
        if isinstance(value, str):
            # Handle duration strings like "30s", "1m"
            return self._parse_duration(value)
        return int(value)
    
    def get_scraper_max_retries(self) -> int:
        """Get maximum retry attempts"""
        return self.get('scraper.maxRetries', 3)
    
    def get_scraper_days_ahead(self) -> int:
        """Get number of days ahead to scrape"""
        return self.get('scraper.daysAhead', 8)
    
    def get_log_level(self) -> str:
        """Get logging level"""
        return self.get('logging.level', 'info').upper()
    
    def get_log_format(self) -> str:
        """Get logging format"""
        return self.get('logging.format', 'json')
    
    def is_log_console_enabled(self) -> bool:
        """Check if console logging is enabled"""
        return self.get('logging.enableConsole', True)
    
    def is_log_file_enabled(self) -> bool:
        """Check if file logging is enabled"""
        return self.get('logging.enableFile', False)
    
    def is_feature_enabled(self, feature: str) -> bool:
        """Check if a feature flag is enabled"""
        # Convert camelCase to lowercase for consistency with Viper behavior
        feature_key = feature.lower()
        enabled = self.get(f'features.{feature_key}', False)
        return bool(enabled)
    
    def get_platform_config(self, platform: str) -> Dict[str, Any]:
        """Get platform-specific configuration"""
        return self.get(f'scraper.platforms.{platform}', {})
    
    def is_platform_enabled(self, platform: str) -> bool:
        """Check if a platform is enabled"""
        platform_config = self.get_platform_config(platform)
        return platform_config.get('enabled', False)
    
    def get_platform_base_url(self, platform: str) -> Optional[str]:
        """Get platform base URL"""
        platform_config = self.get_platform_config(platform)
        return platform_config.get('baseUrl')
    
    def is_development(self) -> bool:
        """Check if running in development environment"""
        return self.get_environment() == 'development'
    
    def is_production(self) -> bool:
        """Check if running in production environment"""
        return self.get_environment() == 'production'
    
    def is_test(self) -> bool:
        """Check if running in test environment"""
        return self.get_environment() == 'test'
    
    def validate(self) -> None:
        """Validate the configuration"""
        errors = []
        
        # Validate required fields
        if not self.get_app_name():
            errors.append("app.name is required")
        
        # Validate positive integers (with proper type handling)
        try:
            if self.get_scraper_interval() <= 0:
                errors.append("scraper.interval must be positive")
        except (ValueError, TypeError) as e:
            errors.append(f"scraper.interval is invalid: {e}")
        
        try:
            if self.get_scraper_timeout() <= 0:
                errors.append("scraper.timeout must be positive")
        except (ValueError, TypeError) as e:
            errors.append(f"scraper.timeout is invalid: {e}")
        
        try:
            if self.get_scraper_max_retries() < 0:
                errors.append("scraper.maxRetries must be non-negative")
        except (ValueError, TypeError) as e:
            errors.append(f"scraper.maxRetries is invalid: {e}")
        
        try:
            if self.get_scraper_days_ahead() <= 0:
                errors.append("scraper.daysAhead must be positive")
        except (ValueError, TypeError) as e:
            errors.append(f"scraper.daysAhead is invalid: {e}")
        
        # Validate log level
        valid_log_levels = ['DEBUG', 'INFO', 'WARNING', 'ERROR', 'CRITICAL']
        if self.get_log_level() not in valid_log_levels:
            errors.append(f"logging.level must be one of: {valid_log_levels}")
        
        if errors:
            raise ValueError(f"Configuration validation failed: {'; '.join(errors)}")


# Global configuration instance
_config_instance: Optional[Config] = None


def load_config() -> Config:
    """Load and return the global configuration instance"""
    global _config_instance
    
    if _config_instance is None:
        _config_instance = Config()
        _config_instance.validate()
    
    return _config_instance


def get_config() -> Config:
    """Get the global configuration instance (loads if not already loaded)"""
    return load_config()


# Convenience functions for common configuration access
def get_scraper_interval() -> int:
    """Get scraper interval in seconds"""
    return get_config().get_scraper_interval()


def get_scraper_timeout() -> int:
    """Get scraper timeout in seconds"""
    return get_config().get_scraper_timeout()


def get_log_level() -> str:
    """Get logging level"""
    return get_config().get_log_level()


def is_feature_enabled(feature: str) -> bool:
    """Check if a feature flag is enabled"""
    return get_config().is_feature_enabled(feature)


def is_platform_enabled(platform: str) -> bool:
    """Check if a platform is enabled"""
    return get_config().is_platform_enabled(platform) 