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

type reservationWithTicketRow struct {
	// reservations カラム
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	GameID    uuid.UUID `db:"game_id"`
	Status    string    `db:"status"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	// tickets カラム
	TicketID            *uuid.UUID `db:"ticket_id"`
	TicketGameID        *uuid.UUID `db:"ticket_game_id"`
	TicketSeatID        *uuid.UUID `db:"ticket_seat_id"`
	TicketReservationID *uuid.UUID `db:"ticket_reservation_id"`
	TicketPrice         *int       `db:"ticket_price"`
	TicketStatus        *string    `db:"ticket_status"`
	TicketCreatedAt     *time.Time `db:"ticket_created_at"`
	TicketUpdatedAt     *time.Time `db:"ticket_updated_at"`
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

// フラット行スライスを Reservation + Tickets スライスにグループ化するヘルパー
func groupReservationsWithTickets(rows []reservationWithTicketRow) []*domain.Reservation {
	resMap := make(map[uuid.UUID]*domain.Reservation)
	var result []*domain.Reservation

	for _, row := range rows {
		res, exists := resMap[row.ID]
		if !exists {
			res = &domain.Reservation{
				ID:        row.ID,
				UserID:    row.UserID,
				GameID:    row.GameID,
				Status:    row.Status,
				ExpiresAt: row.ExpiresAt,
				CreatedAt: row.CreatedAt,
				UpdatedAt: row.UpdatedAt,
				Tickets:   make([]*domain.Ticket, 0),
			}
			resMap[row.ID] = res
			// ORDER BY r.created_at DESC, t.id の順序を保持する
			result = append(result, res)
		}
		// チケットが存在する場合のみ追加（LEFT JOIN で NULL になる場合は除外）
		if row.TicketID != nil {
			res.Tickets = append(res.Tickets, &domain.Ticket{
				ID:            *row.TicketID,
				GameID:        *row.TicketGameID,
				SeatID:        *row.TicketSeatID,
				ReservationID: row.TicketReservationID,
				Price:         *row.TicketPrice,
				Status:        *row.TicketStatus,
				CreatedAt:     *row.TicketCreatedAt,
				UpdatedAt:     *row.TicketUpdatedAt,
			})
		}
	}
	return result
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

// reservations + tickets + seats を LEFT JOIN で一括取得
const findReservationsByUserIDWithTicketsQuery = `
	SELECT
		r.id, r.user_id, r.game_id, r.status, r.expires_at,
		r.created_at, r.updated_at,
		t.id           AS ticket_id,
		t.game_id      AS ticket_game_id,
		t.seat_id      AS ticket_seat_id,
		t.reservation_id AS ticket_reservation_id,
		s.price        AS ticket_price,
		t.status       AS ticket_status,
		t.created_at   AS ticket_created_at,
		t.updated_at   AS ticket_updated_at
	FROM reservations r
	LEFT JOIN tickets t ON t.reservation_id = r.id
	LEFT JOIN seats s ON t.seat_id = s.id
	WHERE r.user_id = $1
	ORDER BY r.created_at DESC, t.id
`

const findReservationByIDWithTicketsQuery = `
	SELECT
		r.id, r.user_id, r.game_id, r.status, r.expires_at,
		r.created_at, r.updated_at,
		t.id           AS ticket_id,
		t.game_id      AS ticket_game_id,
		t.seat_id      AS ticket_seat_id,
		t.reservation_id AS ticket_reservation_id,
		s.price        AS ticket_price,
		t.status       AS ticket_status,
		t.created_at   AS ticket_created_at,
		t.updated_at   AS ticket_updated_at
	FROM reservations r
	LEFT JOIN tickets t ON t.reservation_id = r.id
	LEFT JOIN seats s ON t.seat_id = s.id
	WHERE r.id = $1
	ORDER BY t.id
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

// 指定ユーザーの予約一覧をチケット情報付きで取得する
func (r *reservationRepository) FindByUserIDWithTickets(ctx context.Context, userID uuid.UUID) ([]*domain.Reservation, error) {
	var rows []reservationWithTicketRow
	if err := r.db.SelectContext(ctx, &rows, findReservationsByUserIDWithTicketsQuery, userID); err != nil {
		return nil, fmt.Errorf("reservationRepository.FindByUserIDWithTickets: %w", err)
	}
	return groupReservationsWithTickets(rows), nil
}

// 指定 ID の予約をチケット情報付きで取得する
// 存在しない場合は nil, nil を返す
func (r *reservationRepository) FindByIDWithTickets(ctx context.Context, id uuid.UUID) (*domain.Reservation, error) {
	var rows []reservationWithTicketRow
	if err := r.db.SelectContext(ctx, &rows, findReservationByIDWithTicketsQuery, id); err != nil {
		return nil, fmt.Errorf("reservationRepository.FindByIDWithTickets: %w", err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	results := groupReservationsWithTickets(rows)
	return results[0], nil
}

// 期限切れの 'pending' 予約を 'expired' に更新し、紐づくチケットを 'available' に解放する
func (r *reservationRepository) ExpireOverdue(ctx context.Context) error {
	// 予約ステータス更新とチケット解放を同一トランザクションで実行する
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("reservationRepository.ExpireOverdue: begin: %w", err)
	}
	defer tx.Rollback()

	// 期限切れ予約に紐づくチケットを 'available' に解放し reservation_id を NULL に戻す
	_, err = tx.ExecContext(ctx, `
		UPDATE tickets
		SET status = 'available', reservation_id = NULL, updated_at = NOW()
		WHERE reservation_id IN (
			SELECT id FROM reservations
			WHERE status = 'pending' AND expires_at < NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("reservationRepository.ExpireOverdue: release tickets: %w", err)
	}

	// 予約ステータスを 'expired' に更新
	_, err = tx.ExecContext(ctx, `
		UPDATE reservations
		SET status = 'expired', updated_at = NOW()
		WHERE status = 'pending' AND expires_at < NOW()
	`)
	if err != nil {
		return fmt.Errorf("reservationRepository.ExpireOverdue: expire reservations: %w", err)
	}

	return tx.Commit()
}

// コンパイル時に domain.ReservationRepository インターフェースを満たしているか検証
var _ domain.ReservationRepository = (*reservationRepository)(nil)
