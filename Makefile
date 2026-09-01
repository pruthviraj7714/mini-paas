.PHONY: api worker run-all build-all test fmt

api:
	go run ./cmd/api

worker:
	go run ./cmd/worker

run-all:
	go run ./cmd/api & \
	go run ./cmd/worker

build-all:
	mkdir -p bin
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker

test:
	go test ./...

fmt:
	go fmt ./...