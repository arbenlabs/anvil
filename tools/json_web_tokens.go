package tools

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT represents a JSON Web Token service with configuration for token generation and verification.
// This struct encapsulates the issuer information and signing key needed for JWT operations.
// The issuer is typically the domain or service name that creates the token, and the signing key
// is used to sign and verify the token's authenticity.
type JWT struct {
	Issuer     string `json:"issuer"`      // The issuer of the JWT (typically your service domain)
	SigningKey []byte `json:"signing_key"` // The secret key used to sign and verify tokens
}

// JWTClaims represents the custom claims structure for JSON Web Tokens.
// This struct defines the user-specific data that will be embedded in the JWT.
// The claims are included in the token payload and can be extracted during verification.
type JWTClaims struct {
	ID    string `json:"user_id"` // The unique identifier of the user
	Email string `json:"email"`   // The email address of the user
}

// NewJsonWebToken creates a new JWT service instance with the specified issuer and signing key.
// This function initializes a JWT service that can be used to generate and verify tokens.
// The issuer should be a unique identifier for your service (e.g., "myapp.com"),
// and the signing key should be a secure, randomly generated secret.
//
// Example usage:
//
//	jwtService := NewJsonWebToken("myapp.com", []byte("your-secret-key"))
//	token, err := jwtService.Generate(JWTClaims{ID: "123", Email: "user@example.com"}, nil)
//
// Parameters:
//   - issuer: The issuer identifier for the JWT (typically your service domain)
//   - key: The secret key used to sign and verify tokens (should be at least 32 bytes)
//
// Returns:
//   - *JWT: A new JWT service instance
func NewJsonWebToken(issuer string, key []byte) *JWT {
	return &JWT{
		Issuer:     issuer,
		SigningKey: key,
	}
}

// Generate creates a new JSON Web Token with the specified claims and expiration.
// This function creates a JWT using the HS256 signing algorithm with the configured
// issuer and signing key. The token includes standard JWT claims (exp, iat, nbf, iss, sub, jti)
// along with the custom user claims.
//
// The expiration parameter is optional:
//   - If nil, the token expires in 15 minutes
//   - If >= 0, the token expires in the specified number of minutes
//
// The generated token includes the following claims:
//   - exp: Expiration time
//   - iat: Issued at time
//   - nbf: Not before time
//   - iss: Issuer (from JWT configuration)
//   - sub: Subject (user's email)
//   - jti: JWT ID (user's ID)
//
// Example usage:
//
//	claims := JWTClaims{ID: "user123", Email: "user@example.com"}
//	token, err := jwtService.Generate(claims, nil) // 15 minute expiration
//	token, err := jwtService.Generate(claims, &30) // 30 minute expiration
//
// Parameters:
//   - claims: The user-specific claims to include in the token
//   - expiration: Optional expiration time in minutes (nil for 15 minutes default)
//
// Returns:
//   - string: The signed JWT string
//   - error: Any error that occurred during token generation
func (tkn *JWT) Generate(claims JWTClaims, expiration *int) (string, error) {
	var tokenExpiration time.Duration

	if expiration == nil {
		tokenExpiration = 15 * time.Minute
	}

	if *expiration >= 0 {
		tokenExpiration = time.Duration(*expiration) * time.Minute
	}

	jwtClaims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiration)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    tkn.Issuer,
		Subject:   claims.Email,
		ID:        claims.ID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	ss, err := token.SignedString(tkn.SigningKey)
	if err != nil {
		return "", err
	}

	return ss, nil
}

// Verify validates a JSON Web Token and extracts the user claims.
// This function verifies the token's signature using the configured signing key
// and extracts the user claims if the token is valid. It checks for:
//   - Valid signature using HS256 algorithm
//   - Token expiration
//   - Token not-before time
//   - Issuer validation
//
// The function returns the user claims (ID and email) if the token is valid,
// or an error if the token is invalid, expired, or malformed.
//
// Example usage:
//
//	claims, err := jwtService.Verify(tokenString)
//	if err != nil {
//	    // Token is invalid, expired, or malformed
//	}
//	// Use claims.ID and claims.Email
//
// Parameters:
//   - tokenString: The JWT string to verify
//
// Returns:
//   - JWTClaims: The user claims extracted from the token (ID and email)
//   - error: Any error that occurred during verification (invalid signature, expired, etc.)
func (tkn *JWT) Verify(tokenString string) (JWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return tkn.SigningKey, nil
	})
	if err != nil {
		return JWTClaims{}, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return JWTClaims{
			ID:    fmt.Sprint(claims["jti"]),
			Email: fmt.Sprint(claims["sub"]),
		}, nil
	}

	return JWTClaims{}, errors.New("token claims not found")
}
