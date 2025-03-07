# Message Service

メッセージングプラットフォーム向けの RESTful API サービス。チャンネルベースのメッセージを管理し、認証機能を提供します。

## 主要機能

### メッセージ管理

- チャンネルごとのメッセージの作成・削除
- 送信者、チャンネル、日時による高度な検索機能
- タイムスタンプと一意の ID によるメッセージ管理

### 認証・認可

- Bearer 認証による API アクセス制御
- 有効期限付きアクセストークンの発行・管理
- トークンの無効化機能

## 技術スタック

### バックエンド

![Go](https://img.shields.io/badge/go-1.22.0-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Gin](https://img.shields.io/badge/gin-v1.10.0-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![MongoDB](https://img.shields.io/badge/MongoDB-v1.17.1-47A248?style=for-the-badge&logo=mongodb&logoColor=white)
![OpenAPI](https://img.shields.io/badge/OpenAPI-3.0-6BA539?style=for-the-badge&logo=openapi-initiative&logoColor=white)

### 開発ツール & インフラ

![Docker](https://img.shields.io/badge/Docker-compose-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![Redoc](https://img.shields.io/badge/Redoc-2.0.0-8CA1AF?style=for-the-badge&logo=read-the-docs&logoColor=white)
![Air](https://img.shields.io/badge/Air-Live_Reload-00ADD8?style=for-the-badge&logo=go&logoColor=white)

## プロジェクト構造

```
message-service/
├── .air.toml          # Live reload configuration
├── .docker/           # Docker related files
├── compose.yml        # Docker compose configuration (開発環境用)
├── compose.prod.yml   # Docker compose configuration (本番環境用)
├── src/
│   ├── backend/       # Go backend application
│   │   ├── cmd/      # Application entrypoints
│   │   ├── internal/ # Internal packages
│   │   │   ├── adapter/       # Interface adapters
│   │   │   ├── domain/        # Domain layer
│   │   │   └── infrastructure/# Infrastructure layer
│   │   └── go.mod    # Go modules file
│   └── openapi/      # OpenAPI/Swagger documentation
```

## セットアップ

### 必要条件

- Go 1.22.0 以上
- Docker & Docker Compose
- Make (オプション)

### 環境構築

1. リポジトリのクローン

```bash
git clone https://github.com/yourusername/message-service.git
cd message-service
```

2. 環境変数の設定

```bash
cp .env.example .env
# .envファイルを編集して必要な環境変数を設定
```

必要な環境変数：

- `BACKEND_PORT`: バックエンドサービスのポート
- `MONGO_INITDB_ROOT_USERNAME`: MongoDB の root 用ユーザー名
- `MONGO_INITDB_ROOT_PASSWORD`: MongoDB の root 用パスワード
- `MONGODB_NAME`: データベース名
- `MONGO_EXPRESS_PORT`: Mongo Express のポート
- `MONGO_EXPRESS_BASICAUTH_USERNAME`: Mongo Express 用の基本認証ユーザー名
- `MONGO_EXPRESS_BASICAUTH_PASSWORD`: Mongo Express 用の基本認証パスワード
- `REDOC_PORT`: API ドキュメントのポート

3. アプリケーションの起動

```bash
docker compose up -d
```

## 本番環境へのデプロイ

[![Publish Docker Image](https://github.com/FungiFur-Strikers/message-service/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/FungiFur-Strikers/message-service/actions/workflows/docker-publish.yml)

本番環境では、開発用の機能を除外した `compose.prod.yml` を使用します。このプロジェクトは Docker Hub に公開されたイメージを使用してデプロイできます：

1. 環境変数の設定

```bash
cp .env.example .env
# .envファイルを編集して必要な環境変数を設定
```

2. デプロイの実行

```bash
# Docker Hubからイメージをプル
DOCKERHUB_USERNAME=your-dockerhub-username docker compose -f compose.prod.yml pull

# サービスの起動
DOCKERHUB_USERNAME=your-dockerhub-username docker compose -f compose.prod.yml up -d
```

### 本番環境の特徴

- Docker Hub に公開された最適化されたイメージを使用
- 開発ツール（Mongo Express、Redoc）が除外されています
- コンテナの自動再起動が有効化されています
- リソース制限が設定されています（CPU、メモリ）
- ログローテーションが設定されています
- ホットリロードが無効化され、最適化されたバイナリを直接実行します

### 本番環境用の環境変数

本番環境では以下の環境変数が必要です：

- `DOCKERHUB_USERNAME`: Docker Hub のユーザー名
- `BACKEND_PORT`: バックエンドサービスのポート
- `MONGO_INITDB_ROOT_USERNAME`: MongoDB の root 用ユーザー名
- `MONGO_INITDB_ROOT_PASSWORD`: MongoDB の root 用パスワード
- `MONGODB_NAME`: データベース名

## 提供サービス

起動後、以下のサービスにアクセスできます：

- API サーバー: `http://localhost:{BACKEND_PORT}`
- API ドキュメント: `http://localhost:{REDOC_PORT}`
- データベース管理 UI: `http://localhost:{MONGO_EXPRESS_PORT}`

## API ドキュメント

OpenAPI (Swagger) ドキュメントは以下の URL で確認できます：

```
http://localhost:{REDOC_PORT}
```

[仕様：OpenAPI](https://fungifur-strikers.github.io/message-service/src/openapi/)

API ドキュメントは[Redoc](https://github.com/Redocly/redoc)を使用して生成され、自動的に更新されます。

## 開発ガイド

### アーキテクチャ

このプロジェクトはクリーンアーキテクチャの原則に従って構築されています：

- `domain`: ビジネスロジックとエンティティ
- `adapter`: 外部インターフェースの実装（HTTP ハンドラーなど）
- `infrastructure`: 外部サービスとの統合（データベース、認証など）

### 開発環境の準備

1. 依存関係のインストール

```bash
cd src/backend
go mod download
```

2. ホットリロードでの開発

```bash
air
```

### データベース管理

MongoDB 管理用の Web UI には以下の URL からアクセスできます：

```
http://localhost:{MONGO_EXPRESS_PORT}
```

### OpenAPI 仕様の更新

```bash
docker compose exec backend oapi-codegen -config config.yaml /openapi/index.yaml
```

### テスト

[![Pull Request Checks](https://github.com/FungiFur-Strikers/message-service/actions/workflows/pull-request.yml/badge.svg)](https://github.com/FungiFur-Strikers/message-service/actions/workflows/pull-request.yml)

プロジェクトには以下のテストが含まれています：

- ユニットテスト：ドメインロジックとエンティティのテスト
- 統合テスト：ハンドラーとリポジトリの統合テスト
- E2E テスト：API エンドポイントの動作確認

#### ローカルでのテスト実行

```bash
# すべてのテストを実行
go test ./...

# テストカバレッジレポートの生成
go test -coverprofile=coverage.out ./...

# カバレッジレポートの表示（HTML形式）
go tool cover -html=coverage.out

# 特定のパッケージのテストを実行
go test ./internal/domain/message/...
go test ./internal/adapter/handler/...
```

#### コンテナ上でのテスト実行

Docker Compose を使用してコンテナ上でテストを実行できます：

```bash
# テストコンテナを起動してテストを実行
docker compose -f compose.test.yml up --build --abort-on-container-exit

# テスト終了後にコンテナを削除
docker compose -f compose.test.yml down
```

テスト時の注意事項：

- テストデータベースは自動的に作成・クリーンアップされます
- モックオブジェクトは `testify/mock` を使用して生成されています
- 環境変数は自動的にテスト用の値に置き換えられます
- コンテナ上でのテスト実行時は、専用のテスト用 MongoDB インスタンスが自動的に起動します

## ライセンス

[MIT](LICENSE)
