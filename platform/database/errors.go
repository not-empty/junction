package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	mysqlErrDuplicateEntry  = 1062
	postgresUniqueViolation = "23505"
)

var (
	ErrNotFound  = errors.New("record not found")
	ErrDuplicate = errors.New("duplicate entry")
	ErrInvalidID = errors.New("invalid ULID")
)

func mapError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}

	if isDuplicate(err) {
		return fmt.Errorf("%w: %w", ErrDuplicate, err)
	}

	return err
}

// isDuplicate checks both drivers: only the one in use ever matches.
func isDuplicate(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlErrDuplicateEntry {
		return true
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == postgresUniqueViolation {
		return true
	}

	return false
}
