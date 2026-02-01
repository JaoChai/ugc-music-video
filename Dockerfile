# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download && go mod tidy

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ugc ./cmd/ugc

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ffmpeg for video processing
RUN apk add --no-cache ffmpeg ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/ugc .

# Copy migrations
COPY --from=builder /app/internal/database/migrations ./internal/database/migrations

# Expose port
EXPOSE 8080

# Run
CMD ["./ugc"]
