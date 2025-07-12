package tools

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// params represents the configuration parameters for Argon2 password hashing.
// This struct contains all the parameters needed to configure the Argon2 algorithm,
// including memory usage, iteration count, parallelism, salt length, and key length.
// These parameters determine the security and performance characteristics of the hash.
type params struct {
	memory      uint32 // Memory usage in KiB (64 * 1024 = 64 MiB)
	iterations  uint32 // Number of iterations (3)
	parallelism uint8  // Number of parallel threads (2)
	saltLength  uint32 // Length of the salt in bytes (16)
	keyLength   uint32 // Length of the derived key in bytes (32)
}

var (
	// errInvalidHash is returned when the encoded hash format is incorrect.
	// This error occurs when the hash string doesn't match the expected Argon2 format.
	errInvalidHash = errors.New("the encoded hash is not in the correct format")

	// errIncompatibleVersion is returned when the Argon2 version is incompatible.
	// This error occurs when trying to verify a hash created with a different
	// version of the Argon2 algorithm.
	errIncompatibleVersion = errors.New("incompatible version of argon2")
)

// GenerateHashString creates a secure hash of the input string using Argon2id.
// This function uses the Argon2id variant, which is recommended for password hashing
// due to its resistance to both GPU-based attacks and side-channel attacks.
//
// The function generates a cryptographically secure random salt and applies
// the Argon2id algorithm with the following parameters:
//   - Memory: 64 MiB (64 * 1024 KiB)
//   - Iterations: 3
//   - Parallelism: 2 threads
//   - Salt length: 16 bytes
//   - Key length: 32 bytes
//
// The returned hash string follows the standard Argon2 format:
//
//	$argon2id$v=19$m=65536,t=3,p=2$[salt]$[hash]
//
// Example usage:
//
//	hash, err := GenerateHashString("myPassword123")
//	if err != nil {
//	    // handle error
//	}
//	// Store hash in database
//
// Parameters:
//   - input: The string to hash (typically a password)
//
// Returns:
//   - string: The encoded hash string in Argon2 format
//   - error: Any error that occurred during hashing (e.g., crypto/rand failure)
func GenerateHashString(input string) (string, error) {
	// argon2 params
	p := &params{
		memory:      64 * 1024,
		iterations:  3,
		parallelism: 2,
		saltLength:  16,
		keyLength:   32,
	}

	salt, err := generateRandomBytes(p.saltLength)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(input), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	// Base64 encode the salt and hashed input.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Return a string using the standard encoded hash representation.
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, p.memory, p.iterations, p.parallelism, b64Salt, b64Hash)

	return encodedHash, nil
}

// IsMatchingInputAndHash verifies if an input string matches a previously generated hash.
// This function safely compares the input string against a stored hash using
// constant-time comparison to prevent timing attacks.
//
// The function:
//  1. Decodes the stored hash to extract parameters, salt, and hash
//  2. Generates a new hash using the same parameters and salt
//  3. Compares the hashes using constant-time comparison
//
// This function is typically used for password verification during login.
//
// Example usage:
//
//	match, err := IsMatchingInputAndHash("myPassword123", storedHash)
//	if err != nil {
//	    // handle error (invalid hash format, incompatible version, etc.)
//	}
//	if match {
//	    // password is correct
//	} else {
//	    // password is incorrect
//	}
//
// Parameters:
//   - input: The string to verify (typically a password)
//   - encodedHash: The previously generated hash string to compare against
//
// Returns:
//   - bool: true if the input matches the hash, false otherwise
//   - error: Any error that occurred during verification (e.g., invalid hash format)
func IsMatchingInputAndHash(input, encodedHash string) (match bool, err error) {
	// Extract the parameters, salt and derived key from the encoded input
	// hash.
	p, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	// Derive the key from the other input using the same parameters.
	otherHash := argon2.IDKey([]byte(input), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	// Check that the contents of the hashed inputs are identical. Note
	// that we are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}
	return false, nil
}

// generateRandomBytes creates a cryptographically secure random byte slice.
// This function uses crypto/rand to generate random bytes suitable for use
// as cryptographic salt or other security-sensitive purposes.
//
// The function ensures that the generated bytes are cryptographically secure
// and unpredictable, which is essential for password hashing security.
//
// Parameters:
//   - n: The number of random bytes to generate
//
// Returns:
//   - []byte: A slice of n random bytes
//   - error: Any error that occurred during random generation
func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// decodeHash parses an Argon2 encoded hash string and extracts its components.
// This function decodes the standard Argon2 hash format to extract the algorithm
// parameters, salt, and hash value for verification purposes.
//
// The function expects a hash string in the format:
//
//	$argon2id$v=19$m=65536,t=3,p=2$[base64-salt]$[base64-hash]
//
// It validates the Argon2 version and extracts all parameters needed for
// hash verification.
//
// Parameters:
//   - encodedHash: The encoded hash string to decode
//
// Returns:
//   - *params: The Argon2 parameters (memory, iterations, parallelism, etc.)
//   - []byte: The decoded salt
//   - []byte: The decoded hash
//   - error: Any error that occurred during decoding (invalid format, incompatible version, etc.)
func decodeHash(encodedHash string) (p *params, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, errInvalidHash
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, errIncompatibleVersion
	}

	p = &params{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.saltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.keyLength = uint32(len(hash))

	return p, salt, hash, nil
}
