package repository

import (
	"42tokyo-road-to-dena-server/domain"
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// domain.ReservationRepository インターフェースの PostgreSQL 実装
type reservationRepository struct {
	db *sqlx.DB
}

// コンストラクタ
func NewReservationRepository(db *sqlx.DB) domain.ReservationRepository {
	return &reservationRepository{db: db}
}

// 中間構造体
// sqlx は自動マッピングができないため
type reservationRow struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	GameID    uuid.UUID `db:"game_id"`
	Status    string    `db:"status"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// reservationRow から domain.Reservation を組み立てるヘルパー
func toReservationDomain(row reservationRow) *domain.Reservation {
	return &domain.Reservation{
		ID:        row.ID,
		UserID:    row.UserID,
		GameID:    row.GameID,
		Status:    row.Status,
		ExpiresAt: row.ExpiresAt,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

// ============================================================
// Query 定数
// ============================================================

const createReservationQuery = `
	INSERT INTO reservations (id, user_id, game_id, status, expires_at, created_at, updated_at)
	VALUES (:id, :user_id, :game_id, :status, :expires_at, :created_at, :updated_at)
`

const findReservationByIDQuery = `
	SELECT id, user_id, game_id, status, expires_at, created_at, updated_at
	FROM reservations
	WHERE id = $1
`

// 予約を新規作成する
func (r *reservationRepository) Create(ctx context.Context, res *domain.Reservation) error {
	params := map[string]interface{}{
		"id":         res.ID,
		"user_id":    res.UserID,
		"game_id":    res.GameID,
		"status":     res.Status,
		"expires_at": res.ExpiresAt,
		"created_at": res.CreatedAt,
		"updated_at": res.UpdatedAt,
	}
	if _, err := r.db.NamedExecContext(ctx, createReservationQuery, params); err != nil {
		return fmt.Errorf("reservationRepository.Create: %w", err)
	}
	return nil
}

// 指定 ID の予約を取得する
// 存在しない場合は nil, nil を返す
func (r *reservationRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Reservation, error) {
	var row reservationRow
	if err := r.db.GetContext(ctx, &row, findReservationByIDQuery, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("reservationRepository.FindByID: %w", err)
	}
	return toReservationDomain(row), nil
}

// コンパイル時に domain.ReservationRepository インターフェースを満たしているか検証
var _ domain.ReservationRepository = (*reservationRepository)(nil)
