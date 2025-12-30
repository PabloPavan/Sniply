package apikeys

import (
	"errors"

	"github.com/jackc/pgx/v5"
)

var ErrNotFound = errors.New("api key not found")

func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows) || errors.Is(err, ErrNotFound)
}
