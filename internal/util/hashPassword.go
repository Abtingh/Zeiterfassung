package util

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword takes a plain-text password as input and returns its bcrypt hash as a string.
// It uses a cost factor of 10 for hashing. If hashing fails, an error is returned.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
