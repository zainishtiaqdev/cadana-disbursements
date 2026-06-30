.PHONY: run test build tidy

run:
	go run ./cmd/server

test:
	go test -race ./...

build:
	go build -o bin/server ./cmd/server

tidy:
	go mod tidy
