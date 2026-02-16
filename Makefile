.PHONY: build test lint fmt clean all

BINARY_NAME=release-it-go
BUILD_DIR=bin

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/release-it-go

test:
	go test ./... -v -cover -race

lint:
	golangci-lint run

fmt:
	go fmt ./...
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)

all: fmt lint test build
