package anvil

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/clerkinc/clerk-sdk-go/clerk"
	"golang.org/x/time/rate"
)

// Message represents a standardized error response structure for rate limiting.
// This struct is used to provide consistent error messages when rate limits are exceeded.
// It includes status information, a descriptive message, a locked flag, and a timestamp
// for debugging and monitoring purposes.
type Message struct {
	Status    string    `json:"status"`    // The status of the request (e.g., "Request Failed")
	Body      string    `json:"body"`      // The error message body
	Locked    bool      `json:"locked"`    // Whether the request is locked due to rate limiting
	Timestamp time.Time `json:"timestamp"` // When the rate limit was triggered
}

// RateLimit is a type alias for rate.Limiter to provide semantic meaning.
// This type represents a rate limiter that controls the frequency of requests
// based on the configured rate and burst limits.
type RateLimit *rate.Limiter

var (
	// RateLimitPublicAPI provides rate limiting for public API endpoints.
	// This limiter allows 5000 requests per second with a burst capacity of 100 requests.
	// Suitable for public-facing APIs that need to handle high traffic while preventing abuse.
	RateLimitPublicAPI RateLimit = rate.NewLimiter(5000, 100)

	// RateLimitInternalAPI provides rate limiting for internal API endpoints.
	// This limiter allows 10000 requests per second with a burst capacity of 200 requests.
	// Suitable for internal services that need higher throughput than public APIs.
	RateLimitInternalAPI RateLimit = rate.NewLimiter(10000, 200)

	// RateLimitUserWebAPI provides rate limiting for user-facing web APIs.
	// This limiter allows 300 requests per second with a burst capacity of 30 requests.
	// Suitable for web applications where users interact directly with the API.
	RateLimitUserWebAPI RateLimit = rate.NewLimiter(300, 30)

	// RateLimitStrictAPI provides strict rate limiting for sensitive endpoints.
	// This limiter allows 100 requests per second with a burst capacity of 10 requests.
	// Suitable for authentication endpoints, payment processing, or other sensitive operations.
	RateLimitStrictAPI RateLimit = rate.NewLimiter(100, 10)
)

// LoggerMiddleware creates an HTTP middleware that logs request information.
// This middleware extracts and logs the client's IP address, host, server address,
// user agent, and request details (method, path, remote address) for each HTTP request.
// The logging is done using the structured logging package (slog) for better log parsing.
//
// The middleware logs the following information:
//   - IP address of the client
//   - Host from the request URL
//   - Server address
//   - User agent string
//   - Request method, path, and remote address
//
// Example usage:
//
//	http.Handle("/api", LoggerMiddleware(myHandler))
//
// Parameters:
//   - next: The next HTTP handler in the middleware chain
//
// Returns:
//   - http.Handler: A new handler that logs requests before passing them to the next handler
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the IP address from the request.
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx := r.Context()
		srvAddr := ctx.Value(http.LocalAddrContextKey).(net.Addr)

		slog.Info(
			"ip_address", ip,
			"host", r.URL.Host,
			"server_addr", srvAddr.String(),
			"user_agent", r.UserAgent(),
			fmt.Sprintf("%s - %s (%s)", r.Method, r.URL.Path, r.RemoteAddr),
		)

		next.ServeHTTP(w, r)
	})
}

// RateLimitPublic creates middleware that applies public API rate limiting.
// This middleware uses the RateLimitPublicAPI configuration, which allows
// 5000 requests per second with a burst capacity of 100 requests.
// It's suitable for public-facing endpoints that need to handle high traffic.
//
// The middleware tracks clients by IP address and applies rate limiting per client.
// When a client exceeds the rate limit, it receives a 429 (Too Many Requests) response
// with a JSON error message.
//
// Example usage:
//
//	http.Handle("/api/public", RateLimitPublic(myHandler))
//
// Parameters:
//   - next: The next HTTP handler in the middleware chain
//
// Returns:
//   - http.Handler: A new handler that applies public API rate limiting
func RateLimitPublic(next http.Handler) http.Handler {
	return rateLimiterMiddleware(next, RateLimitPublicAPI)
}

// RateLimitInternal creates middleware that applies internal API rate limiting.
// This middleware uses the RateLimitInternalAPI configuration, which allows
// 10000 requests per second with a burst capacity of 200 requests.
// It's suitable for internal service-to-service communication.
//
// The middleware tracks clients by IP address and applies rate limiting per client.
// When a client exceeds the rate limit, it receives a 429 (Too Many Requests) response
// with a JSON error message.
//
// Example usage:
//
//	http.Handle("/api/internal", RateLimitInternal(myHandler))
//
// Parameters:
//   - next: The next HTTP handler in the middleware chain
//
// Returns:
//   - http.Handler: A new handler that applies internal API rate limiting
func RateLimitInternal(next http.Handler) http.Handler {
	return rateLimiterMiddleware(next, RateLimitInternalAPI)
}

// RateLimitWeb creates middleware that applies user web API rate limiting.
// This middleware uses the RateLimitUserWebAPI configuration, which allows
// 300 requests per second with a burst capacity of 30 requests.
// It's suitable for web applications where users interact directly with the API.
//
// The middleware tracks clients by IP address and applies rate limiting per client.
// When a client exceeds the rate limit, it receives a 429 (Too Many Requests) response
// with a JSON error message.
//
// Example usage:
//
//	http.Handle("/api/web", RateLimitWeb(myHandler))
//
// Parameters:
//   - next: The next HTTP handler in the middleware chain
//
// Returns:
//   - http.Handler: A new handler that applies user web API rate limiting
func RateLimitWeb(next http.Handler) http.Handler {
	return rateLimiterMiddleware(next, RateLimitUserWebAPI)
}

// RateLimitStrict creates middleware that applies strict API rate limiting.
// This middleware uses the RateLimitStrictAPI configuration, which allows
// 100 requests per second with a burst capacity of 10 requests.
// It's suitable for sensitive endpoints like authentication or payment processing.
//
// The middleware tracks clients by IP address and applies rate limiting per client.
// When a client exceeds the rate limit, it receives a 429 (Too Many Requests) response
// with a JSON error message.
//
// Example usage:
//
//	http.Handle("/api/auth", RateLimitStrict(myHandler))
//
// Parameters:
//   - next: The next HTTP handler in the middleware chain
//
// Returns:
//   - http.Handler: A new handler that applies strict API rate limiting
func RateLimitStrict(next http.Handler) http.Handler {
	return rateLimiterMiddleware(next, RateLimitStrictAPI)
}

// rateLimiterMiddleware is the internal implementation of rate limiting middleware.
// This function creates a rate limiter that tracks clients by IP address and applies
// the specified rate limit configuration. It includes automatic cleanup of old client
// entries to prevent memory leaks.
//
// The middleware:
//   - Tracks clients by their IP address
//   - Applies rate limiting per client using the provided rate limiter
//   - Automatically cleans up client entries older than 5 minutes
//   - Returns a 429 status with a JSON error message when rate limits are exceeded
//
// Parameters:
//   - next: The next HTTP handler in the middleware chain
//   - rateLimit: The rate limiter configuration to apply
//
// Returns:
//   - http.Handler: A new handler that applies the specified rate limiting
func rateLimiterMiddleware(next http.Handler, rateLimit RateLimit) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)
	go func() {
		for {
			time.Sleep(time.Minute)
			// Lock the mutex to protect this section from race conditions.
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 5*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the IP address from the request.
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Lock the mutex to protect this section from race conditions.
		mu.Lock()
		if _, found := clients[ip]; !found {
			clients[ip] = &client{limiter: rateLimit}
		}
		clients[ip].lastSeen = time.Now()
		if !clients[ip].limiter.Allow() {
			mu.Unlock()

			message := Message{
				Status:    "Request Failed",
				Body:      "Rate limit reached. Please wait 5 minutes and try again.",
				Locked:    true,
				Timestamp: time.Now(),
			}

			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(&message)
			return
		}
		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

func ClerkAuthMiddleware(clerk clerk.Client) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			type SessionName string
			var ClerkSessionName SessionName = "clerksession"

			// Get the session token from the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// The token should be in the format "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
				return
			}

			sessionToken := parts[1]

			// Verify the session
			session, err := clerk.VerifyToken(sessionToken)
			if err != nil {
				http.Error(w, "Invalid session", http.StatusUnauthorized)
				return
			}

			// Add the session to the request context
			ctx := context.WithValue(r.Context(), ClerkSessionName, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ClerkWebhookMiddleware(clerk clerk.Client, secret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Retrieve the signing secret from environment variables
			signingSecret := secret
			if signingSecret == "" {
				http.Error(w, "Webhook signing secret not configured", http.StatusInternalServerError)
				return
			}

			// Extract the signature from headers
			signature := r.Header.Get("X-Clerk-Signature")
			if signature == "" {
				http.Error(w, "Invalid clerk webhook signature: ", http.StatusUnauthorized)
				return
			}

			// Verify the webhook signature
			if signingSecret != signature {
				http.Error(w, "Invalid webhook signature: ", http.StatusUnauthorized)
				return
			}

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}
