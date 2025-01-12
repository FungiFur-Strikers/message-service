FROM golang:1.23-alpine
WORKDIR /app

RUN go install github.com/air-verse/air@latest
RUN go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]
