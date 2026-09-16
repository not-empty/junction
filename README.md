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
  Mobile ──────▶│   HTTP API  (available now)  │──┐
  CLI    ──────▶│                              │  │
                |______________________________|  |
                                                  |
                                                  ├────▶     Your domain     
                ┌──────────────────────────────┐  │        (business rules)
  Kafka  ──────▶│                              │  │     
  SQS    ──────▶│  Queue/Event  (on the way)   │──┘
  Rabbit ──────▶│                              │
                └──────────────────────────────┘
```

Same services. Same repositories. Same validation and error semantics. Different bridge.

---

## Startup modes

Bridge is built around the idea of **startup modes** — the way a process exposes its capabilities to the outside world.

| Mode | Status | What it does |
|------|--------|--------------|
| **API** | Available | Serves your domain over HTTP with routing, middleware, JSON validation, and graceful shutdown. |
| **Event / Queue consumer** | Planned | Runs the same domain services against messages pulled from a broker, with the same validation and error mapping. |

The path is deliberate:

1. **Start simple.** Spin up an HTTP API, expose your features, ship something real fast.
2. **Grow.** Add domains. The structure keeps them isolated and predictable as the project gets bigger.
3. **Scale sideways.** When an operation outgrows request/response, run it as a queue consumer. The business logic you already wrote comes along untouched — you write the entry point, not the feature again.

The point is that step 3 costs you very little, because step 1 never let HTTP leak into your domain in the first place.

---

## Roadmap

- [x] API startup mode
- [ ] Hexagonal struct
