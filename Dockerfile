# ============================================================================
# Stage 1: Builder
# ============================================================================
FROM golang:1.24.3-alpine AS builder

ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE=unknown

WORKDIR /build

# Cache dependency downloads in a separate layer
COPY go.mod go.sum ./
RUN go mod download

# Copy only source code needed for build
COPY cmd/ cmd/
COPY internal/ internal/

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE}" \
    -o /build/release-it-go \
    ./cmd/release-it-go

# ============================================================================
# Stage 2: Runtime
# ============================================================================
FROM alpine:3.21

ARG USER_UID=1000
ARG USER_GID=1000

# OCI metadata labels
LABEL org.opencontainers.image.title="release-it-go" \
      org.opencontainers.image.description="Release automation tool for Git projects" \
      org.opencontainers.image.source="https://github.com/user/release-it-go" \
      org.opencontainers.image.licenses="MIT"

# Install runtime dependencies
RUN apk add --no-cache git openssh-client ca-certificates

# Create non-root user
RUN addgroup -g ${USER_GID} releaser && \
    adduser -D -u ${USER_UID} -G releaser releaser

# Copy binary from builder
COPY --from=builder /build/release-it-go /usr/local/bin/release-it-go

# Setup workspace
RUN mkdir -p /workspace && chown releaser:releaser /workspace
WORKDIR /workspace

USER releaser

# Allow mounted repositories to be used safely (must be after USER releaser)
RUN git config --global --add safe.directory '*'

ENTRYPOINT ["/usr/local/bin/release-it-go"]
CMD ["--help"]
