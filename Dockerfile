FROM golang:1.23.5-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the application
RUN go build -o pdf-splitter

# Create a smaller final image
FROM alpine:latest

WORKDIR /root/

# Install required dependencies
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/pdf-splitter .

# Copy the templates directory
COPY --from=builder /app/templates ./templates

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./pdf-splitter"]
