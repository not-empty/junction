# Junction

**A Go backend starter where every path leads to the same business logic — and grows with you.**

Junction is the meeting point between whatever is calling you (a web app, a mobile app, a CLI, another service, a queue, an event stream) and the business rules you actually care about writing. You get a project off the ground in minutes as a plain HTTP API, and when traffic, scale, or architecture demands something different, you add a new way in — not a new version of your code.

---

## Why "Junction"

A junction is the point where several paths arrive at the same place. On a railway, it is also where a train switches lines without unloading its cargo, and where a branch line can split off and run on its own.

That is the whole idea of the project:

- **Many paths, one point.** HTTP, a queue and an event stream are different lines into the same domain. The domain doesn't know or care which one carried the request.
- **Switch lines, keep the cargo.** You ship an HTTP API because it's the fastest way to put something in front of a frontend. Six months later, the same operation needs to run off a queue, triggered by an event instead of a request. In most codebases that means a rewrite, because the business rules were written *inside* the HTTP handlers. Here the rules stay where they are; you add an entry.
- **Branch off when it grows.** Each domain is a self-contained folder that every entry reaches through the same `Module` function. When one domain outgrows the project, it moves into another Junction project and runs there as its own service — only its import paths change, since both share the same `platform/`.

```
                ┌──────────────────────────────┐
  Web    ──────▶│                              │
  Mobile ──────▶│   HTTP API        cmd/api    │──┐
  CLI    ──────▶│                              │  │
                └──────────────────────────────┘  │
                                                  │
                ┌──────────────────────────────┐  │
  Redis  ──────▶│   Queue worker    cmd/worker │──┼────▶   Your domain
                └──────────────────────────────┘  │      (business rules)
                                                  │
                ┌──────────────────────────────┐  │
  Kafka  ──────▶│   Event consumer  cmd/event  │──┘
                └──────────────────────────────┘
```

Same services. Same repositories. Same validation and error semantics. Different path in.

---

## Getting started

Requirements: Go 1.26.7+, Docker with Compose.

```
cp .env.example .env
docker compose up -d        # MySQL, Redis and Kafka for local development
make migrate                # apply the schema in migrations/
go run ./cmd/api            # http://localhost:8080/health
```

`docker-compose.yml` only runs the infrastructure; the application runs on your machine with `go run`. Every port is bound to `127.0.0.1`, and MySQL and Redis load their tuned settings from `docker/`.

---

## Startup modes

Junction is built around the idea of **startup modes** — the way a process exposes its capabilities to the outside world.

| Mode | Entrypoint | What it does |
|------|-----------|--------------|
| **API** | `cmd/api` | Serves your domain over HTTP with routing, middleware, JSON validation, and graceful shutdown. |
| **Queue worker** | `cmd/worker` | Consumes an omniq queue on Redis. One process per queue, picked with `--queue`, with a per-job timeout and draining on shutdown. |
| **Event consumer** | `cmd/event` | Subscribes as a Kafka consumer group to every topic in the registry, with retries, backoff, and a `.dlq` topic for messages that exhaust them. |

All three are available, and all three run the same services and repositories:

```
go run ./cmd/api
go run ./cmd/worker --queue=product.create
go run ./cmd/event
```

The worker takes one queue per process, so scaling a hot queue means running more of that one. The event consumer handles every registered topic in a single process.

The path is deliberate:

1. **Start simple.** Spin up an HTTP API, expose your features, ship something real fast.
2. **Grow.** Add domains with `make api name=…`. The structure keeps them isolated and predictable as the project gets bigger.
3. **Scale sideways.** When an operation outgrows request/response, give that domain a second entry with `make worker name=…` or `make event name=…`. The business logic you already wrote comes along untouched — you write the entry point, not the feature again.

The point is that step 3 costs you very little, because step 1 never let HTTP leak into your domain in the first place.

### How each mode treats errors

Business errors are `apperror.Error` values, and every entry reads them the same way: the input was wrong, so trying again will not help.

| Mode | `apperror.Error` | Any other error |
|------|------------------|-----------------|
| API | Mapped to its HTTP status, with `message` and `fields` | `500` with a generic body; the error is logged |
| Worker | Job discarded and logged | Job reported as failed to omniq, which applies the queue's retry policy |
| Event | Message discarded and offset committed | Retried 5 times with a linear 1s backoff, then written to `<topic>.dlq` |

| Kind | HTTP status |
|------|-------------|
| `BadRequest` | 400 |
| `Unauthorized` | 401 |
| `Forbidden` | 403 |
| `NotFound` | 404 |
| `Conflict` | 409 |
| `PayloadTooLarge` | 413 |
| `Validation` | 422 |

---

## Scaffolding

A domain is a folder under `internal/modules/`, and each startup mode reaches it through its own thin layer. The generator creates that structure for you:

```
make api name=product        # HTTP entry,  mounts it in cmd/api/modules.go
make worker name=product     # queue entry, mounts it in cmd/worker/modules.go
make event name=product      # event entry, mounts it in cmd/event/modules.go
make status                  # domains and the entries each one has
```

Every command creates the shared core (`domain`, `repository`, `service`) when it is missing and reuses it when it is already there. There is no ordering and no prerequisite: a domain can start as a worker and gain HTTP later, or never have HTTP at all. Running the same command twice changes nothing.

The domain name accepts `product`, `order-item`, `order_item` or `OrderItem` — all four produce the same result. The table takes the domain name in snake_case, so `name=order-item` queries `order_item`. Pass `table=` when the table is named differently:

```
make api name=product table=tbl_products
```

### What gets generated

For `make api name=product`:

```
internal/modules/product/
├── domain/product.go                       entity and input structs, with validate tags
├── repository/product_repository.go        SQL with soft delete and cursor pagination
├── service/product_service.go              business rules, maps database errors to apperror
├── controller/product_controller.go        HTTP handlers   (api)
├── controller/product_routes.go            Module + routes (api)
├── consumer/product_consumer.go            job handlers    (worker)
├── consumer/product_queues.go              Module + queues (worker)
├── listener/product_listener.go            event handlers  (event)
└── listener/product_events.go              Module + topics (event)
migrations/<version>_create_product_table.{up,down}.sql
```

Each layer comes with its tests. The first command for a domain also writes its `create_<name>_table` migration; later commands find it and leave it alone.

The generated entries expose:

| Entry | Exposes |
|--------|-------|
| API | `POST /product/add`, `GET /product/list?cursor=`, `GET /product/{id}`, `PATCH /product/{id}`, `DELETE /product/{id}` |
| Worker | queue `product.create` |
| Event | topic `product.registered` |

The entity starts with a single `name` field: it is a starting point to be edited, not a schema to keep.

### Conventions in the generated code

- **IDs are ULIDs** stored as `CHAR(26)`. The database layer creates one when the input has no `id` and rejects an invalid one.
- **Deletes are soft**: they set `deleted_at`, and every read filters it out.
- **Lists use cursor pagination** over the ID, 25 per page. The response carries `page_cursor` while there is a next page.
- **HTTP responses** wrap the payload in `{"data": …}`; errors are `{"message": …, "fields": {…}}`, with `fields` holding the failed validation rule of each field.
- **Request bodies** reject unknown fields, more than one JSON object, and anything above `MAX_BODY_BYTES`.

---

## Migrations

The schema lives in `migrations/` as versioned `up`/`down` SQL pairs, applied with [golang-migrate](https://github.com/golang-migrate/migrate). The files are embedded into the `migrate` binary, so a deployed image carries its own schema.

```
make migrate                       # apply every pending migration
make migrate-down                  # roll back the last one
make migrate-status                # current version and dirty flag
make migration name=add_price      # empty pair for a change written by hand
make table name=product_tag        # create table pair, without a domain (e.g. a pivot table)
```

Migrations are append-only, and a failed one leaves the schema marked dirty, since MySQL cannot roll back DDL. `migrations/README.md` explains both.

---

## Configuration

Every process reads its settings from environment variables. A `.env` file in the working directory is loaded when present, and variables already set in the environment win over it.

| Variable | Used by | Default | Description |
|----------|---------|---------|-------------|
| `DATABASE_DSN` | all | required | MySQL DSN, e.g. `user:pass@tcp(host:3306)/db` |
| `PORT` | api | `8080` | HTTP port |
| `CORS_ALLOWED_ORIGINS` | api | required | `*`, or a comma-separated list. `https://*.example.com` matches any subdomain |
| `MAX_BODY_BYTES` | api | `1048576` | Largest accepted request body |
| `REDIS_HOST` | worker | required | omniq Redis host |
| `REDIS_PORT` | worker | required | omniq Redis port |
| `KAFKA_BROKERS` | event | required | Comma-separated broker list |
| `KAFKA_GROUP_ID` | event | `junction` | Consumer group |

`MYSQL_*`, `MYSQL_PORT` and `KAFKA_PORT` in `.env.example` are only read by `docker-compose.yml`.

Every process connects to MySQL on start, since all entries share the same repositories. A new shared dependency is declared once, in `platform/bootstrap/deps.go`, and becomes available to every mode without touching `cmd/`.

---

## Docker

The `Dockerfile` builds a `scratch` image with the `api` and `migrate` binaries, statically compiled, running as an unprivileged user.

```
docker build -t junction .
docker run --rm --env-file .env -p 8080:8080 junction
docker run --rm --env-file .env --entrypoint /migrate junction up
```

`.env` is kept out of the image by `.dockerignore`, so configuration always comes from the environment. The build cross-compiles for the target platform instead of emulating it, so a multi-arch image builds at native speed:

```
docker buildx build --platform linux/amd64,linux/arm64 -t junction .
```

---

## Project structure

```
cmd/
├── api/            HTTP entrypoint; modules.go lists the mounted domains
├── worker/         omniq queue entrypoint
├── event/          Kafka entrypoint
└── migrate/        schema migrations
internal/
└── modules/        one folder per domain; health ships as an example
migrations/         versioned SQL, embedded into cmd/migrate
platform/           shared building blocks, free of business rules
├── apperror/       error kinds every entry understands
├── bootstrap/      dependencies and module mounting for each mode
├── config/         .env and environment loading
├── database/       MySQL pool, transactions, query helpers, ULIDs
├── event/          event registry, dispatcher and Kafka subscriber
├── httpserver/     server, graceful shutdown, recover, CORS and body limit middleware
├── httpx/          JSON request decoding and responses
├── migration/      golang-migrate over the embedded files
├── payload/        JSON decoding for queue and event payloads
├── queue/          queue registry and omniq adapter
└── validation/     struct validation into field errors
tools/junction/     the generator behind the make targets
docker/             MySQL and Redis settings for docker compose
```

`database.DB.WithTx` runs a function inside a transaction carried by the context, so repositories called within it join the same transaction without changing their signatures.

---

## Testing

```
make test     # go test ./...
make cover    # total coverage
```

Generated domains include tests for the repository, service and entry layers. They run without a database: repositories are tested against `go-sqlmock` and the other layers against mocks.
