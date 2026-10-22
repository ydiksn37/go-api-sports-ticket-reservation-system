# Sports Ticket Reservation API

スポーツイベントのチケット予約を管理するREST APIです。試合と座席の参照、ユーザー登録、チケットの仮予約・購入確定・キャンセルを提供します。

## 主な機能

- JWTによる認証（CookieまたはBearerトークン）
- 試合一覧・試合詳細の取得
- 座席グレードごとの在庫確認
- 複数チケットの仮予約と有効期限管理
- 予約の購入確定・キャンセル
- PostgreSQLトランザクションと行ロックによる在庫管理
- OpenAPI仕様書とSwagger UI

## 技術スタック

- Go 1.24.5（標準ライブラリ `net/http`）
- PostgreSQL 15
- sqlx
- golang-migrate
- Docker / Docker Compose

## アーキテクチャ

依存関係を `handler`、`usecase`、`domain`、`repository` に分離しています。HTTP層は入出力と認証を扱い、ユースケース層が予約処理を調整し、リポジトリ層がPostgreSQLへのアクセスを担当します。

データモデルは [ER.md](ER.md) で確認できます。

## セットアップ

### 必要なもの

- Docker
- Docker Compose
- Make（任意）

### Dockerで起動する

```bash
git clone git@github.com:ydiksn37/go-api-sports-ticket-reservation-system.git
cd go-api-sports-ticket-reservation-system
cp .env.example .env
make initialize
```

`make initialize` はコンテナを起動し、マイグレーションと開発用seedデータの投入を行います。APIは `http://localhost:8080` で利用できます。

既存のPostgreSQLボリュームを新しい設定で作り直す場合は、保存データが不要であることを確認してから次を実行してください。

```bash
make fclean
make initialize
```

### ローカルで実行する

Goから直接サーバーを起動する場合は、PostgreSQLを用意して `DATABASE_URL` をローカルホスト向けに設定します。

```bash
cp .env.example .env
# DATABASE_URLのホストを db から localhost に変更
go run ./cmd/seed
go run ./cmd/server
```

## 環境変数

| 変数 | 用途 |
| --- | --- |
| `POSTGRES_USER` | Docker上のPostgreSQLユーザー |
| `POSTGRES_PASSWORD` | Docker上のPostgreSQLパスワード |
| `POSTGRES_DB` | データベース名 |
| `DATABASE_URL` | PostgreSQL接続URL |
| `JWT_SECRET` | JWT署名用秘密鍵。本番環境では必ず変更してください |
| `JWT_ISSUER` | JWT発行者 |
| `JWT_AUDIENCE` | JWT対象者 |
| `COOKIE_DOMAIN` | 認証CookieのDomain属性 |

## API

| Method | Path | 説明 | 認証 |
| --- | --- | --- | --- |
| `GET` | `/health` | ヘルスチェック | 不要 |
| `POST` | `/api/user/signup` | ユーザー登録とトークン発行 | 不要 |
| `GET` | `/api/user/me` | ログインユーザー情報 | 必要 |
| `GET` | `/api/games` | 試合一覧 | 不要 |
| `GET` | `/api/games/{id}` | 試合詳細 | 不要 |
| `GET` | `/api/games/{id}/seats` | 座席在庫 | 不要 |
| `POST` | `/api/reservations` | 仮予約作成 | 必要 |
| `GET` | `/api/reservations` | 自分の予約一覧 | 必要 |
| `GET` | `/api/reservations/{id}` | 予約詳細 | 必要 |
| `PUT` | `/api/reservations/{id}/purchase` | 購入確定 | 必要 |
| `DELETE` | `/api/reservations/{id}` | 予約キャンセル | 必要 |

詳細なリクエスト・レスポンスは、起動後に [Swagger UI](http://localhost:8080/swagger/) で確認できます。OpenAPI定義は [docs/openapi.yaml](docs/openapi.yaml) です。

## 開発コマンド

```bash
make help      # コマンド一覧
make test      # ユニットテスト
make logs      # コンテナログ
make exec-app  # appコンテナのシェル
make exec-db   # PostgreSQLコンソール
make clean     # コンテナを停止
```

Dockerを使わずテストする場合は次を実行します。

```bash
go test ./...
```
