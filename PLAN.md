# SlotBook: Fixed-slot reservation API

Standalone project specification and implementation roadmap.

**Time budget:** 7 days × 6 focused hours = 42 hours. In the three-week plan, this occupies days 1–7.

**Prerequisites:** C++ or another programming language, basic backend knowledge, Git, Linux and basic Docker. The first two days introduce Go from scratch.

**Deliverable:** an independently runnable Go project with documented contracts, architecture decisions, automated tests and a reproducible demo. This document specifies work to implement; it does not contain an already implemented application.

## Technology stack

| Area | Technology | Purpose |
|---|---|---|
| Language and API | Go; net/http, encoding/json, context | HTTP server, JSON contracts and request cancellation |
| Persistence | PostgreSQL; pgx v5 and pgxpool | Parameterized SQL, constraints, transactions and connection pooling |
| Schema changes | goose SQL migrations | Repeatable schema creation and evolution |
| Testing | testing, httptest, race detector | Business, HTTP, concurrency and real-database integration tests |
| Operations | slog; Docker and Compose | Structured logs, reproducible runtime, health and graceful shutdown |
| Automation | GitHub Actions; go vet; govulncheck | Build, tests and static/security checks |
| Optional exercise | Gin | Port one handler to understand framework responsibilities |

Use a supported stable Go release and record it in the README and CI. Pin dependencies in `go.mod`/`go.sum` and container images to explicit versions. Pin migration tools and code generators where used. Consult documentation matching those versions.

## Purpose, boundaries and relevance

A small API for reserving **predefined, non-overlapping appointment slots** for a resource such as a lab workstation. It is deliberately not a general calendar: no recurring events, arbitrary time intervals, payments, notifications or UI. This keeps the domain small while making double booking and authorization real problems.

The project practices the Go, SQL, HTTP, testing and design skills expected in backend roles. Its portfolio value comes from tested invariants and explicit operation boundaries, rather than endpoint count.

## Required behavior and data

- `GET /v1/slots?resource_id=...&after=...&limit=...`: stable, bounded pagination, maximum 100 rows.
- `POST /v1/bookings`: `{slot_id}`, authenticated principal and `Idempotency-Key`; create or replay the original result.
- `GET /v1/bookings/{id}`: owner's booking only.
- `DELETE /v1/bookings/{id}`: cancel an owned booking; repeated cancellation succeeds without another side effect.
- `/livez` checks process health; `/readyz` checks whether requests can be served, including a bounded database probe.

Tables: `slots(id, resource_id, starts_at, ends_at)` and `bookings(id, slot_id, user_id, status, request_key, request_hash, created_at)`. Seed fixed slots; use `timestamptz` and UTC in examples. Slot end must exceed start. Store status as a constrained value. Keep cancelled rows for request replay and audit.

**Invariants:** a partial unique index on `bookings(slot_id)` for active rows prevents two active bookings; `(user_id, request_key)` is unique; foreign keys reject nonexistent slots. A reused key with a different request hash is a conflict. Replaying a previously cancelled booking's creation request returns that booking identity/current state and does not create a new reservation; document this contract. The application validates ownership from authenticated context, never a client-supplied `user_id`.

Use explicit SQL, parameterized values and a shared pgxpool. The storage operation handles the unique-conflict race and retrieves the winning idempotent result after transaction rollback when needed. Do not continue using a PostgreSQL transaction that has entered an error state. Conditional cancellation can be one atomic SQL statement; use a transaction only where multiple changes need a common commit.

## Architecture and idiomatic Go structure

```text
HTTP request → httpapi → booking.Service → booking.Store interface
                                           ↑ implemented by postgres.Store
cmd/api constructs and connects everything; PostgreSQL enforces final constraints.
```

```text
slotbook/
├── cmd/api/main.go
├── internal/
│   ├── booking/          # types, service, consumer-owned Store, domain errors, tests
│   ├── httpapi/          # routes, DTOs, auth middleware, error/status mapping
│   ├── postgres/         # pgx queries and atomic operations; integration tests
│   └── config/           # environment parsing and validation
├── migrations/           # goose SQL files
├── api/openapi.yaml
├── tests/integration/
├── docs/decisions.md
├── scripts/              # seed/demo helpers
├── .github/workflows/ci.yml
├── Dockerfile
├── compose.yaml
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

This is a suggested layout, not a mandatory Go standard. Co-locate unit tests with their code. `booking` imports neither HTTP nor pgx; PostgreSQL implements its small interface. Return concrete implementations from constructors. Do not create `IService`, `BaseRepository`, `utils`, or an interface for every struct. Composition/wiring belongs in `main`, and resource owners close their pools/servers.

**Stack:** Go standard HTTP/JSON/testing/slog/context; PostgreSQL; pgx v5/pgxpool; goose; Docker/Compose; GitHub Actions. Tests use real PostgreSQL for SQL behavior, simple fakes for isolated business behavior and `httptest` for HTTP.

## Acceptance checklist

- [ ] Twenty simultaneous requests for the same slot leave exactly one active booking.
- [ ] Concurrent retries with the same user/key/payload converge on one booking; changed payload conflicts.
- [ ] Cancelling makes the slot bookable again without removing request history.
- [ ] Wrong-owner access fails consistently; anonymous requests are rejected.
- [ ] SQL injection-like input stays data; malformed/oversized bodies fail predictably.
- [ ] Empty-database migrations, persistence across restart and graceful shutdown work.
- [ ] A query-plan note explains one real query and index tradeoff.
- [ ] README includes API examples, startup/test commands and one architecture decision.

**Security scope:** seeded development tokens are a local teaching adapter, not a finished identity platform. Do not commit live secrets. Bind local dependencies to loopback; document TLS and credential rotation as requirements before public deployment. OAuth/OIDC and arbitrary booking intervals are extensions.

## Implementation roadmap

Each day combines **75 minutes of theory, 210 minutes of implementation, 45 minutes of validation/buffer, and 30 minutes of recall**. Read only the assigned sections within that budget. Complete the day's vertical slice and exit check before adding optional features. End with a small commit and notes on what works, what can fail and what proves it.

### Project day 1 — Translate your C++ knowledge into Go

**Three-week schedule:** day 1.

**Learn (75 min):** [Tour](https://go.dev/tour/) basics through slices/maps; [module tutorial](https://go.dev/doc/tutorial/create-module) setup. Focus on zero values, value copying, slice backing arrays, strings/bytes/runes, `defer`, and returned errors. Contrast GC with RAII; close resources explicitly. Skim generics syntax, defer designing generic abstractions.

**Implement (210 min):** create SlotBook's module, `cmd/api`, and a small `booking` package. Define slot/booking values, validate inputs, write table-driven tests. Add examples demonstrating slice aliasing and map lookup's `ok` result. Build a CLI-sized smoke call before introducing HTTP.

**Gate/buffer (45 min):** `go test ./...` passes; explain why modifying a slice element can affect another slice. No unexplained copied tutorial code.

**Recall (30 min):** implement a frequency counter in Go; explain arrays versus slices and pointer receiver versus value receiver.

### Project day 2 — Interfaces, errors and a small application boundary

**Three-week schedule:** day 2.

**Learn:** [Tour](https://go.dev/tour/) methods/interfaces; [Effective Go](https://go.dev/doc/effective_go) interfaces/errors; [module layout](https://go.dev/doc/modules/layout). Pay attention to typed-nil interfaces, wrapping errors and composition.

**Implement:** define `booking.Service` and a consumer-owned storage interface with business operations. Build an in-memory test implementation; return distinguishable invalid-input, not-found and conflict errors. Keep HTTP/SQL types out of the business package. Write the first design record: “Why a small service boundary; why no generic repository.”

**Gate/buffer:** tests cover valid booking and unavailable-slot behavior; changing storage does not change business callers. Constructors make dependencies visible.

**Recall:** explain why an interface containing a nil pointer may not equal nil; compare Go interfaces with C++ virtual base classes.

### Project day 3 — HTTP contract before database plumbing

**Three-week schedule:** day 3.

**Learn:** [net/http](https://pkg.go.dev/net/http) handlers, ServeMux and Server; [RFC 9110](https://www.rfc-editor.org/rfc/rfc9110.html) methods/status codes; skim [OpenAPI](https://spec.openapis.org/oas/latest.html) paths/responses.

**Implement:** list slots, create/get/cancel a booking over JSON. Add request size limits, validation and consistent errors; write OpenAPI and `httptest` cases. Configure server timeouts; introduce request context propagation. The initial in-memory adapter is for local learning only.

**Gate/buffer:** executable examples demonstrate success, malformed JSON, missing record and conflict. HTTP handlers call the service, not storage directly.

**Recall:** describe POST versus PUT, 400/401/403/404/409, HTTP versus TCP and what a timeout does to an already-started operation.

### Project day 4 — PostgreSQL and pgx

**Three-week schedule:** day 4.

**Learn:** [PostgreSQL tutorial](https://www.postgresql.org/docs/current/tutorial.html) tables/joins/transactions; [constraints](https://www.postgresql.org/docs/current/ddl-constraints.html); [pgx](https://pkg.go.dev/github.com/jackc/pgx/v5), [pgxpool](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool) basic usage; [goose quickstart](https://pressly.github.io/goose/).

**Implement:** Compose PostgreSQL, migrations and deterministic seed slots. Implement the SQL storage adapter with parameters and a shared pool. Persist booking ownership/status; return domain errors from recognized constraint violations. Close rows and check iteration errors.

**Gate/buffer:** integration tests against PostgreSQL run from an empty database; restart the API and retrieve a booking. No string-built SQL values or leaked transactions.

**Recall:** write a join listing bookings with slot times; explain connection pooling and why a pool is not a single connection.

### Project day 5 — Correctness under concurrent requests

**Three-week schedule:** day 5.

**Learn:** [transaction isolation](https://www.postgresql.org/docs/current/transaction-iso.html), [EXPLAIN](https://www.postgresql.org/docs/current/using-explain.html) introductory examples, pgx transactions. Distinguish a Go data race from a database race.

**Implement:** enforce one active booking per slot with a database unique rule; make cancellation a conditional update. Add booking-request idempotency scoped to the authenticated user, including rejection of a reused key with a changed payload. Test simultaneous requests. Add an index for your listing query and inspect its plan on seeded data.

**Gate/buffer:** 20 competing bookings for one empty slot produce exactly one active booking. Identical request retries return the same booking; failed transactions leave no half-records.

**Recall:** explain why “SELECT availability, then INSERT” alone is unsafe; interpret your actual query plan without claiming every index guarantees an index scan.

### Project day 6 — Operable service, ownership and containers

**Three-week schedule:** day 6.

**Learn:** [Docker Go guide](https://docs.docker.com/guides/golang/), [Compose lifecycle](https://docs.docker.com/compose/how-tos/startup-order/), [slog](https://go.dev/blog/slog), [OWASP object-level authorization](https://owasp.org/API-Security/editions/2023/en/0x11-t10/).

**Implement:** development bearer-token middleware mapping two environment-provided tokens to seeded users; enforce booking ownership in the service. Add graceful shutdown, health/readiness, configuration validation and structured logs. Build a multi-stage, non-root image; preserve database data with a named volume.

**Gate/buffer:** another user cannot read/cancel a booking; absent token is rejected; stopping the process drains requests within a bounded grace period. Document local-only credential and TLS assumptions.

**Recall:** trace DNS → TCP → TLS → HTTP → handler → SQL; distinguish authentication from authorization and container image from running container.

### Project day 7 — Finish SlotBook and inspect a framework

**Three-week schedule:** day 7.

**Learn:** [GitHub Actions Go guide](https://docs.github.com/en/actions/tutorials/build-and-test-code/go), [Code Review Comments](https://go.dev/wiki/CodeReviewComments); spend at most 30 minutes of the reading block on the [Gin tutorial](https://go.dev/doc/tutorial/web-service-gin).

**Implement:** add CI for formatting checks, vet, tests/build and PostgreSQL integration tests. In a disposable exercise, port one handler to Gin using the same service. Finish SlotBook README, diagram, API examples and design record. Avoid maintaining two complete transports.

**Gate/buffer:** from a clean checkout, another person can start the service, migrate, seed and test it. All acceptance cases in this specification pass.

**Recall:** give a five-minute design walkthrough and solve one short map/slice problem. Explain what a router supplies and what remains application logic.

## Engineering workflow and final handoff

Implement and document `make up`, `make test`, `make integration`, `make demo`, and `make down`, or equivalent commands. These are proposed interfaces, not commands that this document supplies. Normal shutdown preserves persistent data; any destructive reset is separately named.

- Check formatting and run `go vet ./...`, `go test ./...`, `go build ./...`, and `go test -race ./...` on supported targets. Run real-dependency integration tests in a separate CI job. Use `govulncheck` and investigate applicable findings.
- Keep tests deterministic: explicit readiness, bounded polling, isolated fixtures and no public API dependencies. Use barriers/counters for concurrency assertions rather than arbitrary sleeps. Test important invariants and failure paths instead of chasing a coverage percentage.
- Wire concrete dependencies in `main`; use small consumer-owned interfaces where they help isolate behavior. Keep transport types and infrastructure clients out of core application contracts. Context and resource ownership must be explicit.
- Log failures at the boundary handling them. Keep secrets and unbounded user input out of logs and metric labels. Validate configuration on startup and bound shutdown.
- Deliver a README with startup, tests, API examples, architecture, guarantees, limitations and a short demo. Include an architecture decision record and actual test/experiment results. Do not claim production capacity or experience from synthetic local tests.

Database migrations and deterministic seed data are part of the startup instructions. Every transaction path must commit or roll back; use bounded cleanup even when the request context has expired. Integration tests must exercise real SQL constraints and rollback behavior.

## Scope control

Use the daily 45-minute buffer before extending the scope. First remove optional Gin exercise polish and extra filtering. Keep the booking uniqueness rule, request idempotency, ownership tests and persistence checks. If the core acceptance checks still fail, extend the schedule or document the incomplete state; do not weaken the correctness guarantees to meet a date.

## Learning resources

The daily roadmap identifies the sections to read. This table collects the references needed to use this document independently.

| Topic | Resource and focus |
|---|---|
| Go fundamentals | [A Tour of Go](https://go.dev/tour/) — basics, methods/interfaces, concurrency; [module tutorial](https://go.dev/doc/tutorial/create-module) |
| Go idiom | [Effective Go](https://go.dev/doc/effective_go) — names, control flow, interfaces, errors; [Code Review Comments](https://go.dev/wiki/CodeReviewComments) |
| Design/package boundaries | [Organizing a Go module](https://go.dev/doc/modules/layout); [Google Go style guide](https://google.github.io/styleguide/go/guide) — clarity, simplicity, consistency |
| Standard library | [net/http](https://pkg.go.dev/net/http), [context](https://pkg.go.dev/context), [testing](https://pkg.go.dev/testing), [slog introduction](https://go.dev/blog/slog) |
| HTTP/REST | [HTTP Semantics, RFC 9110](https://www.rfc-editor.org/rfc/rfc9110.html) — methods, idempotency, status codes; [OpenAPI](https://spec.openapis.org/oas/latest.html) — paths, schemas, responses |
| Containers | [Docker Go guide](https://docs.docker.com/guides/golang/), [Compose startup/shutdown order](https://docs.docker.com/compose/how-tos/startup-order/) |
| CI | [GitHub Actions: build and test Go](https://docs.github.com/en/actions/tutorials/build-and-test-code/go) |
| Security | [OWASP API Security Top 10](https://owasp.org/API-Security/editions/2023/en/0x11-t10/), [Go security practices](https://go.dev/doc/security/best-practices) |
| Framework literacy | [Official Go/Gin tutorial](https://go.dev/doc/tutorial/web-service-gin) |
| SQL/PostgreSQL | [PostgreSQL tutorial](https://www.postgresql.org/docs/current/tutorial.html) — SQL, joins, aggregates; [constraints](https://www.postgresql.org/docs/current/ddl-constraints.html); [isolation](https://www.postgresql.org/docs/current/transaction-iso.html); [EXPLAIN](https://www.postgresql.org/docs/current/using-explain.html) |
| pgx and migrations | [pgx v5](https://pkg.go.dev/github.com/jackc/pgx/v5) — queries/transactions; [pgxpool](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool); [goose](https://pressly.github.io/goose/) |
