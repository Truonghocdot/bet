package otp

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

func GenerateCode(length int) (string, error) {
	if length <= 0 {
		length = 6
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	digits := make([]byte, length)
	for index, value := range bytes {
		digits[index] = byte('0' + (value % 10))
	}

	return string(digits), nil
}

func Hash(secret, code string) string {
	sum := sha256.Sum256([]byte(secret + ":" + code))
	return hex.EncodeToString(sum[:])
}

func Compare(secret, hash, code string) bool {
	expected := Hash(secret, code)
	return subtle.ConstantTimeCompare([]byte(expected), []byte(hash)) == 1
}

func Last4(code string) string {
	if len(code) <= 4 {
		return code
	}

	return code[len(code)-4:]
}

func NewRequestToken() (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return fmt.Sprintf("rt_%s", hex.EncodeToString(bytes)), nil
}
