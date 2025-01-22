# Build stage
FROM golang:1.23-bookworm AS builder

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o . ./...

# Final stage
FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y ca-certificates tzdata

# Add non-root user
RUN addgroup --system watchdog && adduser --system --group --no-create-home watchdog

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /usr/src/app/watchdog .
RUN chown -R watchdog:watchdog /app

# Use non-root user
USER watchdog

# Expose port
EXPOSE 8080

# Set healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/v1/healthcheck || exit 1

# Run the application
ENTRYPOINT ["/app/watchdog"]
