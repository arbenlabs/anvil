package tools

import (
	"time"

	"github.com/google/uuid"
)

// GenerateNamespaceUUID creates a UUID with a namespace suffix.
// This function generates a standard UUID and appends the provided namespace
// to create a namespaced identifier. This is useful for creating unique
// identifiers that are scoped to a specific context or domain.
//
// The generated UUID follows the standard UUID v4 format, and the namespace
// is appended with a hyphen separator. This creates identifiers that are
// both globally unique and contextually meaningful.
//
// Example usage:
//
//	userID := GenerateNamespaceUUID("user")
//	// Result: "550e8400-e29b-41d4-a716-446655440000-user"
//
//	orderID := GenerateNamespaceUUID("order")
//	// Result: "550e8400-e29b-41d4-a716-446655440000-order"
//
// Parameters:
//   - namespace: The namespace to append to the UUID
//
// Returns:
//   - string: A namespaced UUID string
func GenerateNamespaceUUID(namespace string) string {
	uuid := uuid.NewString()
	return uuid + "-" + namespace
}

// GenerateUUID creates a new random UUID (Universally Unique Identifier).
// This function generates a version 4 UUID using the crypto/rand package,
// ensuring cryptographic randomness and uniqueness. UUIDs are useful for
// creating globally unique identifiers for database records, API endpoints,
// and other resources that need to be uniquely identified.
//
// The generated UUID follows the standard format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
// where x is a hexadecimal digit and y is one of 8, 9, A, or B.
//
// Example usage:
//
//	id := GenerateUUID()
//	// Result: "550e8400-e29b-41d4-a716-446655440000"
//
// Returns:
//   - string: A new random UUID string
func GenerateUUID() string {
	return uuid.NewString()
}

// GetCurrentDate returns the current date at midnight UTC.
// This function returns a time.Time value representing the current date
// with the time set to 00:00:00 UTC. This is useful for date-based
// operations where you need to work with dates without time components,
// such as date ranges, daily statistics, or date-based filtering.
//
// The function extracts the year, month, and day from the current time
// and creates a new time.Time value with those components and zero time.
//
// Example usage:
//
//	today := GetCurrentDate()
//	// Result: 2024-01-15 00:00:00 +0000 UTC
//
// Returns:
//   - time.Time: The current date at midnight UTC
func GetCurrentDate() time.Time {
	currentTime := time.Now()
	yr := currentTime.Year()
	mo := currentTime.Month()
	dy := currentTime.Day()

	date := time.Date(yr, mo, dy, 0, 0, 0, 0, time.UTC)
	return date
}

// GetFutureDate calculates a future date based on the specified offsets.
// This function adds the specified number of years, months, and days to
// the current date and returns the resulting future date. This is useful
// for calculating expiration dates, subscription end dates, or other
// future time-based events.
//
// The function uses time.AddDate which properly handles month and year
// boundaries, including leap years and varying month lengths.
//
// Example usage:
//
//	// Get date 1 year, 6 months, and 15 days from now
//	futureDate := GetFutureDate(1, 6, 15)
//	// Result: Current date + 1 year + 6 months + 15 days
//
//	// Get date 2 years from now
//	twoYearsFromNow := GetFutureDate(2, 0, 0)
//
// Parameters:
//   - years: Number of years to add (can be negative for past dates)
//   - months: Number of months to add (can be negative)
//   - days: Number of days to add (can be negative)
//
// Returns:
//   - time.Time: The calculated future date
func GetFutureDate(years int, months int, days int) time.Time {
	currentTime := time.Now()
	t := currentTime.AddDate(years, months, days)
	return t
}

// SafeString safely extracts a string value from a map[string]interface{}.
// This function provides type-safe access to string values in maps that
// contain mixed types (interface{}). It handles cases where the key
// doesn't exist, the value is nil, or the value is not a string type.
//
// The function returns an empty string if:
//   - The key doesn't exist in the map
//   - The value is nil
//   - The value is not a string type
//
// This is useful when working with JSON data, database results, or
// any map that contains mixed data types and you need to safely
// extract string values without causing panics.
//
// Example usage:
//
//	data := map[string]interface{}{
//	    "name": "John",
//	    "age":  30,
//	    "email": nil,
//	}
//	name := SafeString(data, "name")     // Returns: "John"
//	email := SafeString(data, "email")   // Returns: ""
//	missing := SafeString(data, "missing") // Returns: ""
//
// Parameters:
//   - data: The map containing mixed data types
//   - key: The key to look up in the map
//
// Returns:
//   - string: The string value if found and valid, empty string otherwise
func SafeString(data map[string]interface{}, key string) string {
	if value, ok := data[key]; ok && value != nil {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return "" // Return a default empty string if missing or invalid
}

// SafeTime safely extracts a time.Time value from a map[string]interface{}.
// This function provides type-safe access to time.Time values in maps that
// contain mixed types (interface{}). It handles cases where the key
// doesn't exist, the value is nil, or the value is not a time.Time type.
//
// The function returns the zero value of time.Time if:
//   - The key doesn't exist in the map
//   - The value is nil
//   - The value is not a time.Time type
//
// This is useful when working with JSON data, database results, or
// any map that contains mixed data types and you need to safely
// extract time values without causing panics.
//
// Example usage:
//
//	data := map[string]interface{}{
//	    "created_at": time.Now(),
//	    "updated_at": nil,
//	    "name": "John",
//	}
//	createdAt := SafeTime(data, "created_at") // Returns: actual time
//	updatedAt := SafeTime(data, "updated_at") // Returns: zero time
//	missing := SafeTime(data, "missing")      // Returns: zero time
//
// Parameters:
//   - data: The map containing mixed data types
//   - key: The key to look up in the map
//
// Returns:
//   - time.Time: The time.Time value if found and valid, zero time otherwise
func SafeTime(data map[string]interface{}, key string) time.Time {
	if value, ok := data[key]; ok && value != nil {
		if timeValue, ok := value.(time.Time); ok {
			return timeValue
		}
	}
	return time.Time{} // Return the zero value of time.Time if missing or invalid
}

// SafeBool safely extracts a boolean value from a map[string]interface{}.
// This function provides type-safe access to boolean values in maps that
// contain mixed types (interface{}). It handles cases where the key
// doesn't exist, the value is nil, or the value is not a boolean type.
//
// The function returns false if:
//   - The key doesn't exist in the map
//   - The value is nil
//   - The value is not a boolean type
//
// This is useful when working with JSON data, database results, or
// any map that contains mixed data types and you need to safely
// extract boolean values without causing panics.
//
// Example usage:
//
//	data := map[string]interface{}{
//	    "active": true,
//	    "verified": false,
//	    "email": nil,
//	    "name": "John",
//	}
//	active := SafeBool(data, "active")     // Returns: true
//	verified := SafeBool(data, "verified") // Returns: false
//	email := SafeBool(data, "email")       // Returns: false
//	missing := SafeBool(data, "missing")   // Returns: false
//
// Parameters:
//   - data: The map containing mixed data types
//   - key: The key to look up in the map
//
// Returns:
//   - bool: The boolean value if found and valid, false otherwise
func SafeBool(data map[string]interface{}, key string) bool {
	if value, ok := data[key]; ok && value != nil {
		if boolValue, ok := value.(bool); ok {
			return boolValue
		}
	}
	return false
}
