# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /discord-message-service ./cmd/main.go

# Run stage
FROM alpine:latest

WORKDIR /

COPY --from=builder /discord-message-service .
COPY .env .

EXPOSE 8080

CMD ["/discord-message-service"]