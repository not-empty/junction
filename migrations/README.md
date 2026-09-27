# Migrations

Versioned schema, applied by `make migrate`. Files are embedded into the
`migrate` binary, so the deployed artifact carries its own schema.

`make api|worker|event name=<domain>` writes the migration of a new domain here,
once per domain. Write one by hand with `make migration name=<what_it_does>`.

## Append-only

A migration that already ran somewhere is history: never edit it and never
delete it. To drop a table, add a new migration that drops it — removing the
file that created it leaves `schema_migrations` pointing at a version that no
longer exists on disk.

## Dirty state

MySQL has no transactional DDL. A migration that fails halfway leaves the
version marked `dirty`, and every later run refuses to start until someone
fixes the schema by hand and clears the flag:

    UPDATE schema_migrations SET dirty = 0 WHERE version = <v>;
