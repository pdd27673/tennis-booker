package ratelimit

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tennis-booker/internal/auth"
)

// RateLimitEvent represents a rate limiting event for logging
type RateLimitEvent struct {
	Timestamp    time.Time `json:"timestamp"`
	IP           string    `json:"ip"`
	UserID       string    `json:"user_id,omitempty"`
	Endpoint     string    `json:"endpoint"`
	Method       string    `json:"method"`
	LimitType    string    `json:"limit_type"`
	RequestsMade int64     `json:"requests_made"`
	Limit        int64     `json:"limit"`
	Window       string    `json:"window"`
	Blocked      bool      `json:"blocked"`
	UserAgent    string    `json:"user_agent,omitempty"`
}

// logRateLimitEvent logs rate limiting events for monitoring
func logRateLimitEvent(event RateLimitEvent) {
	if event.Blocked {
		log.Printf("[RATE_LIMIT_BLOCKED] IP=%s UserID=%s Endpoint=%s Method=%s Type=%s Requests=%d/%d Window=%s UserAgent=%s",
			event.IP, event.UserID, event.Endpoint, event.Method, event.LimitType,
			event.RequestsMade, event.Limit, event.Window, event.UserAgent)
	} else {
		// Only log successful requests in debug mode to avoid spam
		// In production, you might want to send this to a metrics system instead
		log.Printf("[RATE_LIMIT_DEBUG] IP=%s UserID=%s Endpoint=%s Requests=%d/%d Remaining=%d",
			event.IP, event.UserID, event.Endpoint, event.RequestsMade, event.Limit, event.Limit-event.RequestsMade)
	}
}

// IPRateLimitMiddleware creates HTTP middleware for IP-based rate limiting
func IPRateLimitMiddleware(limiter *Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract client IP address
			clientIP := extractClientIP(r, limiter.config.TrustedProxies)

			// Check rate limit for this IP
			result, err := limiter.CheckIPLimit(r.Context(), clientIP)
			if err != nil {
				// Log error but don't block request on rate limiter failure
				// In production, you might want to fail open or closed based on your security requirements
				log.Printf("[RATE_LIMIT_ERROR] IP=%s Endpoint=%s Error=%v", clientIP, r.URL.Path, err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Log rate limit event
			event := RateLimitEvent{
				Timestamp:    time.Now(),
				IP:           clientIP,
				Endpoint:     r.URL.Path,
				Method:       r.Method,
				LimitType:    "ip",
				RequestsMade: result.Limit - result.Remaining,
				Limit:        result.Limit,
				Window:       "1m",
				Blocked:      !result.Allowed,
				UserAgent:    r.Header.Get("User-Agent"),
			}
			logRateLimitEvent(event)

			// Add rate limit headers if configured
			if limiter.config.IncludeHeaders {
				addRateLimitHeaders(w, result)
			}

			// Check if rate limit exceeded
			if !result.Allowed {
				// Set Retry-After header
				w.Header().Set("Retry-After", strconv.Itoa(int(result.RetryAfter.Seconds())))

				// Return 429 Too Many Requests
				http.Error(w, fmt.Sprintf("Too many requests from IP %s. Try again in %d seconds.", clientIP, int(result.RetryAfter.Seconds())), http.StatusTooManyRequests)
				return
			}

			// Rate limit not exceeded, continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// AuthRateLimitMiddleware creates HTTP middleware for authentication endpoint rate limiting
func AuthRateLimitMiddleware(limiter *Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract client IP address for auth rate limiting
			clientIP := extractClientIP(r, limiter.config.TrustedProxies)

			// Check auth-specific rate limit for this IP
			result, err := limiter.CheckAuthLimit(r.Context(), clientIP)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Add rate limit headers if configured
			if limiter.config.IncludeHeaders {
				addRateLimitHeaders(w, result)
			}

			// Check if rate limit exceeded
			if !result.Allowed {
				w.Header().Set("Retry-After", strconv.Itoa(int(result.RetryAfter.Seconds())))
				http.Error(w, "Too many authentication requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// DataRateLimitMiddleware creates HTTP middleware for data endpoint rate limiting
func DataRateLimitMiddleware(limiter *Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract client IP address
			clientIP := extractClientIP(r, limiter.config.TrustedProxies)

			// Check data-specific rate limit for this IP
			result, err := limiter.CheckDataLimit(r.Context(), clientIP)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Add rate limit headers if configured
			if limiter.config.IncludeHeaders {
				addRateLimitHeaders(w, result)
			}

			// Check if rate limit exceeded
			if !result.Allowed {
				w.Header().Set("Retry-After", strconv.Itoa(int(result.RetryAfter.Seconds())))
				http.Error(w, "Too many data requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SensitiveRateLimitMiddleware creates HTTP middleware for sensitive endpoint rate limiting
func SensitiveRateLimitMiddleware(limiter *Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract client IP address
			clientIP := extractClientIP(r, limiter.config.TrustedProxies)

			// Check sensitive-specific rate limit for this IP
			result, err := limiter.CheckSensitiveLimit(r.Context(), clientIP)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Add rate limit headers if configured
			if limiter.config.IncludeHeaders {
				addRateLimitHeaders(w, result)
			}

			// Check if rate limit exceeded
			if !result.Allowed {
				w.Header().Set("Retry-After", strconv.Itoa(int(result.RetryAfter.Seconds())))
				http.Error(w, "Too many sensitive requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CustomRateLimitMiddleware creates HTTP middleware with custom rate limiting configuration
func CustomRateLimitMiddleware(limiter *Limiter, rateLimit RateLimit, keyPrefix string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract client IP address
			clientIP := extractClientIP(r, limiter.config.TrustedProxies)
			identifier := fmt.Sprintf("%s:%s", keyPrefix, clientIP)

			// Check custom rate limit
			result, err := limiter.CheckCustomLimit(r.Context(), identifier, rateLimit)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Add rate limit headers if configured
			if limiter.config.IncludeHeaders {
				addRateLimitHeaders(w, result)
			}

			// Check if rate limit exceeded
			if !result.Allowed {
				w.Header().Set("Retry-After", strconv.Itoa(int(result.RetryAfter.Seconds())))
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// UserRateLimitMiddleware creates HTTP middleware for user-based rate limiting
// This middleware should be applied AFTER JWT authentication middleware
func UserRateLimitMiddleware(limiter *Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to extract user ID from JWT context
			userID, err := auth.GetUserIDFromContext(r.Context())
			if err != nil {
				// No user context found - this could be an unauthenticated request
				// or middleware applied in wrong order. Fall back to IP-based limiting.
				clientIP := extractClientIP(r, limiter.config.TrustedProxies)
				result, err := limiter.CheckIPLimit(r.Context(), clientIP)
				if err != nil {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

				// Add rate limit headers if configured
				if limiter.config.IncludeHeaders {
					addRateLimitHeaders(w, result)
				}

				// Check if rate limit exceeded
				if !result.Allowed {
					w.Header().Set("Retry-After", strconv.Itoa(int(result.RetryAfter.Seconds())))
					http.Error(w, "Too many requests", http.StatusTooManyRequests)
					return
				}

				next.ServeHTTP(w, r)
				return
			}

			// Check user-specific rate limit
			result, err := limiter.CheckUserLimit(r.Context(), userID)
			if err != nil {
				log.Printf("[RATE_LIMIT_ERROR] UserID=%s Endpoint=%s Error=%v", userID, r.URL.Path, err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Log rate limit event
			clientIP := extractClientIP(r, limiter.config.TrustedProxies)
			event := RateLimitEvent{
				Timestamp:    time.Now(),
				IP:           clientIP,
				UserID:       userID,
				Endpoint:     r.URL.Path,
				Method:       r.Method,
				LimitType:    "user",
				RequestsMade: result.Limit - result.Remaining,
				Limit:        result.Limit,
				Window:       "1m",
				Blocked:      !result.Allowed,
				UserAgent:    r.Header.Get("User-Agent"),
			}
			logRateLimitEvent(event)

			// Add rate limit headers if configured
			if limiter.config.IncludeHeaders {
				addRateLimitHeaders(w, result)
			}

			// Check if rate limit exceeded
			if !result.Allowed {
				w.Header().Set("Retry-After", strconv.Itoa(int(result.RetryAfter.Seconds())))
				http.Error(w, "Too many requests for user", http.StatusTooManyRequests)
				return
			}

			// Rate limit not exceeded, continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// CombinedRateLimitMiddleware creates HTTP middleware that applies both IP and user-based rate limiting
// This provides defense in depth - both IP and user limits must be respected
func CombinedRateLimitMiddleware(limiter *Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Always check IP-based rate limit first
			clientIP := extractClientIP(r, limiter.config.TrustedProxies)
			ipResult, err := limiter.CheckIPLimit(r.Context(), clientIP)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Check if IP rate limit exceeded
			if !ipResult.Allowed {
				if limiter.config.IncludeHeaders {
					addRateLimitHeaders(w, ipResult)
				}
				w.Header().Set("Retry-After", strconv.Itoa(int(ipResult.RetryAfter.Seconds())))
				http.Error(w, "Too many requests from IP", http.StatusTooManyRequests)
				return
			}

			// Try to get user ID for additional user-based limiting
			userID, err := auth.GetUserIDFromContext(r.Context())
			if err == nil {
				// User is authenticated, also check user-based rate limit
				userResult, err := limiter.CheckUserLimit(r.Context(), userID)
				if err != nil {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

				// Check if user rate limit exceeded
				if !userResult.Allowed {
					if limiter.config.IncludeHeaders {
						addRateLimitHeaders(w, userResult)
					}
					w.Header().Set("Retry-After", strconv.Itoa(int(userResult.RetryAfter.Seconds())))
					http.Error(w, "Too many requests for user", http.StatusTooManyRequests)
					return
				}

				// Use the more restrictive limit for headers (lower remaining count)
				if limiter.config.IncludeHeaders {
					if userResult.Remaining < ipResult.Remaining {
						addRateLimitHeaders(w, userResult)
					} else {
						addRateLimitHeaders(w, ipResult)
					}
				}
			} else {
				// User not authenticated, just use IP rate limit headers
				if limiter.config.IncludeHeaders {
					addRateLimitHeaders(w, ipResult)
				}
			}

			// Both IP and user limits (if applicable) are within bounds
			next.ServeHTTP(w, r)
		})
	}
}

// UserAuthRateLimitMiddleware creates HTTP middleware for user-based authentication endpoint rate limiting
// This applies stricter limits for auth endpoints on a per-user basis
func UserAuthRateLimitMiddleware(limiter *Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// For auth endpoints, we might not have user context yet (e.g., login)
			// So we'll use a combination of IP and any available user identifier

			// Try to extract user ID from context (might be available for some auth operations)
			userID, err := auth.GetUserIDFromContext(r.Context())
			var identifier string

			if err != nil {
				// No user context, fall back to IP-based auth limiting
				clientIP := extractClientIP(r, limiter.config.TrustedProxies)
				identifier = clientIP
			} else {
				// Use user ID for more specific limiting
				identifier = userID
			}

			// Check auth-specific rate limit
			result, err := limiter.CheckAuthLimit(r.Context(), identifier)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Add rate limit headers if configured
			if limiter.config.IncludeHeaders {
				addRateLimitHeaders(w, result)
			}

			// Check if rate limit exceeded
			if !result.Allowed {
				w.Header().Set("Retry-After", strconv.Itoa(int(result.RetryAfter.Seconds())))
				http.Error(w, "Too many authentication requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractClientIP extracts the real client IP address from the request
// It checks various headers in order of preference and validates against trusted proxies
func extractClientIP(r *http.Request, trustedProxies []string) string {
	// Helper function to check if IP is in trusted proxies
	isTrustedProxy := func(ip string) bool {
		for _, trusted := range trustedProxies {
			if ip == trusted {
				return true
			}
			// Check if it's a CIDR range
			if strings.Contains(trusted, "/") {
				_, cidr, err := net.ParseCIDR(trusted)
				if err == nil {
					if parsedIP := net.ParseIP(ip); parsedIP != nil {
						if cidr.Contains(parsedIP) {
							return true
						}
					}
				}
			}
		}
		return false
	}

	// Check X-Forwarded-For header (most common for load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
		ips := strings.Split(xff, ",")

		// Start from the leftmost IP (original client) and work right
		for i := 0; i < len(ips); i++ {
			ip := strings.TrimSpace(ips[i])
			if ip != "" && net.ParseIP(ip) != nil {
				// If this is not a trusted proxy, it's likely the real client IP
				if !isTrustedProxy(ip) {
					return ip
				}
				// If it is a trusted proxy and we're at the leftmost position,
				// this might be the client IP in a trusted proxy setup
				if i == 0 {
					return ip
				}
			}
		}
	}

	// Check X-Real-IP header (used by nginx and others)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if ip := strings.TrimSpace(xri); ip != "" && net.ParseIP(ip) != nil {
			return ip
		}
	}

	// Check X-Client-IP header (less common)
	if xci := r.Header.Get("X-Client-IP"); xci != "" {
		if ip := strings.TrimSpace(xci); ip != "" && net.ParseIP(ip) != nil {
			return ip
		}
	}

	// Check CF-Connecting-IP header (Cloudflare)
	if cfip := r.Header.Get("CF-Connecting-IP"); cfip != "" {
		if ip := strings.TrimSpace(cfip); ip != "" && net.ParseIP(ip) != nil {
			return ip
		}
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If SplitHostPort fails, RemoteAddr might not have a port
		host = r.RemoteAddr
	}

	// Validate and return the IP
	if ip := net.ParseIP(host); ip != nil {
		return host
	}

	// Last resort: return a default IP if parsing fails
	return "unknown"
}

// addRateLimitHeaders adds standard rate limiting headers to the response
func addRateLimitHeaders(w http.ResponseWriter, result *LimitResult) {
	w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(result.Limit, 10))
	w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(result.Remaining, 10))
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetTime.Unix(), 10))

	// Add human-readable reset time
	w.Header().Set("X-RateLimit-Reset-Time", result.ResetTime.Format(time.RFC3339))
}
