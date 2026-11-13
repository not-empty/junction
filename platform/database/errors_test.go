package database

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestMapError(t *testing.T) {
	other := errors.New("boom")

	cases := []struct {
		name string
		err  error
		want error
	}{
		{name: "no rows", err: sql.ErrNoRows, want: ErrNotFound},
		{name: "mysql duplicate", err: &mysql.MySQLError{Number: 1062}, want: ErrDuplicate},
		{name: "postgres duplicate", err: &pgconn.PgError{Code: "23505"}, want: ErrDuplicate},
		{name: "wrapped postgres duplicate", err: fmt.Errorf("insert: %w", &pgconn.PgError{Code: "23505"}), want: ErrDuplicate},
		{name: "other mysql error", err: &mysql.MySQLError{Number: 1146}, want: nil},
		{name: "other postgres error", err: &pgconn.PgError{Code: "42P01"}, want: nil},
		{name: "unknown", err: other, want: other},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mapError(tc.err)

			if tc.want == nil {
				if errors.Is(got, ErrDuplicate) || errors.Is(got, ErrNotFound) {
					t.Fatalf("mapError(%v) = %v, want it unmapped", tc.err, got)
				}
				return
			}

			if !errors.Is(got, tc.want) {
				t.Fatalf("mapError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
