package apikeys

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/PabloPavan/sniply_api/internal"
)

func GenerateToken() string {
	return "sk_" + internal.RandomHex(32)
}

func TokenPrefix(token string) string {
	if token == "" {
		return ""
	}
	if len(token) <= 8 {
		return token
	}
	return token[:8]
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}
