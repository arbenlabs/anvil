package anvil

import (
	"encoding/json"
	"net/http"
	"time"
)

// APIFunc represents a function signature for HTTP handlers that return errors.
// This type is used to standardize error handling across all API endpoints.
// Functions implementing this signature should handle the HTTP request and return
// an error if something goes wrong, which will be automatically converted to
// a JSON error response.
type APIFunc func(http.ResponseWriter, *http.Request) error

// HandlerFunc converts an APIFunc to a standard http.HandlerFunc with automatic error handling.
// This function wraps API handlers to provide consistent error response formatting.
// When the wrapped function returns an error, it automatically calls RespondWithError
// to send a properly formatted JSON error response to the client.
//
// Example usage:
//
//	http.HandleFunc("/api/users", HandlerFunc(createUserHandler))
//
// Parameters:
//   - f: The APIFunc to wrap with error handling
//
// Returns:
//   - http.HandlerFunc: A standard HTTP handler function with error handling
func HandlerFunc(f APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			RespondWithError(w, err)
		}
	}
}

// writeJSON is a helper function that writes JSON data to an HTTP response.
// It sets the appropriate Content-Type header and writes the response with the given status code.
// This function is used internally by RespondWithError and RespondWithSuccess.
//
// Parameters:
//   - w: The HTTP response writer
//   - status: The HTTP status code to return
//   - v: The data to encode as JSON
//
// Returns:
//   - error: Any error that occurred during JSON encoding or writing
func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// RespondWithError sends a JSON error response to the client.
// This function formats the error message and includes a timestamp in the response.
// It automatically sets the HTTP status code to 400 (Bad Request) and the
// Content-Type header to application/json.
//
// The error response follows this structure:
//
//	{
//	  "error": "error message here",
//	  "timestamp": "2024-01-01 12:00:00 +0000 UTC"
//	}
//
// Parameters:
//   - w: The HTTP response writer
//   - e: The error to format and send
//
// Returns:
//   - error: Any error that occurred during response writing
func RespondWithError(w http.ResponseWriter, e error) error {
	return writeJSON(w, http.StatusBadRequest, formatError(e))
}

// RespondWithSuccess sends a JSON success response to the client.
// This function sends the provided data as JSON with the specified HTTP status code.
// It automatically sets the Content-Type header to application/json.
//
// Parameters:
//   - w: The HTTP response writer
//   - status: The HTTP status code to return (e.g., 200, 201, 204)
//   - v: The data to encode as JSON in the response body
//
// Returns:
//   - error: Any error that occurred during JSON encoding or writing
func RespondWithSuccess(w http.ResponseWriter, status int, v any) error {
	return writeJSON(w, status, v)
}

// formatError creates a standardized error response structure.
// This function takes an error and formats it into a map with an error message
// and a timestamp. The timestamp is useful for debugging and logging purposes.
//
// Parameters:
//   - err: The error to format
//
// Returns:
//   - map[string]string: A map containing the error message and timestamp
func formatError(err error) map[string]string {
	var handlerError = err.Error()

	return map[string]string{
		"error":     handlerError,
		"timestamp": time.Now().String(),
	}
}
