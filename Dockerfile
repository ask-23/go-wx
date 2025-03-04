# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /build

# Install dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go.mod and go.sum files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -o go-wx ./cmd/go-wx

# Final stage
FROM alpine:3.18

# Set working directory
WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache tzdata ca-certificates

# Create necessary directories
RUN mkdir -p /app/logs /app/config /app/web/templates /app/web/static

# Copy binary from builder stage
COPY --from=builder /build/go-wx /app/

# Copy web files
COPY web/templates /app/web/templates
COPY web/static /app/web/static

# Copy default config
COPY config/config.yaml /app/config/

# Expose ports
EXPOSE 8080 8000

# Set environment variables
ENV TZ=UTC

# Run the application
CMD ["/app/go-wx", "--config", "/app/config/config.yaml"] 