package migration

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	driver "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/not-empty/junction/migrations"
)

var ErrNoChange = migrate.ErrNoChange

func Up(dsn string) error {
	return run(dsn, func(m *migrate.Migrate) error {
		return m.Up()
	})
}

func Down(dsn string) error {
	return run(dsn, func(m *migrate.Migrate) error {
		_, _, err := m.Version()
		if errors.Is(err, migrate.ErrNilVersion) {
			return migrate.ErrNoChange
		}

		if err != nil {
			return err
		}

		return m.Steps(-1)
	})
}

func Version(dsn string) (uint, bool, error) {
	var version uint
	var dirty bool

	err := run(dsn, func(m *migrate.Migrate) error {
		var err error
		version, dirty, err = m.Version()

		if errors.Is(err, migrate.ErrNilVersion) {
			return nil
		}

		return err
	})

	return version, dirty, err
}

func run(dsn string, fn func(*migrate.Migrate) error) error {
	m, err := open(dsn)
	if err != nil {
		return err
	}

	defer m.Close()

	err = fn(m)
	if err != nil {
		return explain(err)
	}

	return nil
}

func open(dsn string) (*migrate.Migrate, error) {
	cfg, err := driver.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("parsing database DSN: %w", err)
	}

	cfg.MultiStatements = true
	cfg.ParseTime = true
	cfg.Loc = time.UTC

	connector, err := driver.NewConnector(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating database connector: %w", err)
	}

	pool := sql.OpenDB(connector)

	target, err := migratemysql.WithInstance(pool, &migratemysql.Config{})
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("preparing the migrate driver: %w", err)
	}

	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("reading the embedded migrations: %w", err)
	}

	// m.Close closes both the source and the pool.
	m, err := migrate.NewWithInstance("iofs", source, "mysql", target)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("preparing migrate: %w", err)
	}

	return m, nil
}

func explain(err error) error {
	var dirty migrate.ErrDirty

	if errors.As(err, &dirty) {
		return fmt.Errorf(
			"%w: migration %d failed halfway and MySQL cannot roll back DDL. "+
				"Fix the schema by hand, then clear the flag with "+
				"`UPDATE schema_migrations SET dirty = 0 WHERE version = %d`",
			err, dirty.Version, dirty.Version,
		)
	}

	return err
}
