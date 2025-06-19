/**
 * Production-ready logging utility for frontend
 * Provides structured logging with different levels and environment-based behavior
 */

export enum LogLevel {
  DEBUG = 0,
  INFO = 1,
  WARN = 2,
  ERROR = 3,
}

interface LogEntry {
  timestamp: string;
  level: string;
  service: string;
  message: string;
  context?: Record<string, any>;
  error?: {
    message: string;
    stack?: string;
    name?: string;
  };
}

class Logger {
  private serviceName: string;
  private minLevel: LogLevel;
  private isDevelopment: boolean;

  constructor(serviceName: string = 'tennis-frontend') {
    this.serviceName = serviceName;
    this.isDevelopment = import.meta.env.MODE === 'development';
    
    // Set minimum log level based on environment
    const envLevel = import.meta.env.VITE_LOG_LEVEL?.toLowerCase();
    switch (envLevel) {
      case 'debug':
        this.minLevel = LogLevel.DEBUG;
        break;
      case 'info':
        this.minLevel = LogLevel.INFO;
        break;
      case 'warn':
        this.minLevel = LogLevel.WARN;
        break;
      case 'error':
        this.minLevel = LogLevel.ERROR;
        break;
      default:
        this.minLevel = this.isDevelopment ? LogLevel.DEBUG : LogLevel.INFO;
    }
  }

  private shouldLog(level: LogLevel): boolean {
    return level >= this.minLevel;
  }

  private formatMessage(level: LogLevel, message: string, context?: Record<string, any>, error?: Error): LogEntry {
    const entry: LogEntry = {
      timestamp: new Date().toISOString(),
      level: LogLevel[level],
      service: this.serviceName,
      message,
    };

    if (context && Object.keys(context).length > 0) {
      entry.context = context;
    }

    if (error) {
      entry.error = {
        message: error.message,
        name: error.name,
        stack: error.stack,
      };
    }

    return entry;
  }

  private log(level: LogLevel, message: string, context?: Record<string, any>, error?: Error): void {
    if (!this.shouldLog(level)) {
      return;
    }

    const entry = this.formatMessage(level, message, context, error);

    if (this.isDevelopment) {
      // Human-readable format for development
      const levelStr = `[${entry.level}]`;
      const timestamp = new Date(entry.timestamp).toLocaleTimeString();
      const contextStr = entry.context ? ` ${JSON.stringify(entry.context)}` : '';
      const errorStr = entry.error ? ` Error: ${entry.error.message}` : '';
      
      const fullMessage = `${timestamp} ${levelStr} ${entry.service}: ${entry.message}${contextStr}${errorStr}`;
      
      switch (level) {
        case LogLevel.ERROR:
          console.error(fullMessage, error);
          break;
        case LogLevel.WARN:
          console.warn(fullMessage);
          break;
        case LogLevel.INFO:
          console.info(fullMessage);
          break;
        case LogLevel.DEBUG:
          console.debug(fullMessage);
          break;
      }
    } else {
      // Structured JSON format for production (could be sent to logging service)
      console.log(JSON.stringify(entry));
      
      // In production, you might want to send logs to a service like LogRocket, Sentry, etc.
      this.sendToLoggingService(entry);
    }
  }

  private sendToLoggingService(entry: LogEntry): void {
    // In production, send logs to external service
    // Example: LogRocket, Sentry, DataDog, etc.
    
    // For now, we'll just store in sessionStorage for debugging
    try {
      const logs = JSON.parse(sessionStorage.getItem('tennis_logs') || '[]');
      logs.push(entry);
      
      // Keep only last 100 logs to prevent memory issues
      if (logs.length > 100) {
        logs.splice(0, logs.length - 100);
      }
      
      sessionStorage.setItem('tennis_logs', JSON.stringify(logs));
    } catch (e) {
      // Silently fail if sessionStorage is not available
    }
  }

  debug(message: string, context?: Record<string, any>): void {
    this.log(LogLevel.DEBUG, message, context);
  }

  info(message: string, context?: Record<string, any>): void {
    this.log(LogLevel.INFO, message, context);
  }

  warn(message: string, context?: Record<string, any>): void {
    this.log(LogLevel.WARN, message, context);
  }

  error(message: string, context?: Record<string, any>, error?: Error): void {
    this.log(LogLevel.ERROR, message, context, error);
  }

  // Convenience methods for common scenarios
  apiRequest(method: string, url: string, context?: Record<string, any>): void {
    this.debug(`API Request: ${method} ${url}`, { 
      method, 
      url, 
      ...context,
      event: 'api_request' 
    });
  }

  apiResponse(method: string, url: string, status: number, context?: Record<string, any>): void {
    const level = status >= 400 ? LogLevel.ERROR : LogLevel.DEBUG;
    this.log(level, `API Response: ${method} ${url} - ${status}`, {
      method,
      url,
      status,
      ...context,
      event: 'api_response'
    });
  }

  userAction(action: string, context?: Record<string, any>): void {
    this.info(`User Action: ${action}`, {
      action,
      ...context,
      event: 'user_action'
    });
  }

  navigationEvent(from: string, to: string): void {
    this.debug(`Navigation: ${from} → ${to}`, {
      from,
      to,
      event: 'navigation'
    });
  }

  authEvent(event: string, success: boolean, context?: Record<string, any>): void {
    const level = success ? LogLevel.INFO : LogLevel.WARN;
    this.log(level, `Auth Event: ${event}`, {
      event: 'auth',
      success,
      ...context
    });
  }

  // Performance logging
  performanceEvent(name: string, duration: number, context?: Record<string, any>): void {
    this.info(`Performance: ${name}`, {
      name,
      duration_ms: duration,
      ...context,
      event: 'performance'
    });
  }
}

// Create and export singleton instance
export const logger = new Logger();

// Export convenience functions
export const log = {
  debug: (message: string, context?: Record<string, any>) => logger.debug(message, context),
  info: (message: string, context?: Record<string, any>) => logger.info(message, context),
  warn: (message: string, context?: Record<string, any>) => logger.warn(message, context),
  error: (message: string, context?: Record<string, any>, error?: Error) => logger.error(message, context, error),
  
  // Convenience methods
  apiRequest: (method: string, url: string, context?: Record<string, any>) => logger.apiRequest(method, url, context),
  apiResponse: (method: string, url: string, status: number, context?: Record<string, any>) => logger.apiResponse(method, url, status, context),
  userAction: (action: string, context?: Record<string, any>) => logger.userAction(action, context),
  navigation: (from: string, to: string) => logger.navigationEvent(from, to),
  auth: (event: string, success: boolean, context?: Record<string, any>) => logger.authEvent(event, success, context),
  performance: (name: string, duration: number, context?: Record<string, any>) => logger.performanceEvent(name, duration, context),
};

export default logger; 