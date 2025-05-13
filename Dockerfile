# Build
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o /app/coupon_server ./cmd/coupon_server

# Run
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/coupon_server /app/coupon_server

EXPOSE 8080
ENTRYPOINT ["/app/coupon_server"]