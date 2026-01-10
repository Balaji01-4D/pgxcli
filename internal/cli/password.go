package cli

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// shouldAskForPassword decides whether to prompt for password based on error code and flags.
func shouldAskForPassword(err error, neverPrompt bool) bool {
	if neverPrompt {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "28P01" { // invalid_password
		return true
	}
	return false
}
