package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	Secret    []byte
	Issuer    string
	Audience  string
	AccessTTL time.Duration
}

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func (s *Service) IssueAccessToken(userID string, role string) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(s.AccessTTL)

	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    s.Issuer,
			Audience:  []string{s.Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	str, err := tok.SignedString(s.Secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return str, exp, nil
}

func (s *Service) ValidateAccessToken(tokenString string) (*Claims, error) {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuedAt(),
		jwt.WithAudience(s.Audience),
		jwt.WithIssuer(s.Issuer),
	)

	var claims Claims
	_, err := parser.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		return s.Secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims.Subject == "" {
		return nil, errors.New("missing sub")
	}
	return &claims, nil
}
