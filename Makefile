.DEFAULT_GOAL := check

GOOSE_VERSION := v3.28.0
GOVULNCHECK_VERSION := v1.8.0
GOOSE := go run -tags='no_clickhouse no_mssql no_mysql no_sqlite3 no_vertica no_ydb no_libsql' github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION)

.PHONY: fmt fmt-check vet test race check build run tidy migrate integration vuln up down

fmt:
	gofmt -w cmd internal

fmt-check:
	@test -z "$$(gofmt -l cmd internal)" || (gofmt -l cmd internal; exit 1)

vet:
	go vet ./...

test:
	go test ./...

race:
	go test -race ./...

check: fmt-check vet test
	go build ./...

build:
	mkdir -p bin
	go build -o bin/api ./cmd/api

run:
	go run ./cmd/api

tidy:
	go mod tidy

migrate:
	@test -n "$$DATABASE_URL" || (echo 'DATABASE_URL is required'; exit 1)
	@$(GOOSE) -dir migrations postgres "$$DATABASE_URL" up

integration:
	@test -n "$$TEST_DATABASE_URL" || (echo 'TEST_DATABASE_URL is required; use bookanook_test'; exit 1)
	@DATABASE_URL="$$TEST_DATABASE_URL" $(MAKE) migrate
	go test -race -count=1 ./internal/postgres

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./...

up:
	docker compose up --build -d api

down:
	docker compose down
