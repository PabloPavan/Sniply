package internal

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

var ErrNotFound = errors.New("not found")

func Env(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func MustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing env: %s", key)
	}
	return v
}

func RandomHex(nBytes int) string {
	b := make([]byte, nBytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func DefaultPasswordHasher(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
