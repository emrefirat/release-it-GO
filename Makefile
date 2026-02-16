.PHONY: build test lint fmt clean all docker-build docker-run

BINARY_NAME=release-it-go
BUILD_DIR=bin
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

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

docker-build:
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(BINARY_NAME):latest .

docker-run:
	docker run --rm -v $(PWD):/workspace -e GITHUB_TOKEN $(BINARY_NAME):latest --ci
