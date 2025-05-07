# Stage 1: Builder
FROM golang:1.21 AS builder

WORKDIR /app

# Copy go.mod and go.sum and download dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go binary
RUN go build -o /app/coupon_server ./cmd/coupon_server

# Stage 2: Runner
FROM alpine:latest

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/coupon_server /app/coupon_server

# Copy the database schema
COPY internal/storage/database/schema.sql /app/schema.sql

# Expose the application port
EXPOSE 8080

# Run the application
ENTRYPOINT ["/app/coupon_server"]

# Note: For data persistence, you should mount a volume for the SQLite database file.
# Example: docker run -p 8080:8080 -v /path/on/host/db:/app coupon-server
# The database file will be created as 'coupons.db' inside the working directory (/app)
# if it doesn't exist when the application starts.