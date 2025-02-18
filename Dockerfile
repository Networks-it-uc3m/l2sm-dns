# Stage 1: Build the Go application
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Install any required dependencies
RUN apk add --no-cache git

# Copy the Go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go binary
RUN go build -o server ./cmd/server/

# Stage 2: Create a minimal runtime environment
FROM alpine:3.18

# Set the working directory inside the container
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/server /app/server

# Copy any necessary configuration files (optional)
COPY config/server /app/config/server

# Expose the port your service will be running on
EXPOSE 8081

# Command to run the Go binary
ENTRYPOINT ["/app/server"]
