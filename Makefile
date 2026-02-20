.PHONY: build install test test-unit test-integration lint fmt vet tidy vuln coverage check clean all docker-build docker-run

BINARY_NAME=release-it-go
BUILD_DIR=bin
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_DATE)"

build:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/release-it-go

install:
	go install $(LDFLAGS) ./cmd/release-it-go

test:
	go test ./... -v -cover -race

test-unit:
	go test ./internal/... -v -cover -race

test-integration:
	go test ./test/integration/... -v -race

lint:
	golangci-lint run

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy
	go mod verify

vuln:
	govulncheck ./...

coverage:
	go test ./... -coverprofile=coverage.out -race
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

check: fmt vet lint vuln test build
	@echo "All checks passed."

clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html

all: check

docker-build:
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(BINARY_NAME):latest .

docker-run:
	docker run --rm -v $(PWD):/workspace -e GITHUB_TOKEN $(BINARY_NAME):latest --ci
