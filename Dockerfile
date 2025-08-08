# Build stage
FROM golang:1.24-alpine AS builder

# Install git (needed for some Go modules)
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Install swag CLI for generating swagger docs
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Generate swagger documentation in the cmd/server directory
RUN swag init -g cmd/server/main.go -o ./cmd/server/docs

# Build the application from the correct path
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy the docs folder (swagger documentation) from cmd/server
COPY --from=builder /app/cmd/server/docs ./docs

COPY --from=builder /app/db/migrations ./db/migrations

# Expose port
EXPOSE 8080

# Command to run
CMD ["./main"]