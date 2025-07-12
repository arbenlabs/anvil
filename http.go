package anvil

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rs/cors"
)

const (
	// DefaultShutdownGracePeriod is the default duration for graceful server shutdown.
	// This allows existing connections to finish before the server terminates.
	DefaultShutdownGracePeriod = time.Second * 30

	// DefaultWriteTimeout is the default maximum duration for writing the entire request.
	// This prevents slow clients from consuming server resources indefinitely.
	DefaultWriteTimeout = time.Second * 30

	// DefaultReadTimeout is the default maximum duration for reading the entire request.
	// This prevents slow clients from consuming server resources indefinitely.
	DefaultReadTimeout = time.Second * 15

	// DefaultIdleTimeout is the default maximum amount of time to wait for the next request.
	// This helps manage connection pooling and resource utilization.
	DefaultIdleTimeout = time.Second * 120
)

// AllowedMethods represents HTTP methods that are allowed in CORS configuration.
// This type provides type safety when specifying allowed HTTP methods.
type AllowedMethods string

const (
	// AllowedMethodsGet represents the HTTP GET method for CORS configuration.
	AllowedMethodsGet AllowedMethods = "GET"

	// AllowedMethodsPost represents the HTTP POST method for CORS configuration.
	AllowedMethodsPost AllowedMethods = "POST"

	// AllowedMethodsPatch represents the HTTP PATCH method for CORS configuration.
	AllowedMethodsPatch AllowedMethods = "PATCH"

	// AllowedMethodsPut represents the HTTP PUT method for CORS configuration.
	AllowedMethodsPut AllowedMethods = "PUT"

	// AllowedMethodsDelete represents the HTTP DELETE method for CORS configuration.
	AllowedMethodsDelete AllowedMethods = "DELETE"
)

// HTTPServer represents a configurable HTTP server with timeout settings.
// This struct provides a builder pattern for creating HTTP servers with
// customizable timeout configurations and graceful shutdown capabilities.
type HTTPServer struct {
	Address      string        // The server address (e.g., ":8080")
	WriteTimeout time.Duration // Maximum duration for writing the entire request
	ReadTimeout  time.Duration // Maximum duration for reading the entire request
	IdleTimeout  time.Duration // Maximum amount of time to wait for the next request
	Handler      http.Handler  // The HTTP handler to serve requests
}

// NewServer creates a new HTTPServer instance with default timeout settings.
// This function initializes a server with sensible defaults for production use.
// The address parameter should be just the port number (e.g., "8080"), and it will
// be automatically formatted as ":port".
//
// Example usage:
//
//	server := NewServer("8080")
//	server.Start(context.Background(), "8080")
//
// Parameters:
//   - address: The port number for the server (e.g., "8080")
//
// Returns:
//   - *HTTPServer: A new HTTPServer instance with default settings
func NewServer(address string) *HTTPServer {
	return &HTTPServer{
		Address:      fmt.Sprintf(":%s", address),
		ReadTimeout:  DefaultReadTimeout,
		WriteTimeout: DefaultWriteTimeout,
		IdleTimeout:  DefaultIdleTimeout,
	}
}

// WithWriteTimeout sets the write timeout for the HTTP server.
// This method returns a new HTTPServer instance with the specified write timeout,
// following the builder pattern for configuration.
//
// The write timeout is the maximum duration for writing the entire request,
// including the body. This helps prevent slow clients from consuming server resources.
//
// Parameters:
//   - wto: The write timeout duration
//
// Returns:
//   - *HTTPServer: A new HTTPServer instance with the updated write timeout
func (h *HTTPServer) WithWriteTimeout(wto time.Duration) *HTTPServer {
	return &HTTPServer{
		WriteTimeout: wto,
	}
}

// WithReadTimeout sets the read timeout for the HTTP server.
// This method returns a new HTTPServer instance with the specified read timeout,
// following the builder pattern for configuration.
//
// The read timeout is the maximum duration for reading the entire request,
// including the body. This helps prevent slow clients from consuming server resources.
//
// Parameters:
//   - rto: The read timeout duration
//
// Returns:
//   - *HTTPServer: A new HTTPServer instance with the updated read timeout
func (h *HTTPServer) WithReadTimeout(rto time.Duration) *HTTPServer {
	return &HTTPServer{
		ReadTimeout: rto,
	}
}

// WithIdleTimeout sets the idle timeout for the HTTP server.
// This method returns a new HTTPServer instance with the specified idle timeout,
// following the builder pattern for configuration.
//
// The idle timeout is the maximum amount of time to wait for the next request
// when keep-alives are enabled. This helps manage connection pooling.
//
// Parameters:
//   - ito: The idle timeout duration
//
// Returns:
//   - *HTTPServer: A new HTTPServer instance with the updated idle timeout
func (h *HTTPServer) WithIdleTimeout(ito time.Duration) *HTTPServer {
	return &HTTPServer{
		IdleTimeout: ito,
	}
}

// WithHandler sets the HTTP handler for the server.
// This method returns a new HTTPServer instance with the specified handler,
// following the builder pattern for configuration.
//
// The handler is responsible for processing HTTP requests and generating responses.
// This can be a router, middleware chain, or any http.Handler implementation.
//
// Parameters:
//   - handler: The HTTP handler to use for processing requests
//
// Returns:
//   - *HTTPServer: A new HTTPServer instance with the updated handler
func (h *HTTPServer) WithHandler(handler http.Handler) *HTTPServer {
	return &HTTPServer{
		Handler: handler,
	}
}

// Start begins listening for HTTP requests and handles graceful shutdown.
// This method starts the HTTP server on the configured address and sets up
// graceful shutdown handling. The server will listen for shutdown signals
// through the provided context and gracefully terminate when the context is cancelled.
//
// The server runs in a goroutine and the method blocks until the context is cancelled.
// When shutdown is initiated, the server waits for existing connections to finish
// before terminating, with a configurable grace period.
//
// Example usage:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	server.Start(ctx)
//
// Parameters:
//   - ctx: Context for controlling server lifecycle and shutdown
func (h *HTTPServer) Start(ctx context.Context) {
	server := &http.Server{
		Addr:         h.Address,
		WriteTimeout: h.WriteTimeout,
		ReadTimeout:  h.ReadTimeout,
		IdleTimeout:  h.IdleTimeout,
		Handler:      h.Handler,
	}

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", DefaultShutdownGracePeriod, "duration for which the server gracefully waits for existing connections to finish")
	flag.Parse()

	fmt.Printf("SERVER ADDR", h.Address)

	go func() {
		fmt.Printf("api running on port %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Print(fmt.Errorf("unexpected server error: %v", err))
			panic(err)
		}
	}()

	<-ctx.Done()
	fmt.Print("received shutdown signal, shutting down marketplace service gracefully")

	cx, cancel := context.WithTimeout(ctx, wait)
	defer cancel()

	if err := server.Shutdown(cx); err != nil {
		fmt.Print(fmt.Errorf("error during server shutdown"))
	}

	os.Exit(0)
}

// CORS creates a new CORS middleware with the specified configuration.
// This function creates a CORS handler that can be used to handle Cross-Origin
// Resource Sharing requests. It configures which origins, methods, and credentials
// are allowed for cross-origin requests.
//
// Example usage:
//
//	corsHandler := CORS([]string{"https://example.com"}, []string{"GET", "POST"}, true)
//	handler := PopulateHandlerWithCORS(corsHandler, myHandler)
//
// Parameters:
//   - origins: List of allowed origin URLs (e.g., ["https://example.com"])
//   - methods: List of allowed HTTP methods (e.g., ["GET", "POST", "PUT"])
//   - allowCredentials: Whether to allow credentials in cross-origin requests
//
// Returns:
//   - *cors.Cors: A configured CORS middleware handler
func CORS(origins []string, methods []string, allowCredentials bool) *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   methods,
		AllowCredentials: allowCredentials,
	})
}

// PopulateHandlerWithCORS wraps an HTTP handler with CORS middleware.
// This function applies CORS configuration to an existing HTTP handler,
// enabling cross-origin requests according to the specified CORS settings.
//
// The CORS middleware will be applied to all requests handled by the provided handler,
// adding the necessary CORS headers to responses based on the CORS configuration.
//
// Example usage:
//
//	corsHandler := CORS([]string{"https://example.com"}, []string{"GET", "POST"}, true)
//	finalHandler := PopulateHandlerWithCORS(corsHandler, myHandler)
//	http.Handle("/api", finalHandler)
//
// Parameters:
//   - crossOrigin: The CORS configuration to apply
//   - handler: The HTTP handler to wrap with CORS middleware
//
// Returns:
//   - http.Handler: A new handler with CORS middleware applied
func PopulateHandlerWithCORS(crossOrigin *cors.Cors, handler http.Handler) http.Handler {
	return crossOrigin.Handler(handler)
}
