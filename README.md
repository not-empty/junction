# Bridge

**A Go backend starter that bridges any consumer to your business logic — and grows with you.**

Bridge is the gap-closer between whatever is calling you (a web app, a mobile app, a CLI, another service, a queue) and the business rules you actually care about writing. You get a project off the ground in minutes as a plain HTTP API, and when traffic, scale, or architecture demands something different, you switch the entry point — not your code.

---

## Why "Bridge"

Every project starts with the same question: *how do consumers reach my logic?*

The answer usually changes over time. You ship an HTTP API because it's the fastest way to put something in front of a frontend. Six months later, the same operation needs to run asynchronously off a queue, triggered by an event instead of a request. In most codebases that means a rewrite: the business rules were written *inside* the HTTP handlers, so they leave with them.

Bridge treats the delivery mechanism as a detail. HTTP is one bridge to your domain. A queue consumer is another. The domain itself doesn't know or care which one is in use:

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

Same services. Same repositories. Same validation and error semantics. Different bridge.

---

## Startup modes

Bridge is built around the idea of **startup modes** — the way a process exposes its capabilities to the outside world.

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
3. **Scale sideways.** When an operation outgrows request/response, give that domain a second bridge with `make worker name=…` or `make event name=…`. The business logic you already wrote comes along untouched — you write the entry point, not the feature again.

The point is that step 3 costs you very little, because step 1 never let HTTP leak into your domain in the first place.

---

## Scaffolding

A domain is a folder under `internal/modules/`, and each startup mode reaches it through its own thin layer. The generator creates that structure for you:

```
make api name=product        # HTTP bridge,  mounts it in cmd/api/modules.go
make worker name=product     # queue bridge, mounts it in cmd/worker/modules.go
make event name=product      # event bridge, mounts it in cmd/event/modules.go
make status                  # domains and the bridges each one has
```

Every command creates the shared core (`domain`, `repository`, `service`) when it is missing and reuses it when it is already there. There is no ordering and no prerequisite: a domain can start as a worker and gain HTTP later, or never have HTTP at all. Running the same command twice changes nothing.

The domain name accepts `product`, `order-item`, `order_item` or `OrderItem` — all four produce the same result. The table takes the name exactly as you wrote it, so `name=product` queries `product`. Pass `table=` when the table is named differently:

```
make api name=product table=tbl_products
```

The generator does not touch the database. The repository it writes expects the table to carry a `deleted_at` column for the soft delete, on top of the columns listed in the generated `domain` package.

---

## Roadmap

- [x] API startup mode
- [x] Queue worker startup mode
- [x] Event consumer startup mode
- [x] Domain scaffolding
- [ ] Hexagonal struct
