package database

import (
	"context"
	"database/sql"
)

type Scanner interface {
	Scan(dest ...any) error
}

type querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func QueryOne[T any](ctx context.Context, db *DB, scan func(Scanner) (T, error), query string, args ...any) (T, error) {
	item, err := scan(db.querier(ctx).QueryRowContext(ctx, query, args...))
	if err != nil {
		var zero T
		return zero, mapError(err)
	}

	return item, nil
}

func QueryMany[T any](ctx context.Context, db *DB, scan func(Scanner) (T, error), query string, args ...any) ([]T, error) {
	rows, err := db.querier(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	items := []T{}
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, mapError(err)
		}

		items = append(items, item)
	}

	err = rows.Err()
	if err != nil {
		return nil, mapError(err)
	}

	return items, nil
}

func Insert(ctx context.Context, db *DB, query string, id string, args ...any) (string, error) {
	id, err := resolveID(id)
	if err != nil {
		return "", err
	}

	_, err = Exec(ctx, db, query, append([]any{id}, args...)...)
	if err != nil {
		return "", err
	}

	return id, nil
}

func Exec(ctx context.Context, db *DB, query string, args ...any) (int64, error) {
	result, err := db.querier(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return 0, mapError(err)
	}

	return result.RowsAffected()
}
