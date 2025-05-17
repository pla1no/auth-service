package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GenerateSecureToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("cannot generate secure token: %w", err)
	}
	return hex.EncodeToString(b), nil
}
