# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o gowebmail ./cmd/gowebmail

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs tzdata

# Create app user
RUN addgroup -g 1000 gowebmail && \
    adduser -D -u 1000 -G gowebmail gowebmail

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/gowebmail .

# Copy web files
COPY web ./web

# Copy example config
COPY configs/gowebmail.example.yml ./gowebmail.yml

# Create data directory
RUN mkdir -p /app/data && chown -R gowebmail:gowebmail /app

# Switch to app user
USER gowebmail

# Expose ports
EXPOSE 1025 8080

# Set volume for data persistence
VOLUME ["/app/data"]

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Run the application
CMD ["./gowebmail"]
