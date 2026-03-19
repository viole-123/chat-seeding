FROM golang:1.23-alpine AS builder
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Enable automatic toolchain download
ENV GOTOOLCHAIN=auto

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o seeding-bot ./cmd/seeding-bot

FROM alpine:3.18
WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary
COPY --from=builder /app/seeding-bot /app/seeding-bot

# Copy config files and assets
COPY config.docker.yaml /app/config.docker.yaml
COPY personas.yaml /app/personas.yaml
COPY assets /app/assets
COPY internal/adapter/postgres/migrations /app/migrations

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/seeding-bot"]