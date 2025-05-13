# Build
FROM golang:alpine AS builder

WORKDIR /app

RUN apk add gcc
RUN apk add musl-dev

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o /app/coupon_server ./cmd/coupon_server

# Run
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/coupon_server /app/coupon_server

EXPOSE 8080
ENTRYPOINT ["/app/coupon_server"]