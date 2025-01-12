# Message Service API

メッセージを管理するためのREST APIサービス

## ディレクトリ構成

```text
.
├── cmd
│   └── api
│       └── main.go           # エントリーポイント
├── pkg
│   └── api                   # 生成コード
│       └── openapi.gen.go    # oapi-codegenの出力
└── internal
    ├── domain
    │   └── message          # メッセージドメイン
    │       ├── entity.go    # ドメインモデル
    │       └── repository.go # リポジトリインターフェース
    ├── adapter
    │   ├── handler         # HTTPハンドラー
    │   └── repository      # MongoDB実装
    └── infrastructure      # 技術的関心
        ├── config
        └── mongodb
```

## セットアップ

### 必要条件

- Go 1.21+
- MongoDB 5.0

### 環境変数

```env
PORT_BACKEND=8080
MONGO_ROOT_USERNAME=root
MONGO_ROOT_PASSWORD=password
```

## API仕様

[仕様：OpenAPI](https://fungifur-strikers.github.io/message-service/src/openapi/)

### メッセージ登録

```http
POST /api/message
Content-Type: application/json

{
    "uid": "msg123",
    "sent_at": "2024-01-04T10:00:00Z",
    "sender": "user1",
    "channel_id": "ch1",
    "content": "Hello World"
}
```

### メッセージ検索

```http
GET /api/message/search?channel_id=ch1&from_date=2024-01-01T00:00:00Z
```

### メッセージ削除

```http
DELETE /api/message/msg123
```

## 開発

### OpenAPI仕様の更新

```bash
docker compose exec backend oapi-codegen -config config.yaml /openapi/index.yaml
```
