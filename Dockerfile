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

# Create data directory for database
RUN mkdir -p /app/data

COPY --from=builder /app/timer-bot .

# Set default database location (can be overridden)
ENV DATABASE_URL=/app/data/timerbot.db

CMD ["/app/timer-bot"]
