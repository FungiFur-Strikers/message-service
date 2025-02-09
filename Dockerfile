FROM golang:1.22.0-alpine AS builder

WORKDIR /app

# 必要なパッケージのインストール
RUN apk add --no-cache gcc musl-dev

# 依存関係のコピーとダウンロード
COPY src/backend/go.mod src/backend/go.sum ./
RUN go mod download

# ソースコードのコピー
COPY src/backend ./

# アプリケーションのビルド
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# 最終イメージの作成
FROM alpine:latest

WORKDIR /app

# タイムゾーンデータのインストール
RUN apk --no-cache add tzdata

# ビルドしたバイナリのコピー
COPY --from=builder /app/main .

CMD ["/app/main"]