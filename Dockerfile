# Build stage
FROM golang:1.20-alpine AS builder

RUN apk add --no-cache build-base

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o /app/timer-bot .

# Final stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/timer-bot .

CMD ["/app/timer-bot"]
