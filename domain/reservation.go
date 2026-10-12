package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// 予約のエンティティ
type Reservation struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	GameID    uuid.UUID
	Status    string
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	Tickets   []*Ticket
}

// チケットのエンティティ
type Ticket struct {
	ID            uuid.UUID
	GameID        uuid.UUID
	SeatID        uuid.UUID
	ReservationID *uuid.UUID
	Price         int
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// 予約の DB 操作に関するインターフェース
type ReservationRepository interface {
	Create(ctx context.Context, tx TxContext, r *Reservation) error
	FindByID(ctx context.Context, id uuid.UUID) (*Reservation, error)
	FindByUserIDWithTickets(ctx context.Context, userID uuid.UUID) ([]*Reservation, error)
	FindByIDWithTickets(ctx context.Context, id uuid.UUID) (*Reservation, error)
	ExpireOverdue(ctx context.Context) error
	UpdateStatus(ctx context.Context, tx TxContext, id uuid.UUID, status string) error
}

// チケットの DB 操作に関するインターフェース
type TicketRepository interface {
	LockAvailable(ctx context.Context, tx TxContext, gameID uuid.UUID, seatGrade string, quantity int) ([]*Ticket, error)
	BulkReserve(ctx context.Context, tx TxContext, ticketIDs []uuid.UUID, reservationID uuid.UUID) error
	FindByReservationID(ctx context.Context, reservationID uuid.UUID) ([]*Ticket, error)
	BulkUpdateStatus(ctx context.Context, tx TxContext, reservationID uuid.UUID, status string) error
	BulkRelease(ctx context.Context, tx TxContext, reservationID uuid.UUID) error
}
