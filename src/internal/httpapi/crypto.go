package httpapi

import (
	"golang.org/x/crypto/bcrypt"
)

func defaultPasswordHasher(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
