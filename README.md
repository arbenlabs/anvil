# Anvil

An extra basic Go server toolkit built on top of Gorilla Mux, providing essential components for building robust HTTP servers and APIs.

## Features

- **HTTP Server Management**: Configurable HTTP server with graceful shutdown and timeout controls
- **Middleware Suite**: Built-in logging, rate limiting, and CORS middleware
- **Error Handling**: Standardized JSON error responses with automatic error handling
- **Security Tools**: JWT authentication, Argon2 password hashing, and secure utilities
- **Utility Functions**: UUID generation, date manipulation, and type-safe data extraction

## Installation

```bash
go get github.com/rajisteb/anvil
```

## Quick Start

```go
package main

import (
    "context"
    "net/http"
    
    "github.com/gorilla/mux"
    "rajisteb/anvil"
)

func main() {
    // Create a new router
    router := mux.NewRouter()
    
    // Add middleware
    router.Use(anvil.LoggerMiddleware)
    router.Use(anvil.RateLimitWeb)
    
    // Define your handlers
    router.HandleFunc("/api/health", anvil.HandlerFunc(healthHandler))
    
    // Create and start server
    server := anvil.NewServer("8080").
        WithHandler(router).
        WithWriteTimeout(30 * time.Second)
    
    ctx := context.Background()
    server.Start(ctx)
}

func healthHandler(w http.ResponseWriter, r *http.Request) error {
    return anvil.RespondWithSuccess(w, http.StatusOK, map[string]string{
        "status": "healthy",
    })
}
```

## Core Components

### HTTP Server

The `HTTPServer` provides a configurable HTTP server with graceful shutdown capabilities.

```go
// Create a new server with default settings
server := anvil.NewServer("8080")

// Configure timeouts
server := anvil.NewServer("8080").
    WithReadTimeout(15 * time.Second).
    WithWriteTimeout(30 * time.Second).
    WithIdleTimeout(120 * time.Second)

// Start the server with graceful shutdown
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
server.Start(ctx)
```

### Error Handling

Standardized error handling with automatic JSON response formatting.

```go
// Define handlers that return errors
func createUserHandler(w http.ResponseWriter, r *http.Request) error {
    // Your logic here
    if err := validateUser(r); err != nil {
        return err // Automatically converted to JSON error response
    }
    
    return anvil.RespondWithSuccess(w, http.StatusCreated, user)
}

// Wrap handlers with error handling
http.HandleFunc("/api/users", anvil.HandlerFunc(createUserHandler))
```

### Middleware

#### Logging Middleware

Automatically logs request information including IP address, user agent, and request details.

```go
router.Use(anvil.LoggerMiddleware)
```

#### Rate Limiting

Multiple rate limiting configurations for different use cases:

```go
// Public API rate limiting (5000 req/sec, burst 100)
router.Use(anvil.RateLimitPublic)

// Internal API rate limiting (10000 req/sec, burst 200)
router.Use(anvil.RateLimitInternal)

// Web API rate limiting (300 req/sec, burst 30)
router.Use(anvil.RateLimitWeb)

// Strict rate limiting (100 req/sec, burst 10)
router.Use(anvil.RateLimitStrict)
```

#### CORS

Configure CORS for cross-origin requests:

```go
corsConfig := anvil.CORS(
    []string{"https://yourdomain.com"},
    []string{"GET", "POST", "PUT", "DELETE"},
    true,
)

handler := anvil.PopulateHandlerWithCORS(corsConfig, router)
```

### Security Tools

#### JWT Authentication

```go
import "rajisteb/anvil/tools"

// Create JWT service
jwtService := tools.NewJsonWebToken("your-app.com", []byte("your-secret-key"))

// Generate token
claims := tools.JWTClaims{
    ID:    "user123",
    Email: "user@example.com",
}
token, err := jwtService.Generate(claims, nil) // 15 min default
token, err := jwtService.Generate(claims, &30) // 30 min custom

// Verify token
claims, err := jwtService.Verify(tokenString)
```

#### Password Hashing

Secure password hashing using Argon2id:

```go
import "rajisteb/anvil/tools"

// Hash password
hash, err := tools.GenerateHashString("myPassword123")

// Verify password
match, err := tools.IsMatchingInputAndHash("myPassword123", storedHash)
```

### Utility Functions

#### UUID Generation

```go
import "rajisteb/anvil/tools"

// Generate standard UUID
id := tools.GenerateUUID()

// Generate namespaced UUID
userID := tools.GenerateNamespaceUUID("user")
orderID := tools.GenerateNamespaceUUID("order")
```

#### Date Utilities

```go
import "rajisteb/anvil/tools"

// Get current date at midnight UTC
today := tools.GetCurrentDate()

// Calculate future date
futureDate := tools.GetFutureDate(1, 6, 15) // 1 year, 6 months, 15 days
```

#### Type-Safe Data Extraction

```go
import "rajisteb/anvil/tools"

data := map[string]interface{}{
    "name": "John",
    "age":  30,
    "email": nil,
}

// Safe type extraction
name := tools.SafeString(data, "name")     // "John"
email := tools.SafeString(data, "email")   // ""
age := tools.SafeInt(data, "age")          // 30
```

## API Reference

### HTTPServer

- `NewServer(address string) *HTTPServer` - Create new server with defaults
- `WithReadTimeout(duration) *HTTPServer` - Set read timeout
- `WithWriteTimeout(duration) *HTTPServer` - Set write timeout
- `WithIdleTimeout(duration) *HTTPServer` - Set idle timeout
- `WithHandler(handler) *HTTPServer` - Set HTTP handler
- `Start(ctx context.Context)` - Start server with graceful shutdown

### Error Handling

- `HandlerFunc(APIFunc) http.HandlerFunc` - Wrap handler with error handling
- `RespondWithError(w, err) error` - Send JSON error response
- `RespondWithSuccess(w, status, data) error` - Send JSON success response

### Middleware

- `LoggerMiddleware(next) http.Handler` - Request logging
- `RateLimitPublic(next) http.Handler` - Public API rate limiting
- `RateLimitInternal(next) http.Handler` - Internal API rate limiting
- `RateLimitWeb(next) http.Handler` - Web API rate limiting
- `RateLimitStrict(next) http.Handler` - Strict rate limiting
- `CORS(origins, methods, credentials) *cors.Cors` - CORS configuration

### Tools Package

#### JWT
- `NewJsonWebToken(issuer, key) *JWT` - Create JWT service
- `Generate(claims, expiration) (string, error)` - Generate token
- `Verify(token) (JWTClaims, error)` - Verify token

#### Hashing
- `GenerateHashString(input) (string, error)` - Hash password
- `IsMatchingInputAndHash(input, hash) (bool, error)` - Verify password

#### Utilities
- `GenerateUUID() string` - Generate UUID
- `GenerateNamespaceUUID(namespace) string` - Generate namespaced UUID
- `GetCurrentDate() time.Time` - Get current date
- `GetFutureDate(years, months, days) time.Time` - Calculate future date
- `SafeString(data, key) string` - Safe string extraction
- `SafeInt(data, key) int` - Safe int extraction
- `SafeBool(data, key) bool` - Safe bool extraction
- `SafeTime(data, key) time.Time` - Safe time extraction

## Configuration

### Default Timeouts

- Read Timeout: 15 seconds
- Write Timeout: 30 seconds
- Idle Timeout: 120 seconds
- Graceful Shutdown: 30 seconds

### Rate Limiting Configurations

- **Public API**: 5000 req/sec, burst 100
- **Internal API**: 10000 req/sec, burst 200
- **Web API**: 300 req/sec, burst 30
- **Strict API**: 100 req/sec, burst 10

## Dependencies

- `github.com/gorilla/mux` - HTTP router
- `github.com/golang-jwt/jwt/v5` - JWT handling
- `github.com/rs/cors` - CORS middleware
- `github.com/rs/zerolog` - Structured logging
- `golang.org/x/crypto` - Argon2 hashing
- `golang.org/x/time` - Rate limiting
- `github.com/google/uuid` - UUID generation

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.