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
type TxManager interface {
	RunInTx(ctx context.Context, fn func(tx TxContext) error) error
}
