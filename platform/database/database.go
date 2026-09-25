package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

const (
	defaultMaxOpenConns    = 25
	defaultMaxIdleConns    = 10
	defaultConnMaxLifetime = 5 * time.Minute
	defaultConnMaxIdleTime = 1 * time.Minute
	pingTimeout            = 5 * time.Second
)

type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type DB struct {
	pool *sql.DB
}

type txKey struct{}

func Open(ctx context.Context, cfg Config) (*DB, error) {
	mysqlCfg, err := mysql.ParseDSN(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("parsing database DSN: %w", err)
	}

	mysqlCfg.ParseTime = true
	mysqlCfg.Loc = time.UTC
	mysqlCfg.ClientFoundRows = true

	connector, err := mysql.NewConnector(mysqlCfg)
	if err != nil {
		return nil, fmt.Errorf("creating database connector: %w", err)
	}

	pool := sql.OpenDB(connector)
	pool.SetMaxOpenConns(valueOrDefault(cfg.MaxOpenConns, defaultMaxOpenConns))
	pool.SetMaxIdleConns(valueOrDefault(cfg.MaxIdleConns, defaultMaxIdleConns))
	pool.SetConnMaxLifetime(valueOrDefault(cfg.ConnMaxLifetime, defaultConnMaxLifetime))
	pool.SetConnMaxIdleTime(valueOrDefault(cfg.ConnMaxIdleTime, defaultConnMaxIdleTime))

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	err = pool.PingContext(pingCtx)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return &DB{pool: pool}, nil
}

// NewFromPool skips the DSN parsing and the ping Open does, so a test can
// inject a pool backed by a fake driver.
func NewFromPool(pool *sql.DB) *DB {
	return &DB{pool: pool}
}

func (d *DB) Ping(ctx context.Context) error {
	return d.pool.PingContext(ctx)
}

func (d *DB) Close() error {
	return d.pool.Close()
}

func (d *DB) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return fn(ctx)
	}

	tx, err := d.pool.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	err = fn(context.WithValue(ctx, txKey{}, tx))
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (d *DB) querier(ctx context.Context) querier {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}

	return d.pool
}

func valueOrDefault[T comparable](value T, fallback T) T {
	var zero T
	if value == zero {
		return fallback
	}

	return value
}
