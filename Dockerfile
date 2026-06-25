# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.25-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/server ./cmd/server

# Runtime stage
FROM gcr.io/distroless/static-debian12:nonroot

# Install certificates for TLS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /bin/server /server

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/server"]
