package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type LoadTestConfig struct {
	BaseURL       string
	Endpoint      string
	Method        string
	Requests      int
	Concurrent    int
	Duration      time.Duration
	TestRateLimit bool
	AuthToken     string
	Body          string
	Headers       map[string]string
}

type TestResult struct {
	TotalRequests       int64
	SuccessfulRequests  int64
	RateLimitedRequests int64
	ErrorRequests       int64
	AverageLatency      time.Duration
	MaxLatency          time.Duration
	MinLatency          time.Duration
	RequestsPerSecond   float64
	RateLimitHeaders    map[string]string
}

type RequestResult struct {
	StatusCode  int
	Latency     time.Duration
	RateLimited bool
	Error       error
	Headers     map[string]string
}

func main() {
	var config LoadTestConfig

	// Command line flags
	flag.StringVar(&config.BaseURL, "base-url", "http://localhost:8080", "Base URL of the API")
	flag.StringVar(&config.Endpoint, "endpoint", "/api/health", "API endpoint to test")
	flag.StringVar(&config.Method, "method", "GET", "HTTP method")
	flag.IntVar(&config.Requests, "requests", 100, "Total number of requests")
	flag.IntVar(&config.Concurrent, "concurrent", 10, "Number of concurrent requests")
	flag.DurationVar(&config.Duration, "duration", 0, "Test duration (0 for request-based)")
	flag.BoolVar(&config.TestRateLimit, "test-rate-limit", false, "Test rate limiting behavior")
	flag.StringVar(&config.AuthToken, "auth-token", "", "Authorization token for protected endpoints")
	flag.StringVar(&config.Body, "body", "", "Request body (JSON)")

	flag.Parse()

	// Initialize headers
	config.Headers = make(map[string]string)
	if config.AuthToken != "" {
		config.Headers["Authorization"] = "Bearer " + config.AuthToken
	}
	if config.Body != "" {
		config.Headers["Content-Type"] = "application/json"
	}

	fmt.Printf("🚀 Starting Load Test\n")
	fmt.Printf("Target: %s%s\n", config.BaseURL, config.Endpoint)
	fmt.Printf("Method: %s\n", config.Method)
	fmt.Printf("Requests: %d\n", config.Requests)
	fmt.Printf("Concurrent: %d\n", config.Concurrent)
	if config.Duration > 0 {
		fmt.Printf("Duration: %v\n", config.Duration)
	}
	fmt.Printf("Rate Limit Test: %v\n", config.TestRateLimit)
	fmt.Println()

	// Run the load test
	result := runLoadTest(config)

	// Print results
	printResults(result)

	// Run specific rate limit tests if requested
	if config.TestRateLimit {
		fmt.Println("\n🔒 Running Rate Limit Specific Tests...")
		runRateLimitTests(config)
	}
}

func runLoadTest(config LoadTestConfig) TestResult {
	var result TestResult
	var wg sync.WaitGroup
	var totalLatency int64
	var maxLatency int64
	var minLatency int64 = int64(time.Hour) // Initialize to a large value

	resultChan := make(chan RequestResult, config.Requests)

	// Semaphore to control concurrency
	semaphore := make(chan struct{}, config.Concurrent)

	startTime := time.Now()

	// Determine if we're running duration-based or request-based test
	if config.Duration > 0 {
		// Duration-based test
		endTime := startTime.Add(config.Duration)

		for time.Now().Before(endTime) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				semaphore <- struct{}{}        // Acquire
				defer func() { <-semaphore }() // Release

				reqResult := makeRequest(config)
				resultChan <- reqResult
			}()
		}
	} else {
		// Request-based test
		for i := 0; i < config.Requests; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				semaphore <- struct{}{}        // Acquire
				defer func() { <-semaphore }() // Release

				reqResult := makeRequest(config)
				resultChan <- reqResult
			}()
		}
	}

	// Close result channel when all requests are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for reqResult := range resultChan {
		atomic.AddInt64(&result.TotalRequests, 1)

		if reqResult.Error != nil {
			atomic.AddInt64(&result.ErrorRequests, 1)
		} else if reqResult.RateLimited {
			atomic.AddInt64(&result.RateLimitedRequests, 1)
			// Store rate limit headers from the first rate limited response
			if len(result.RateLimitHeaders) == 0 {
				result.RateLimitHeaders = reqResult.Headers
			}
		} else if reqResult.StatusCode >= 200 && reqResult.StatusCode < 300 {
			atomic.AddInt64(&result.SuccessfulRequests, 1)
		} else {
			atomic.AddInt64(&result.ErrorRequests, 1)
		}

		// Update latency statistics
		latencyNs := reqResult.Latency.Nanoseconds()
		atomic.AddInt64(&totalLatency, latencyNs)

		// Update max latency
		for {
			currentMax := atomic.LoadInt64(&maxLatency)
			if latencyNs <= currentMax || atomic.CompareAndSwapInt64(&maxLatency, currentMax, latencyNs) {
				break
			}
		}

		// Update min latency
		for {
			currentMin := atomic.LoadInt64(&minLatency)
			if latencyNs >= currentMin || atomic.CompareAndSwapInt64(&minLatency, currentMin, latencyNs) {
				break
			}
		}
	}

	totalDuration := time.Since(startTime)

	// Calculate final statistics
	if result.TotalRequests > 0 {
		result.AverageLatency = time.Duration(totalLatency / result.TotalRequests)
		result.RequestsPerSecond = float64(result.TotalRequests) / totalDuration.Seconds()
	}
	result.MaxLatency = time.Duration(maxLatency)
	result.MinLatency = time.Duration(minLatency)

	return result
}

func makeRequest(config LoadTestConfig) RequestResult {
	var reqResult RequestResult
	startTime := time.Now()

	// Prepare request body
	var bodyReader io.Reader
	if config.Body != "" {
		bodyReader = bytes.NewBufferString(config.Body)
	}

	// Create request
	req, err := http.NewRequest(config.Method, config.BaseURL+config.Endpoint, bodyReader)
	if err != nil {
		reqResult.Error = err
		reqResult.Latency = time.Since(startTime)
		return reqResult
	}

	// Add headers
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	reqResult.Latency = time.Since(startTime)

	if err != nil {
		reqResult.Error = err
		return reqResult
	}
	defer resp.Body.Close()

	reqResult.StatusCode = resp.StatusCode
	reqResult.RateLimited = resp.StatusCode == 429

	// Capture rate limit headers
	reqResult.Headers = make(map[string]string)
	rateLimitHeaders := []string{
		"X-RateLimit-Limit",
		"X-RateLimit-Remaining",
		"X-RateLimit-Reset",
		"X-RateLimit-Reset-Time",
		"Retry-After",
	}

	for _, header := range rateLimitHeaders {
		if value := resp.Header.Get(header); value != "" {
			reqResult.Headers[header] = value
		}
	}

	return reqResult
}

func printResults(result TestResult) {
	fmt.Printf("📊 Load Test Results\n")
	fmt.Printf("==================\n")
	fmt.Printf("Total Requests:      %d\n", result.TotalRequests)
	fmt.Printf("Successful:          %d (%.1f%%)\n", result.SuccessfulRequests,
		float64(result.SuccessfulRequests)/float64(result.TotalRequests)*100)
	fmt.Printf("Rate Limited:        %d (%.1f%%)\n", result.RateLimitedRequests,
		float64(result.RateLimitedRequests)/float64(result.TotalRequests)*100)
	fmt.Printf("Errors:              %d (%.1f%%)\n", result.ErrorRequests,
		float64(result.ErrorRequests)/float64(result.TotalRequests)*100)
	fmt.Printf("Requests/Second:     %.2f\n", result.RequestsPerSecond)
	fmt.Printf("Average Latency:     %v\n", result.AverageLatency)
	fmt.Printf("Min Latency:         %v\n", result.MinLatency)
	fmt.Printf("Max Latency:         %v\n", result.MaxLatency)

	if len(result.RateLimitHeaders) > 0 {
		fmt.Printf("\n🔒 Rate Limit Headers (from first rate limited response):\n")
		for header, value := range result.RateLimitHeaders {
			fmt.Printf("  %s: %s\n", header, value)
		}
	}
}

func runRateLimitTests(config LoadTestConfig) {
	fmt.Println("\n1. Testing Rate Limit Threshold...")
	testRateLimitThreshold(config)

	fmt.Println("\n2. Testing Rate Limit Recovery...")
	testRateLimitRecovery(config)

	fmt.Println("\n3. Testing Burst Traffic...")
	testBurstTraffic(config)
}

func testRateLimitThreshold(config LoadTestConfig) {
	// Send requests one by one until rate limited
	client := &http.Client{Timeout: 10 * time.Second}

	for i := 1; i <= 50; i++ { // Max 50 attempts to find the limit
		req, err := http.NewRequest(config.Method, config.BaseURL+config.Endpoint, nil)
		if err != nil {
			log.Printf("Error creating request: %v", err)
			return
		}

		// Add headers
		for key, value := range config.Headers {
			req.Header.Set(key, value)
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error making request: %v", err)
			return
		}
		resp.Body.Close()

		fmt.Printf("Request %d: Status %d", i, resp.StatusCode)

		if limit := resp.Header.Get("X-RateLimit-Limit"); limit != "" {
			remaining := resp.Header.Get("X-RateLimit-Remaining")
			fmt.Printf(" (Limit: %s, Remaining: %s)", limit, remaining)
		}

		if resp.StatusCode == 429 {
			retryAfter := resp.Header.Get("Retry-After")
			fmt.Printf(" - RATE LIMITED! Retry after: %s seconds", retryAfter)
			fmt.Printf("\n✅ Rate limit threshold found at request %d\n", i)
			return
		}

		fmt.Println()
		time.Sleep(100 * time.Millisecond) // Small delay between requests
	}

	fmt.Println("❌ Rate limit threshold not reached within 50 requests")
}

func testRateLimitRecovery(config LoadTestConfig) {
	client := &http.Client{Timeout: 10 * time.Second}

	// First, trigger rate limiting
	fmt.Println("Triggering rate limit...")
	for i := 0; i < 20; i++ {
		req, _ := http.NewRequest(config.Method, config.BaseURL+config.Endpoint, nil)
		for key, value := range config.Headers {
			req.Header.Set(key, value)
		}

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == 429 {
			retryAfter := resp.Header.Get("Retry-After")
			fmt.Printf("Rate limited! Retry after: %s seconds\n", retryAfter)

			// Wait for the retry period
			if retryAfter != "" {
				fmt.Printf("Waiting for rate limit to reset...\n")
				time.Sleep(65 * time.Second) // Wait a bit longer than the retry period

				// Test recovery
				req, _ := http.NewRequest(config.Method, config.BaseURL+config.Endpoint, nil)
				for key, value := range config.Headers {
					req.Header.Set(key, value)
				}

				resp, err := client.Do(req)
				if err != nil {
					fmt.Printf("❌ Error testing recovery: %v\n", err)
					return
				}
				resp.Body.Close()

				if resp.StatusCode == 429 {
					fmt.Println("❌ Rate limit did not reset properly")
				} else {
					fmt.Println("✅ Rate limit reset successfully")
				}
				return
			}
		}
	}

	fmt.Println("❌ Could not trigger rate limit for recovery test")
}

func testBurstTraffic(config LoadTestConfig) {
	fmt.Println("Testing burst traffic pattern...")

	// Send a burst of requests
	burstSize := 20
	var wg sync.WaitGroup
	results := make(chan RequestResult, burstSize)

	startTime := time.Now()

	for i := 0; i < burstSize; i++ {
		wg.Add(1)
		go func(requestNum int) {
			defer wg.Done()

			reqResult := makeRequest(config)
			reqResult.Headers["RequestNum"] = fmt.Sprintf("%d", requestNum)
			results <- reqResult
		}(i + 1)
	}

	wg.Wait()
	close(results)

	burstDuration := time.Since(startTime)

	var successful, rateLimited, errors int
	var firstRateLimit int = -1

	for result := range results {
		if result.Error != nil {
			errors++
		} else if result.RateLimited {
			rateLimited++
			if firstRateLimit == -1 {
				if reqNum := result.Headers["RequestNum"]; reqNum != "" {
					fmt.Sscanf(reqNum, "%d", &firstRateLimit)
				}
			}
		} else if result.StatusCode >= 200 && result.StatusCode < 300 {
			successful++
		} else {
			errors++
		}
	}

	fmt.Printf("Burst Results (Duration: %v):\n", burstDuration)
	fmt.Printf("  Successful: %d\n", successful)
	fmt.Printf("  Rate Limited: %d\n", rateLimited)
	fmt.Printf("  Errors: %d\n", errors)

	if firstRateLimit > 0 {
		fmt.Printf("  First rate limit at request: %d\n", firstRateLimit)
	}

	if rateLimited > 0 {
		fmt.Println("✅ Burst traffic properly rate limited")
	} else {
		fmt.Println("⚠️  No rate limiting detected in burst traffic")
	}
}
