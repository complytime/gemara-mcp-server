FROM golang:1.25.4-alpine AS builder
ARG VERSION="dev"

# Set the working directory
WORKDIR /build

# Install git
RUN --mount=type=cache,target=/var/cache/apk \
    apk add git

# Build the server
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=bind,target=. \
    CGO_ENABLED=0 go build -ldflags="-s -w -X github.com/complytime/gemara-mcp-server/version.Version=${VERSION} -X github.com/complytime/gemara-mcp-server/version.Build=$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')" \
    -o /bin/gemara-mcp-server cmd/gemara-mcp-server/main.go

FROM registry.access.redhat.com/ubi9/ubi-minimal:latest

RUN microdnf install -y ca-certificates && microdnf clean all

# Create non-root user
RUN groupadd -r gemara && useradd -r -g gemara gemara

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /bin/gemara-mcp-server /app/gemara-mcp-server

# Create artifacts directory placeholder (will be overridden by volume mount)
# Set permissions so any user in the gemara group can write
RUN mkdir -p /app/artifacts && \
    chmod 775 /app && \
    chmod 775 /app/artifacts && \
    chown -R gemara:gemara /app

# Switch to non-root user
USER gemara

# Expose port 8080 for StreamableHTTP transport
EXPOSE 8080

# Default command runs with StreamableHTTP transport on port 8080
CMD ["./gemara-mcp-server", "--transport=streamable-http", "--port=8080"]
