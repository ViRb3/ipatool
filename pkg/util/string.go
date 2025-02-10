package util

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func IfEmpty(value, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}

func MakeGuid(input string) string {
	hasher := sha256.New()
	hasher.Write([]byte(input))
	hash := hasher.Sum(nil)
	return strings.ToUpper(hex.EncodeToString(hash[:6]))
}
