package migration

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"time"

	driver "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/not-empty/junction/migrations"
	"github.com/not-empty/junction/platform/config"
)

var ErrNoChange = migrate.ErrNoChange

// The table name goes into golang-migrate's SQL as an identifier, so only
// plain identifiers, valid on both MySQL and Postgres, are accepted.
var validTable = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,63}$`)

func Up(cfg config.Config) error {
	return run(cfg, func(m *migrate.Migrate) error {
		return m.Up()
	})
}

func Down(cfg config.Config) error {
	return run(cfg, func(m *migrate.Migrate) error {
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

func Version(cfg config.Config) (uint, bool, error) {
	var version uint
	var dirty bool

	err := run(cfg, func(m *migrate.Migrate) error {
		var err error
		version, dirty, err = m.Version()

		if errors.Is(err, migrate.ErrNilVersion) {
			return nil
		}

		return err
	})

	return version, dirty, err
}

func run(cfg config.Config, fn func(*migrate.Migrate) error) error {
	m, err := open(cfg)
	if err != nil {
		return err
	}

	defer m.Close()

	err = fn(m)
	if err != nil {
		return explain(err, cfg)
	}

	return nil
}

func open(cfg config.Config) (*migrate.Migrate, error) {
	if !validTable.MatchString(cfg.MigrationsTable) {
		return nil, fmt.Errorf("invalid migrations table %q: must start with a letter or _ and contain only letters, digits or _", cfg.MigrationsTable)
	}

	pool, target, name, err := openTarget(cfg)
	if err != nil {
		return nil, err
	}

	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("reading the embedded migrations: %w", err)
	}

	// m.Close closes both the source and the pool.
	m, err := migrate.NewWithInstance("iofs", source, name, target)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("preparing migrate: %w", err)
	}

	return m, nil
}

func openTarget(cfg config.Config) (*sql.DB, database.Driver, string, error) {
	switch cfg.DatabaseDriver {
	case "", config.DriverMySQL:
		return openMySQL(cfg)
	case config.DriverPostgres:
		return openPostgres(cfg)
	default:
		return nil, nil, "", fmt.Errorf("unsupported database driver %q", cfg.DatabaseDriver)
	}
}

func openMySQL(cfg config.Config) (*sql.DB, database.Driver, string, error) {
	mysqlCfg, err := driver.ParseDSN(cfg.DatabaseDSN)
	if err != nil {
		return nil, nil, "", fmt.Errorf("parsing database DSN: %w", err)
	}

	mysqlCfg.MultiStatements = true
	mysqlCfg.ParseTime = true
	mysqlCfg.Loc = time.UTC

	connector, err := driver.NewConnector(mysqlCfg)
	if err != nil {
		return nil, nil, "", fmt.Errorf("creating database connector: %w", err)
	}

	pool := sql.OpenDB(connector)

	target, err := migratemysql.WithInstance(pool, &migratemysql.Config{
		MigrationsTable: cfg.MigrationsTable,
	})
	if err != nil {
		pool.Close()
		return nil, nil, "", fmt.Errorf("preparing the migrate driver: %w", err)
	}

	return pool, target, "mysql", nil
}

// Each file runs as a single Exec with no arguments, which pgx sends over the
// simple protocol: every statement in it runs in one implicit transaction.
func openPostgres(cfg config.Config) (*sql.DB, database.Driver, string, error) {
	pgCfg, err := pgx.ParseConfig(cfg.DatabaseDSN)
	if err != nil {
		return nil, nil, "", fmt.Errorf("parsing database DSN: %w", err)
	}

	pgCfg.RuntimeParams["timezone"] = "UTC"

	pool := stdlib.OpenDB(*pgCfg)

	target, err := migratepgx.WithInstance(pool, &migratepgx.Config{
		MigrationsTable: cfg.MigrationsTable,
	})
	if err != nil {
		pool.Close()
		return nil, nil, "", fmt.Errorf("preparing the migrate driver: %w", err)
	}

	return pool, target, "pgx5", nil
}

func explain(err error, cfg config.Config) error {
	var dirty migrate.ErrDirty

	if !errors.As(err, &dirty) {
		return err
	}

	if cfg.DatabaseDriver == config.DriverPostgres {
		return fmt.Errorf(
			"%w: migration %d failed and Postgres rolled its statements back, so it never ran. "+
				"Fix the file, then point the table back to the version before it with %s",
			err, dirty.Version, rewind(cfg.MigrationsTable, uint(dirty.Version)),
		)
	}

	return fmt.Errorf(
		"%w: migration %d failed halfway and MySQL cannot roll back DDL. "+
			"Fix the schema by hand, then clear the flag with "+
			"`UPDATE %s SET dirty = 0 WHERE version = %d`",
		err, dirty.Version, cfg.MigrationsTable, dirty.Version,
	)
}

// rewind is the statement that undoes the version a rolled back migration
// left behind: the previous version, or no version at all when it was the
// first one.
func rewind(table string, version uint) string {
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Sprintf("`UPDATE %s SET version = <previous version>, dirty = false`", table)
	}

	defer source.Close()

	previous, err := source.Prev(version)
	if err != nil {
		return fmt.Sprintf("`DELETE FROM %s`", table)
	}

	return fmt.Sprintf("`UPDATE %s SET version = %d, dirty = false`", table, previous)
}
