package domain

import (
	"context"

	"github.com/google/uuid"
)

// 試合ごとのグレード別座席空き状況を表す値オブジェクト
type SeatAvailability struct {
	SeatGrade      string
	Price          int
	TotalSeats     int
	AvailableSeats int
	ReservedSeats  int
	SoldSeats      int
}

// 座席情報の DB 操作に関するインターフェース
type SeatRepository interface {
	// 指定試合の座席空き状況をグレード別に取得する
	FindAvailabilityByGameID(ctx context.Context, gameID uuid.UUID) ([]*SeatAvailability, error)
}
