# Discord Message Service

Discord Message Service は、Discord のメッセージを保存し、検索する機能を提供する Go 言語で書かれたサービスです。

## 機能

- Discord メッセージの保存
- キーワードによるメッセージ検索

## 技術スタック

- Go
- Gin (Web フレームワーク)
- GORM (ORM)
- PostgreSQL (データベース)
- Docker & Docker Compose (コンテナ化)
- Air (ホットリロード)

## プロジェクト構造

```
discord-message-service/
├── cmd/
│   └── main.go
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── message_handler.go
│   │   │   └── search_handler.go
│   │   └── routes.go
│   ├── config/
│   │   └── config.go
│   ├── models/
│   │   └── message.go
│   ├── repository/
│   │   └── message_repository.go
│   └── service/
│       └── message_service.go
├── pkg/
│   └── database/
│       └── postgres.go
├── .env
├── .gitignore
├── go.mod
├── go.sum
├── Dockerfile
├── compose.yml
└── .air.toml
```

## セットアップ

### 前提条件

- Docker と Docker Compose がインストールされていること

### 開発環境のセットアップ

1. リポジトリをクローンします：

   ```
   git clone https://github.com/FungiFur-Strikers/discord-message-service.git
   cd discord-message-service
   ```

2. 環境変数ファイルを作成します：

   ```
   cp .env.example .env
   ```

   `.env` ファイルを編集し、必要な環境変数を設定します。

3. 開発用 Docker コンテナを起動します：

   ```
   docker compose up --build
   ```

   これにより、アプリケーションと PostgreSQL データベースが起動します。

4. アプリケーションは `http://localhost:8081` で利用可能になります。

### 本番環境のビルドとデプロイ

1. 本番用 Docker イメージをビルドします：

   ```
   docker build -t discord-message-service .
   ```

2. 本番環境用の `compose.prd.yml` を使用してサービスを起動します：

   ```
   docker compose -f compose.prd.yml up -d
   ```

## API エンドポイント

- `POST /api/messages`: 新しいメッセージを作成
- `GET /api/messages/search`: メッセージを検索

## 開発

- コードを変更すると、Air がホットリロードを行い、自動的にアプリケーションを再起動します。
- 新しい機能を追加する場合は、適切なテストを書いてください。

## テスト

テストを実行するには、以下のコマンドを使用します：

```
go test ./...
```
