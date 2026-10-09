package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// 予約のエンティティ
// DB の reservations テーブルと対応
type Reservation struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	GameID    uuid.UUID
	Status    string // "pending" | "confirmed" | "cancelled" | "expired"
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	Tickets   []*Ticket // インピーダンス・ミスマッチ
}

// チケットのエンティティ
// DB の tickets テーブルと対応
type Ticket struct {
	ID            uuid.UUID
	GameID        uuid.UUID
	SeatID        uuid.UUID
	ReservationID *uuid.UUID // nullable (未予約の場合は nil)
	Price         int        // seats.price を JOIN して取得
	Status        string     // "available" | "reserved" | "sold"
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// 予約の DB 操作に関するインターフェース
type ReservationRepository interface {
	// 予約を新規作成する
	Create(ctx context.Context, tx TxContext, r *Reservation) error
	// 指定 ID の予約を取得する
	FindByID(ctx context.Context, id uuid.UUID) (*Reservation, error)
	// 指定ユーザーの予約一覧をチケット情報付きで取得する
	FindByUserIDWithTickets(ctx context.Context, userID uuid.UUID) ([]*Reservation, error)
	// 指定 ID の予約をチケット情報付きで取得する
	FindByIDWithTickets(ctx context.Context, id uuid.UUID) (*Reservation, error)
	// 期限切れの 'pending' 予約を 'expired' に更新し、紐づくチケットを 'available' に解放する
	ExpireOverdue(ctx context.Context) error
	// 指定 ID の予約ステータスを更新する
	UpdateStatus(ctx context.Context, tx TxContext, id uuid.UUID, status string) error
}

// チケットの DB 操作に関するインターフェース
type TicketRepository interface {
	// 指定試合・座席グレードの 'available' チケットを quantity 枚ロック取得する
	LockAvailable(ctx context.Context, tx TxContext, gameID uuid.UUID, seatGrade string, quantity int) ([]*Ticket, error)
	// 指定チケットのステータスを 'reserved' に一括更新し、reservation_id を紐づける
	BulkReserve(ctx context.Context, tx TxContext, ticketIDs []uuid.UUID, reservationID uuid.UUID) error
	// 指定予約に紐づくチケット一覧を取得する
	FindByReservationID(ctx context.Context, reservationID uuid.UUID) ([]*Ticket, error)
	// 指定予約に紐づく全チケットのステータスを一括更新する
	BulkUpdateStatus(ctx context.Context, tx TxContext, reservationID uuid.UUID, status string) error
	// 指定予約に紐づく全チケットを解放する（キャンセル時に使用）
	BulkRelease(ctx context.Context, tx TxContext, reservationID uuid.UUID) error
}
