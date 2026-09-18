package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

const mysqlErrDuplicateEntry = 1062

var (
	ErrNotFound  = errors.New("record not found")
	ErrDuplicate = errors.New("duplicate entry")
	ErrInvalidID = errors.New("invalid ULID")
)

func mapError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlErrDuplicateEntry {
		return fmt.Errorf("%w: %w", ErrDuplicate, err)
	}

	return err
}
