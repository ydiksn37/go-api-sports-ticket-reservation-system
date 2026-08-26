package domain

import (
	"context"
	"database/sql"
)

// トランザクション内で DB 操作を行うためのインターフェース
type TxContext interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

// トランザクション境界を管理するインターフェース
// Usecase 層がトランザクション範囲を明示的にコントロールするために使用する
type TxManager interface {
	// トランザクション内で fn を実行する
	// fn がエラーを返した場合は自動ロールバック、成功時は自動コミットを行う
	RunInTx(ctx context.Context, fn func(tx TxContext) error) error
}
