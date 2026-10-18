# Book a nook

__book-a-nook__ is a Go service for booking time slots for shared [ITMO](https://itmo.ru) resources.

## Usage

| Environment variable | Description |
| -------------------- | ----------- |
| `DATABASE_URL` | PostgreSQL connection string (required). |
| `HTTP_ADDR` | HTTP listen address (defaults to `127.0.0.1:8080`). |
| `DEV_TOKEN_TIM` | Development bearer token for user `tim` (required). |
| `DEV_TOKEN_HELEN` | Development bearer token for user `helen` (required). |


| Method | Endpoint | Description |
| ------ | -------- | ----------- |
| `GET` | `/v1/slots?resource_id=...&after=...&limit=...` | List slots for a resource. |
| `POST` | `/v1/bookings` | Book a slot. |
| `GET` | `/v1/bookings/{id}` | Retrieve an owned booking. |
| `DELETE` | `/v1/bookings/{id}` | Cancel an owned booking. |
| `GET` | `/livez` | Check process health. |
| `GET` | `/readyz` | Check database connectivity. |

Health endpoints do not require authentication. Readiness returns `503` if the database probe fails.

### Docker installation

```sh
cp .env.example .env
make up
```

Compose starts PostgreSQL, applies goose migrations, and starts the API at `http://127.0.0.1:8080`. PostgreSQL is available at `127.0.0.1:6666`.

Stop the services:

```sh
make down
```

### Local installation

Requirements:

- Go 1.27.2 or later.
- Docker with Compose for PostgreSQL.
- Make.

```sh
cp .env.example .env
set -a
. ./.env
set +a
docker compose up -d db
make migrate
make run
```


### Seed data

```sh
docker compose exec -T db psql -U bookanook -d bookanook -v ON_ERROR_STOP=1 <<'SQL'
INSERT INTO users (id) VALUES ('tim'), ('helen') ON CONFLICT DO NOTHING;
INSERT INTO resources (id) VALUES ('nook-1') ON CONFLICT DO NOTHING;
INSERT INTO slots (resource_id, starts_at, ends_at)
VALUES ('nook-1', '2026-11-07T12:00:00Z', '2026-11-07T13:00:00Z')
ON CONFLICT DO NOTHING
RETURNING id;
SQL
```

### Specifications

- Slots are listed as a JSON array ordered by ID, including booked slots. `after` is an exclusive slot ID cursor; use the last returned ID for the next page. The default limit is 20 (also used for `limit=0`); the maximum is 100.
- Slot end must exceed start. Slots for the same resource cannot overlap; adjacent slots are allowed. Times are stored as PostgreSQL `timestamptz`.
- A partial unique index allows only one active booking per slot, including concurrent requests.
- Booking status is `active` or `cancelled`. Cancellation frees the slot and preserves the booking record.
- `Idempotency-Key` is scoped to a user, must contain non-whitespace characters, and may contain at most 128 bytes. The prefix `__legacy_booking__:` is reserved.
- Reusing a key with the same slot returns the existing booking's current state. Reusing it with another slot returns `409`. Replaying a cancelled booking returns that cancelled booking; booking the slot again requires a new key.
- Another user's booking returns `404` for both retrieval and cancellation. Missing or invalid authentication returns `401`.
- Creation requires `Content-Type: application/json`. Unknown fields, malformed JSON, and multiple JSON values return `400`; bodies exceeding 16 KiB return `413`; unsupported content types return `415`.
- API validation errors return `400`, missing records return `404`, and booking conflicts return `409`. API errors use the form `{"error":"message"}`.
- The server writes structured JSON logs and allows up to five seconds for graceful shutdown.

Development tokens and the supplied database credentials are intended for local use. Public deployment requires production authentication, TLS, and credential rotation.

## Building and testing

Build the project:

```sh
make build
```

Check formatting, run static analysis and tests:

```sh
make check
make race
make vuln
```

Test migrations and integration:

```sh
docker compose up -d --wait db
docker compose exec db createdb -U bookanook bookanook_test
TEST_DATABASE_URL='postgres://bookanook:bookanook@127.0.0.1:6666/bookanook_test?sslmode=disable' \
  make integration
```

## Structure

- `cmd/api`: executable.
- `internal/booking`: domain types, validation, ownership checks, service, and storage interface.
- `internal/httpapi`: routing and authentication.
- `internal/postgres`: pgx storage implementation.
- `internal/memory`: in-memory storage.
- `internal/config`: environment parsing.
- `migrations`: goose SQL migrations.

## License

The source code is licensed under the [MIT License](LICENSE).
