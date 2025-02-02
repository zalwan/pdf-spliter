# Use official Golang image as builder
FROM golang:1.23.5-alpine AS builder

WORKDIR /app

# Copy files and install dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o pdf-splitter

# Use lightweight Alpine image
FROM alpine:latest

WORKDIR /root/

# Install required dependencies
RUN apk --no-cache add ca-certificates

# Copy built binary and templates
COPY --from=builder /app/pdf-splitter .
COPY --from=builder /app/templates ./templates

# Set environment variable for production
ENV GIN_MODE=release

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./pdf-splitter"]
