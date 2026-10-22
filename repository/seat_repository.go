package repository

import (
	"context"
	"fmt"
	"github.com/ydiksn37/go-api-sports-ticket-reservation-system/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// domain.SeatRepository インターフェースの PostgreSQL 実装
type seatRepository struct {
	db *sqlx.DB
}

// コンストラクタ
func NewSeatRepository(db *sqlx.DB) domain.SeatRepository {
	return &seatRepository{db: db}
}

// 中間構造体
type seatAvailabilityRow struct {
	SeatGrade      string `db:"seat_grade"`
	Price          int    `db:"price"`
	TotalSeats     int    `db:"total_seats"`
	AvailableSeats int    `db:"available_seats"`
	ReservedSeats  int    `db:"reserved_seats"`
	SoldSeats      int    `db:"sold_seats"`
}

// seatAvailabilityRow から domain.SeatAvailability を組み立てるヘルパー
func toSeatDomain(row seatAvailabilityRow) *domain.SeatAvailability {
	return &domain.SeatAvailability{
		SeatGrade:      row.SeatGrade,
		Price:          row.Price,
		TotalSeats:     row.TotalSeats,
		AvailableSeats: row.AvailableSeats,
		ReservedSeats:  row.ReservedSeats,
		SoldSeats:      row.SoldSeats,
	}
}

// Query 定数
const findAvailabilityByGameIDQuery = `
	SELECT
		s.seat_grade,
		s.price,
		s.total_seats,
		COUNT(*) FILTER (WHERE t.status = 'available') AS available_seats,
		COUNT(*) FILTER (WHERE t.status = 'reserved')  AS reserved_seats,
		COUNT(*) FILTER (WHERE t.status = 'sold')      AS sold_seats
	FROM seats s
	JOIN tickets t ON t.seat_id = s.id
	WHERE t.game_id = $1
	GROUP BY s.id
	ORDER BY s.price DESC
`

// 指定試合の座席空き状況をグレード別に取得する
func (r *seatRepository) FindAvailabilityByGameID(ctx context.Context, gameID uuid.UUID) ([]*domain.SeatAvailability, error) {
	var rows []seatAvailabilityRow
	if err := r.db.SelectContext(ctx, &rows, findAvailabilityByGameIDQuery, gameID); err != nil {
		return nil, fmt.Errorf("seatRepository.FindAvailabilityByGameID: %w", err)
	}
	result := make([]*domain.SeatAvailability, 0, len(rows))
	for _, row := range rows {
		result = append(result, toSeatDomain(row))
	}
	return result, nil
}

// コンパイル時に domain.SeatRepository インターフェースを満たしているか検証
var _ domain.SeatRepository = (*seatRepository)(nil)
