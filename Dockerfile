## ---- Builder Stage ----
FROM golang:1.27-alpine AS builder

WORKDIR /app

# Cache module download
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o duckops ./main.go

## ---- Runtime Stage ----
FROM alpine:3.20 AS runtime

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/duckops /usr/local/bin/duckops

# Use non-root user
USER appuser

# Default command (show help)
ENTRYPOINT ["duckops"]
CMD ["--help"]
