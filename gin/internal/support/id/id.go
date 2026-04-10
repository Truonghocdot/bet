package id

import (
	"crypto/rand"
	"encoding/hex"
)

func New() string {
	buffer := make([]byte, 16)
	_, err := rand.Read(buffer)
	if err != nil {
		return "fallback-connection-id"
	}

	return hex.EncodeToString(buffer)
}
