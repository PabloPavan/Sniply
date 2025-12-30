package users

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrNotFound = errors.New("user not found")

func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows) || errors.Is(err, ErrNotFound)
}

func IsUniqueViolationEmail(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	// 23505 = unique_violation
	if pgErr.Code != "23505" {
		return false
	}

	if pgErr.ConstraintName == "users_email_key" {
		return true
	}

	if pgErr.ColumnName == "email" {
		return true
	}

	return false
}
