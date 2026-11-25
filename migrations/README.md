# Migrations

Versioned schema, applied by `make migrate`. Files are embedded into the
`migrate` binary, so the deployed artifact carries its own schema.

`make api|worker|event name=<domain>` writes the migration of a new domain here,
once per domain. Write one by hand with `make migration name=<what_it_does>`.

## Append-only

A migration that already ran somewhere is history: never edit it and never
delete it. To drop a table, add a new migration that drops it — removing the
file that created it leaves the migrations table pointing at a version that no
longer exists on disk.

## Dialect

The files are plain SQL for the database in `DATABASE_DRIVER`, and the
generator writes them in that dialect. Pick the database before the first
migration: switching later means rewriting every file here.

## Dirty state

A migration that fails leaves the version marked `dirty`, and every later run
refuses to start until someone resolves it. `make migrate` prints the exact
statement for the case at hand.

**MySQL** has no transactional DDL: the statements before the failure stay
applied. Fix the schema by hand, finishing what the migration meant to do, then
clear the flag:

    UPDATE schema_migrations SET dirty = 0 WHERE version = <v>;

**Postgres** runs each file in a single transaction, so a failed migration
leaves nothing behind. Fix the file, then point the table back to the version
before it, so the next run applies it again:

    UPDATE schema_migrations SET version = <previous version>, dirty = false;

When it was the first migration, `DELETE FROM schema_migrations;` instead.
A file holding a statement that cannot run in a transaction, such as
`CREATE INDEX CONCURRENTLY`, loses that guarantee and is resolved as on MySQL.

The table is `schema_migrations` unless `MIGRATIONS_TABLE` says otherwise.

## Sharing a database

Several services built on junction can use the same database. Each one tracks
its own version in a separate table: set a distinct `MIGRATIONS_TABLE` per
service (e.g. `orders_schema_migrations`), and keep the tables each service
creates disjoint from the others'.
