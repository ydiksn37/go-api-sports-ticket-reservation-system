package repository

import (
	"context"
	"fmt"
	"github.com/ydiksn37/go-api-sports-ticket-reservation-system/domain"

	"github.com/jmoiron/sqlx"
)

// domain.TxManager インターフェースの PostgreSQL 実装
type txManager struct {
	db *sqlx.DB
}

// コンストラクタ
func NewTxManager(db *sqlx.DB) domain.TxManager {
	return &txManager{db: db}
}

// トランザクション内で fn を実行する
func (m *txManager) RunInTx(ctx context.Context, fn func(tx domain.TxContext) error) error {
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("txManager.RunInTx: begin: %w", err)
	}

	if err := fn(tx); err != nil {
		// fn がエラーを返した場合はロールバック
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("txManager.RunInTx: rollback: %w (original: %v)", rbErr, err)
		}
		return err
	}

	// 成功時はコミット
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("txManager.RunInTx: commit: %w", err)
	}
	return nil
}

// コンパイル時に domain.TxManager インターフェースを満たしているか検証
var _ domain.TxManager = (*txManager)(nil)
