package repository

import (
	"context"
	"fmt"
	"github.com/ydiksn37/go-api-sports-ticket-reservation-system/domain"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// domain.TicketRepository インターフェースの PostgreSQL 実装
type ticketRepository struct {
	db *sqlx.DB
}

// コンストラクタ
func NewTicketRepository(db *sqlx.DB) domain.TicketRepository {
	return &ticketRepository{db: db}
}

// 中間構造体
type ticketRow struct {
	ID            uuid.UUID  `db:"id"`
	GameID        uuid.UUID  `db:"game_id"`
	SeatID        uuid.UUID  `db:"seat_id"`
	ReservationID *uuid.UUID `db:"reservation_id"`
	Price         int        `db:"price"`
	Status        string     `db:"status"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
}

// ticketRow から domain.Ticket を組み立てるヘルパー
func toTicketDomain(row ticketRow) *domain.Ticket {
	return &domain.Ticket{
		ID:            row.ID,
		GameID:        row.GameID,
		SeatID:        row.SeatID,
		ReservationID: row.ReservationID,
		Price:         row.Price,
		Status:        row.Status,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}

// Query 定数
const lockAvailableTicketsQuery = `
	SELECT
		t.id,
		t.game_id,
		t.seat_id,
		t.reservation_id,
		s.price,
		t.status,
		t.created_at,
		t.updated_at
	FROM tickets t
	JOIN seats s ON t.seat_id = s.id
	WHERE t.game_id = $1
		AND s.seat_grade = $2
		AND t.status = 'available'
	ORDER BY t.id
	LIMIT $3
	FOR UPDATE OF t SKIP LOCKED
`

const bulkReserveTicketsQuery = `
	UPDATE tickets
	SET status = 'reserved',
			reservation_id = $1,
			updated_at = NOW()
	WHERE id = ANY($2)
		AND status = 'available'
`

const findTicketsByReservationIDQuery = `
	SELECT
		t.id,
		t.game_id,
		t.seat_id,
		t.reservation_id,
		s.price,
		t.status,
		t.created_at,
		t.updated_at
	FROM tickets t
	JOIN seats s ON t.seat_id = s.id
	WHERE t.reservation_id = $1
	ORDER BY t.id
`

const bulkUpdateTicketStatusQuery = `
	UPDATE tickets
	SET status = $1, updated_at = NOW()
	WHERE reservation_id = $2
`

const bulkReleaseTicketsQuery = `
	UPDATE tickets
	SET status = 'available', reservation_id = NULL, updated_at = NOW()
	WHERE reservation_id = $1
`

// 指定試合・座席グレードの 'available' チケットを quantity 枚ロック取得する
func (r *ticketRepository) LockAvailable(ctx context.Context, tx domain.TxContext, gameID uuid.UUID, seatGrade string, quantity int) ([]*domain.Ticket, error) {
	var rows []ticketRow
	if err := tx.SelectContext(ctx, &rows, lockAvailableTicketsQuery, gameID, seatGrade, quantity); err != nil {
		return nil, fmt.Errorf("ticketRepository.LockAvailable: %w", err)
	}
	tickets := make([]*domain.Ticket, 0, len(rows))
	for _, row := range rows {
		tickets = append(tickets, toTicketDomain(row))
	}
	return tickets, nil
}

// 指定チケットのステータスを 'reserved' に一括更新し、reservation_id を紐づける
func (r *ticketRepository) BulkReserve(ctx context.Context, tx domain.TxContext, ticketIDs []uuid.UUID, reservationID uuid.UUID) error {
	ids := make([]string, len(ticketIDs))
	for i, id := range ticketIDs {
		ids[i] = id.String()
	}
	result, err := tx.ExecContext(ctx, bulkReserveTicketsQuery, reservationID, pq.Array(ids))
	if err != nil {
		return fmt.Errorf("ticketRepository.BulkReserve: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ticketRepository.BulkReserve: rows affected: %w", err)
	}
	if affected != int64(len(ticketIDs)) {
		return fmt.Errorf(
			"ticketRepository.BulkReserve: updated ticket count mismatch: expected=%d actual=%d",
			len(ticketIDs),
			affected,
		)
	}
	return nil
}

// 指定予約に紐づくチケット一覧を取得する
func (r *ticketRepository) FindByReservationID(ctx context.Context, reservationID uuid.UUID) ([]*domain.Ticket, error) {
	var rows []ticketRow
	if err := r.db.SelectContext(ctx, &rows, findTicketsByReservationIDQuery, reservationID); err != nil {
		return nil, fmt.Errorf("ticketRepository.FindByReservationID: %w", err)
	}
	tickets := make([]*domain.Ticket, 0, len(rows))
	for _, row := range rows {
		tickets = append(tickets, toTicketDomain(row))
	}
	return tickets, nil
}

// 指定予約に紐づく全チケットのステータスを一括更新する
func (r *ticketRepository) BulkUpdateStatus(ctx context.Context, tx domain.TxContext, reservationID uuid.UUID, status string) error {
	if _, err := tx.ExecContext(ctx, bulkUpdateTicketStatusQuery, status, reservationID); err != nil {
		return fmt.Errorf("ticketRepository.BulkUpdateStatus: %w", err)
	}
	return nil
}

// 指定予約に紐づく全チケットを解放する（キャンセル時に使用）
func (r *ticketRepository) BulkRelease(ctx context.Context, tx domain.TxContext, reservationID uuid.UUID) error {
	if _, err := tx.ExecContext(ctx, bulkReleaseTicketsQuery, reservationID); err != nil {
		return fmt.Errorf("ticketRepository.BulkRelease: %w", err)
	}
	return nil
}

// コンパイル時に domain.TicketRepository インターフェースを満たしているか検証
var _ domain.TicketRepository = (*ticketRepository)(nil)
