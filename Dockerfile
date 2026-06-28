# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build without CGO (pure Go SQLite driver)
RUN CGO_ENABLED=0 GOOS=linux go build -o anticheat-server .

# Runtime stage
FROM alpine:3.20

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata wget

# Create app user
RUN adduser -D -u 1000 acuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/anticheat-server .

# Copy static files for web dashboard
COPY --from=builder /app/web/static /app/web/static

# Create directories for runtime data
RUN mkdir -p /app/data/screenshots && \
    chown -R acuser:acuser /app

USER acuser

EXPOSE 27915
EXPOSE 27916

VOLUME ["/app/data"]

ENTRYPOINT ["./anticheat-server"]
CMD ["--listen", "0.0.0.0:27915", "--web", "0.0.0.0:27916", "--db", "/app/data/anticheat.db", "--screenshots", "/app/data/screenshots"]