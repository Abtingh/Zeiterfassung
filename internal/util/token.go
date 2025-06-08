package util

import (
	"crypto/rand"
	"encoding/base64"
	"log"
)

// GenerateToken generates a random token of the specified length in bytes,
// encodes it using base64 URL encoding, and returns the resulting string.
// If random byte generation fails, the function logs a fatal error and exits.
//
// Parameters:
//   length - the number of random bytes to generate before encoding.
//
// Returns:
//   A base64 URL-encoded string representation of the random bytes.
func GenerateToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}
	return base64.URLEncoding.EncodeToString(bytes)
}
