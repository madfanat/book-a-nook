.DEFAULT_GOAL := check

.PHONY: fmt vet test race check build run tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./...

race:
	go test -race ./...

check: vet test
	go build ./...

build:
	mkdir -p bin
	go build -o bin/api ./cmd/api

run:
	go run ./cmd/api

tidy:
	go mod tidy