package usecase_test

import (
	"42tokyo-road-to-dena-server/domain"
	"42tokyo-road-to-dena-server/usecase"
	"42tokyo-road-to-dena-server/usecase/mock"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ============================================================
// テストヘルパー
// ============================================================

// テスト用の ReservationUseCase を組み立てるヘルパー
func newReservationUseCase(
	reservationRepo *mock.ReservationRepository,
	ticketRepo *mock.TicketRepository,
	gameRepo *mock.GameRepository,
	txManager *mock.TxManager,
) usecase.ReservationUseCase {
	return usecase.NewReservationUseCase(reservationRepo, ticketRepo, gameRepo, txManager)
}

// ============================================================
// CreateReservation テスト
// ============================================================

func TestCreateReservation_Success(t *testing.T) {
	gameID := uuid.New()
	userID := uuid.New()
	ticketID := uuid.New()

	gameRepo := &mock.GameRepository{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Game, error) {
			return mock.NewGame(id), nil
		},
	}
	ticketRepo := &mock.TicketRepository{
		LockAvailableFn: func(_ context.Context, _ domain.TxContext, gID uuid.UUID, qty int) ([]*domain.Ticket, error) {
			// qty 枚分のチケットを返す
			tickets := make([]*domain.Ticket, qty)
			for i := range tickets {
				tickets[i] = mock.NewTicket(ticketID, gID, 5000)
			}
			return tickets, nil
		},
		BulkReserveFn: func(_ context.Context, _ domain.TxContext, _ []uuid.UUID, _ uuid.UUID) error {
			return nil
		},
	}
	reservationRepo := &mock.ReservationRepository{
		CreateFn: func(_ context.Context, _ *domain.Reservation) error { return nil },
	}

	uc := newReservationUseCase(reservationRepo, ticketRepo, gameRepo, &mock.TxManager{})

	out, err := uc.CreateReservation(context.Background(), usecase.CreateReservationInput{
		UserID: userID,
		GameID: gameID,
		Seats:  []usecase.SeatRequest{{SeatGrade: "S指定席", Quantity: 2}},
	})

	if err != nil {
		t.Fatalf("expected: no error, got %v", err)
	}
	if out.Reservation == nil {
		t.Fatal("expected: reservation, got nil")
	}
	if out.Reservation.UserID != userID {
		t.Errorf("expected: UserID %v, got %v", userID, out.Reservation.UserID)
	}
	if len(out.Reservation.Tickets) != 2 {
		t.Errorf("expected: 2 tickets, got %d", len(out.Reservation.Tickets))
	}
}

func TestCreateReservation_GameNotFound(t *testing.T) {
	gameRepo := &mock.GameRepository{
		FindByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Game, error) { return nil, nil },
	}

	uc := newReservationUseCase(&mock.ReservationRepository{}, &mock.TicketRepository{}, gameRepo, &mock.TxManager{})

	_, err := uc.CreateReservation(context.Background(), usecase.CreateReservationInput{
		UserID: uuid.New(),
		GameID: uuid.New(),
		Seats:  []usecase.SeatRequest{{SeatGrade: "S指定席", Quantity: 1}},
	})

	if !errors.Is(err, usecase.ErrGameNotFound) {
		t.Errorf("expected: ErrGameNotFound, got %v", err)
	}
}

func TestCreateReservation_InsufficientSeats(t *testing.T) {
	gameRepo := &mock.GameRepository{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Game, error) {
			return mock.NewGame(id), nil
		},
	}
	ticketRepo := &mock.TicketRepository{
		// 空スライスを返す = 残席 0
		LockAvailableFn: func(_ context.Context, _ domain.TxContext, _ uuid.UUID, _ int) ([]*domain.Ticket, error) {
			return []*domain.Ticket{}, nil
		},
	}

	uc := newReservationUseCase(&mock.ReservationRepository{}, ticketRepo, gameRepo, &mock.TxManager{})

	_, err := uc.CreateReservation(context.Background(), usecase.CreateReservationInput{
		UserID: uuid.New(),
		GameID: uuid.New(),
		Seats:  []usecase.SeatRequest{{SeatGrade: "S指定席", Quantity: 2}},
	})

	if !errors.Is(err, usecase.ErrInsufficientSeats) {
		t.Errorf("expected ErrInsufficientSeats, got %v", err)
	}
}

func TestCreateReservation_TxError(t *testing.T) {
	wantErr := fmt.Errorf("tx begin error")
	gameRepo := &mock.GameRepository{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Game, error) {
			return mock.NewGame(id), nil
		},
	}
	txManager := &mock.TxManager{
		RunInTxFn: func(_ context.Context, _ func(domain.TxContext) error) error {
			return wantErr
		},
	}

	uc := newReservationUseCase(&mock.ReservationRepository{}, &mock.TicketRepository{}, gameRepo, txManager)

	_, err := uc.CreateReservation(context.Background(), usecase.CreateReservationInput{
		UserID: uuid.New(),
		GameID: uuid.New(),
		Seats:  []usecase.SeatRequest{{SeatGrade: "S指定席", Quantity: 1}},
	})

	if !errors.Is(err, wantErr) {
		t.Errorf("expected: tx begin error, got %v", err)
	}
}

// ============================================================
// ListReservations テスト
// ============================================================

func TestListReservations_Success(t *testing.T) {
	userID := uuid.New()
	reservations := []*domain.Reservation{
		mock.NewReservation(uuid.New(), userID, uuid.New(), "pending", time.Now().Add(10*time.Minute)),
		mock.NewReservation(uuid.New(), userID, uuid.New(), "confirmed", time.Now().Add(10*time.Minute)),
	}

	reservationRepo := &mock.ReservationRepository{
		ExpireOverdueFn: func(_ context.Context) error { return nil },
		FindByUserIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) ([]*domain.Reservation, error) {
			return reservations, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &mock.TicketRepository{}, &mock.GameRepository{}, &mock.TxManager{})

	out, err := uc.ListReservations(context.Background(), usecase.ListReservationsInput{UserID: userID})

	if err != nil {
		t.Fatalf("expected: no error, got %v", err)
	}
	if len(out.Reservations) != 2 {
		t.Errorf("expected: 2 reservations, got %d", len(out.Reservations))
	}
}

func TestListReservations_ExpireOverdueError(t *testing.T) {
	wantErr := fmt.Errorf("expire overdue error")
	reservationRepo := &mock.ReservationRepository{
		ExpireOverdueFn: func(_ context.Context) error { return wantErr },
	}

	uc := newReservationUseCase(reservationRepo, &mock.TicketRepository{}, &mock.GameRepository{}, &mock.TxManager{})

	_, err := uc.ListReservations(context.Background(), usecase.ListReservationsInput{UserID: uuid.New()})

	if !errors.Is(err, wantErr) {
		t.Errorf("expected: expire overdue error, got %v", err)
	}
}

// ============================================================
// GetReservation テスト
// ============================================================

func TestGetReservation_Success(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	res := mock.NewReservation(reservationID, userID, uuid.New(), "pending", time.Now().Add(10*time.Minute))

	reservationRepo := &mock.ReservationRepository{
		ExpireOverdueFn: func(_ context.Context) error { return nil },
		FindByIDWithTicketsFn: func(_ context.Context, id uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &mock.TicketRepository{}, &mock.GameRepository{}, &mock.TxManager{})

	out, err := uc.GetReservation(context.Background(), usecase.GetReservationInput{
		UserID:        userID,
		ReservationID: reservationID,
	})

	if err != nil {
		t.Fatalf("expected: no error, got %v", err)
	}
	if out.Reservation.ID != reservationID {
		t.Errorf("expected: ReservationID %v, got %v", reservationID, out.Reservation.ID)
	}
}

func TestGetReservation_NotFound(t *testing.T) {
	reservationRepo := &mock.ReservationRepository{
		ExpireOverdueFn: func(_ context.Context) error { return nil },
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return nil, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &mock.TicketRepository{}, &mock.GameRepository{}, &mock.TxManager{})

	_, err := uc.GetReservation(context.Background(), usecase.GetReservationInput{
		UserID:        uuid.New(),
		ReservationID: uuid.New(),
	})

	if !errors.Is(err, usecase.ErrReservationNotFound) {
		t.Errorf("expected: ErrReservationNotFound, got %v", err)
	}
}

func TestGetReservation_Forbidden(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	reservationID := uuid.New()
	res := mock.NewReservation(reservationID, ownerID, uuid.New(), "pending", time.Now().Add(10*time.Minute))

	reservationRepo := &mock.ReservationRepository{
		ExpireOverdueFn: func(_ context.Context) error { return nil },
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &mock.TicketRepository{}, &mock.GameRepository{}, &mock.TxManager{})

	// otherUserID で ownerID の予約を取得しようとする
	_, err := uc.GetReservation(context.Background(), usecase.GetReservationInput{
		UserID:        otherUserID,
		ReservationID: reservationID,
	})

	if !errors.Is(err, usecase.ErrReservationForbidden) {
		t.Errorf("expected: ErrReservationForbidden, got %v", err)
	}
}

// ============================================================
// PurchaseReservation テスト
// ============================================================

func TestPurchaseReservation_Success(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	res := mock.NewReservation(reservationID, userID, uuid.New(), "pending", time.Now().Add(10*time.Minute))

	reservationRepo := &mock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
		UpdateStatusFn: func(_ context.Context, _ domain.TxContext, _ uuid.UUID, _ string) error {
			return nil
		},
	}
	ticketRepo := &mock.TicketRepository{
		BulkUpdateStatusFn: func(_ context.Context, _ domain.TxContext, _ uuid.UUID, _ string) error {
			return nil
		},
	}

	uc := newReservationUseCase(reservationRepo, ticketRepo, &mock.GameRepository{}, &mock.TxManager{})

	out, err := uc.PurchaseReservation(context.Background(), usecase.PurchaseReservationInput{
		UserID:        userID,
		ReservationID: reservationID,
	})

	if err != nil {
		t.Fatalf("expected: no error, got %v", err)
	}
	if out.Reservation.Status != "confirmed" {
		t.Errorf("expected: status 'confirmed', got %q", out.Reservation.Status)
	}
}

func TestPurchaseReservation_NotFound(t *testing.T) {
	reservationRepo := &mock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return nil, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &mock.TicketRepository{}, &mock.GameRepository{}, &mock.TxManager{})

	_, err := uc.PurchaseReservation(context.Background(), usecase.PurchaseReservationInput{
		UserID:        uuid.New(),
		ReservationID: uuid.New(),
	})

	if !errors.Is(err, usecase.ErrReservationNotFound) {
		t.Errorf("expected: ErrReservationNotFound, got %v", err)
	}
}

func TestPurchaseReservation_Forbidden(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	reservationID := uuid.New()
	res := mock.NewReservation(reservationID, ownerID, uuid.New(), "pending", time.Now().Add(10*time.Minute))

	reservationRepo := &mock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &mock.TicketRepository{}, &mock.GameRepository{}, &mock.TxManager{})

	_, err := uc.PurchaseReservation(context.Background(), usecase.PurchaseReservationInput{
		UserID:        otherUserID,
		ReservationID: reservationID,
	})

	if !errors.Is(err, usecase.ErrReservationForbidden) {
		t.Errorf("expected: ErrReservationForbidden, got %v", err)
	}
}

func TestPurchaseReservation_Expired_ByStatus(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	// status が "expired" の予約
	res := mock.NewReservation(reservationID, userID, uuid.New(), "expired", time.Now().Add(-1*time.Minute))

	reservationRepo := &mock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &mock.TicketRepository{}, &mock.GameRepository{}, &mock.TxManager{})

	_, err := uc.PurchaseReservation(context.Background(), usecase.PurchaseReservationInput{
		UserID:        userID,
		ReservationID: reservationID,
	})

	if !errors.Is(err, usecase.ErrReservationExpired) {
		t.Errorf("expected: ErrReservationExpired, got %v", err)
	}
}

func TestPurchaseReservation_Expired_ByTime(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	// status は pending だが期限超過
	res := mock.NewReservation(reservationID, userID, uuid.New(), "pending", time.Now().Add(-1*time.Minute))

	reservationRepo := &mock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &mock.TicketRepository{}, &mock.GameRepository{}, &mock.TxManager{})

	_, err := uc.PurchaseReservation(context.Background(), usecase.PurchaseReservationInput{
		UserID:        userID,
		ReservationID: reservationID,
	})

	if !errors.Is(err, usecase.ErrReservationExpired) {
		t.Errorf("expected: ErrReservationExpired, got %v", err)
	}
}

func TestPurchaseReservation_NotPending(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	// status が "confirmed" = pending 以外
	res := mock.NewReservation(reservationID, userID, uuid.New(), "confirmed", time.Now().Add(10*time.Minute))

	reservationRepo := &mock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &mock.TicketRepository{}, &mock.GameRepository{}, &mock.TxManager{})

	_, err := uc.PurchaseReservation(context.Background(), usecase.PurchaseReservationInput{
		UserID:        userID,
		ReservationID: reservationID,
	})

	if !errors.Is(err, usecase.ErrReservationNotPending) {
		t.Errorf("expected: ErrReservationNotPending, got %v", err)
	}
}

// ============================================================
// CancelReservation テスト
// ============================================================

func TestCancelReservation_Success(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	res := mock.NewReservation(reservationID, userID, uuid.New(), "pending", time.Now().Add(10*time.Minute))

	reservationRepo := &mock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
		UpdateStatusFn: func(_ context.Context, _ domain.TxContext, _ uuid.UUID, _ string) error {
			return nil
		},
	}
	ticketRepo := &mock.TicketRepository{
		BulkReleaseFn: func(_ context.Context, _ domain.TxContext, _ uuid.UUID) error {
			return nil
		},
	}

	uc := newReservationUseCase(reservationRepo, ticketRepo, &mock.GameRepository{}, &mock.TxManager{})

	err := uc.CancelReservation(context.Background(), usecase.CancelReservationInput{
		UserID:        userID,
		ReservationID: reservationID,
	})

	if err != nil {
		t.Fatalf("expected: no error, got %v", err)
	}
}

func TestCancelReservation_NotFound(t *testing.T) {
	reservationRepo := &mock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return nil, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &mock.TicketRepository{}, &mock.GameRepository{}, &mock.TxManager{})

	err := uc.CancelReservation(context.Background(), usecase.CancelReservationInput{
		UserID:        uuid.New(),
		ReservationID: uuid.New(),
	})

	if !errors.Is(err, usecase.ErrReservationNotFound) {
		t.Errorf("expected: ErrReservationNotFound, got %v", err)
	}
}

func TestCancelReservation_Forbidden(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	reservationID := uuid.New()
	res := mock.NewReservation(reservationID, ownerID, uuid.New(), "pending", time.Now().Add(10*time.Minute))

	reservationRepo := &mock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &mock.TicketRepository{}, &mock.GameRepository{}, &mock.TxManager{})

	err := uc.CancelReservation(context.Background(), usecase.CancelReservationInput{
		UserID:        otherUserID,
		ReservationID: reservationID,
	})

	if !errors.Is(err, usecase.ErrReservationForbidden) {
		t.Errorf("expected: ErrReservationForbidden, got %v", err)
	}
}

func TestCancelReservation_AlreadyPurchased(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	// status が "confirmed" = 購入済み
	res := mock.NewReservation(reservationID, userID, uuid.New(), "confirmed", time.Now().Add(10*time.Minute))

	reservationRepo := &mock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &mock.TicketRepository{}, &mock.GameRepository{}, &mock.TxManager{})

	err := uc.CancelReservation(context.Background(), usecase.CancelReservationInput{
		UserID:        userID,
		ReservationID: reservationID,
	})

	if !errors.Is(err, usecase.ErrReservationAlreadyPurchased) {
		t.Errorf("expected: ErrReservationAlreadyPurchased, got %v", err)
	}
}

func TestCancelReservation_NotCancellable_Cancelled(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	// status が "cancelled" = 既にキャンセル済み
	res := mock.NewReservation(reservationID, userID, uuid.New(), "cancelled", time.Now().Add(10*time.Minute))

	reservationRepo := &mock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &mock.TicketRepository{}, &mock.GameRepository{}, &mock.TxManager{})

	err := uc.CancelReservation(context.Background(), usecase.CancelReservationInput{
		UserID:        userID,
		ReservationID: reservationID,
	})

	if !errors.Is(err, usecase.ErrReservationNotCancellable) {
		t.Errorf("expected: ErrReservationNotCancellable, got %v", err)
	}
}

func TestCancelReservation_NotCancellable_Expired(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	// status が "expired"
	res := mock.NewReservation(reservationID, userID, uuid.New(), "expired", time.Now().Add(10*time.Minute))

	reservationRepo := &mock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &mock.TicketRepository{}, &mock.GameRepository{}, &mock.TxManager{})

	err := uc.CancelReservation(context.Background(), usecase.CancelReservationInput{
		UserID:        userID,
		ReservationID: reservationID,
	})

	if !errors.Is(err, usecase.ErrReservationNotCancellable) {
		t.Errorf("expected: ErrReservationNotCancellable, got %v", err)
	}
}

func TestCancelReservation_Expired_ByTime(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	// status は pending だが期限超過
	res := mock.NewReservation(reservationID, userID, uuid.New(), "pending", time.Now().Add(-1*time.Minute))

	reservationRepo := &mock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &mock.TicketRepository{}, &mock.GameRepository{}, &mock.TxManager{})

	err := uc.CancelReservation(context.Background(), usecase.CancelReservationInput{
		UserID:        userID,
		ReservationID: reservationID,
	})

	if !errors.Is(err, usecase.ErrReservationExpired) {
		t.Errorf("expected: ErrReservationExpired, got %v", err)
	}
}

func TestCancelReservation_TxError(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	wantErr := fmt.Errorf("tx error")
	res := mock.NewReservation(reservationID, userID, uuid.New(), "pending", time.Now().Add(10*time.Minute))

	reservationRepo := &mock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
	}
	txManager := &mock.TxManager{
		RunInTxFn: func(_ context.Context, _ func(domain.TxContext) error) error {
			return wantErr
		},
	}

	uc := newReservationUseCase(reservationRepo, &mock.TicketRepository{}, &mock.GameRepository{}, txManager)

	err := uc.CancelReservation(context.Background(), usecase.CancelReservationInput{
		UserID:        userID,
		ReservationID: reservationID,
	})

	if !errors.Is(err, wantErr) {
		t.Errorf("expected: tx error, got %v", err)
	}
}
