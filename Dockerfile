# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go

# Final stage
FROM alpine:3.21

WORKDIR /app

# Add a non-root user for security
RUN adduser -D appuser
USER appuser

# Copy the binary from the build stage
COPY --from=builder /app/server .

# Expose ports (App and Metrics)
EXPOSE 8080 9091

# Run the server
CMD ["./server"]
