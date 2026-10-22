package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/ydiksn37/go-api-sports-ticket-reservation-system/domain"
	"github.com/ydiksn37/go-api-sports-ticket-reservation-system/usecase"
	repositorymock "github.com/ydiksn37/go-api-sports-ticket-reservation-system/usecase/mock"
	"testing"
	"time"

	"github.com/google/uuid"
)

// テスト用の ReservationUseCase を組み立てるヘルパー
func newReservationUseCase(
	reservationRepo *repositorymock.ReservationRepository,
	ticketRepo *repositorymock.TicketRepository,
	gameRepo *repositorymock.GameRepository,
	txManager *repositorymock.TxManager,
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
	tx := &repositorymock.TxContext{}

	gameRepo := &repositorymock.GameRepository{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Game, error) {
			return repositorymock.NewGame(id), nil
		},
	}
	ticketRepo := &repositorymock.TicketRepository{
		LockAvailableFn: func(_ context.Context, gotTx domain.TxContext, gID uuid.UUID, qty int) ([]*domain.Ticket, error) {
			if gotTx != tx {
				t.Errorf("LockAvailable: want same transaction, got: %T", gotTx)
			}
			tickets := make([]*domain.Ticket, qty)
			for i := range tickets {
				tickets[i] = repositorymock.NewTicket(ticketID, gID, 5000)
			}
			return tickets, nil
		},
		BulkReserveFn: func(_ context.Context, gotTx domain.TxContext, _ []uuid.UUID, _ uuid.UUID) error {
			if gotTx != tx {
				t.Errorf("BulkReserve: want same transaction, got: %T", gotTx)
			}
			return nil
		},
	}
	reservationRepo := &repositorymock.ReservationRepository{
		CreateFn: func(_ context.Context, gotTx domain.TxContext, _ *domain.Reservation) error {
			if gotTx != tx {
				t.Errorf("Create: want same transaction, got: %T", gotTx)
			}
			return nil
		},
	}

	txManager := &repositorymock.TxManager{
		RunInTxFn: func(ctx context.Context, fn func(domain.TxContext) error) error {
			return fn(tx)
		},
	}
	uc := newReservationUseCase(reservationRepo, ticketRepo, gameRepo, txManager)

	out, err := uc.CreateReservation(context.Background(), usecase.CreateReservationInput{
		UserID: userID,
		GameID: gameID,
		Seats:  []usecase.SeatRequest{{SeatGrade: "SS", Quantity: 2}},
	})

	if err != nil {
		t.Fatalf("want: no error, got: %v", err)
	}
	if out.Reservation == nil {
		t.Fatal("want: reservation, got: nil")
	}
	if out.Reservation.UserID != userID {
		t.Errorf("want: UserID %v, got: %v", userID, out.Reservation.UserID)
	}
	if len(out.Reservation.Tickets) != 2 {
		t.Errorf("want: 2 tickets, got: %d", len(out.Reservation.Tickets))
	}
}

func TestCreateReservation_GameNotFound(t *testing.T) {
	gameRepo := &repositorymock.GameRepository{
		FindByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Game, error) { return nil, nil },
	}

	uc := newReservationUseCase(&repositorymock.ReservationRepository{}, &repositorymock.TicketRepository{}, gameRepo, &repositorymock.TxManager{})

	_, err := uc.CreateReservation(context.Background(), usecase.CreateReservationInput{
		UserID: uuid.New(),
		GameID: uuid.New(),
		Seats:  []usecase.SeatRequest{{SeatGrade: "SS", Quantity: 1}},
	})

	if !errors.Is(err, usecase.ErrGameNotFound) {
		t.Errorf("want: ErrGameNotFound, got: %v", err)
	}
}

func TestCreateReservation_InsufficientSeats(t *testing.T) {
	gameRepo := &repositorymock.GameRepository{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Game, error) {
			return repositorymock.NewGame(id), nil
		},
	}
	ticketRepo := &repositorymock.TicketRepository{
		LockAvailableFn: func(_ context.Context, _ domain.TxContext, _ uuid.UUID, _ int) ([]*domain.Ticket, error) {
			return []*domain.Ticket{}, nil
		},
	}

	uc := newReservationUseCase(&repositorymock.ReservationRepository{}, ticketRepo, gameRepo, &repositorymock.TxManager{})

	_, err := uc.CreateReservation(context.Background(), usecase.CreateReservationInput{
		UserID: uuid.New(),
		GameID: uuid.New(),
		Seats:  []usecase.SeatRequest{{SeatGrade: "SS", Quantity: 2}},
	})

	if !errors.Is(err, usecase.ErrInsufficientSeats) {
		t.Errorf("want: ErrInsufficientSeats, got: %v", err)
	}
}

func TestCreateReservation_DuplicateSeatGrade(t *testing.T) {
	gameRepoCalled := false
	txManagerCalled := false

	gameRepo := &repositorymock.GameRepository{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Game, error) {
			gameRepoCalled = true
			return repositorymock.NewGame(id), nil
		},
	}
	txManager := &repositorymock.TxManager{
		RunInTxFn: func(_ context.Context, _ func(domain.TxContext) error) error {
			txManagerCalled = true
			return nil
		},
	}

	uc := newReservationUseCase(
		&repositorymock.ReservationRepository{},
		&repositorymock.TicketRepository{},
		gameRepo,
		txManager,
	)

	_, err := uc.CreateReservation(context.Background(), usecase.CreateReservationInput{
		UserID: uuid.New(),
		GameID: uuid.New(),
		Seats: []usecase.SeatRequest{
			{SeatGrade: "S", Quantity: 2},
			{SeatGrade: "S", Quantity: 1},
		},
	})

	if !errors.Is(err, usecase.ErrDuplicateSeatGrade) {
		t.Errorf("want: ErrDuplicateSeatGrade, got: %v", err)
	}
	if gameRepoCalled {
		t.Error("game repository must not be called for duplicate seat grades")
	}
	if txManagerCalled {
		t.Error("transaction must not start for duplicate seat grades")
	}
}

func TestCreateReservation_BulkReserveError(t *testing.T) {
	wantErr := errors.New("bulk reserve error")
	gameID := uuid.New()

	gameRepo := &repositorymock.GameRepository{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Game, error) {
			return repositorymock.NewGame(id), nil
		},
	}
	ticketRepo := &repositorymock.TicketRepository{
		LockAvailableFn: func(_ context.Context, _ domain.TxContext, gID uuid.UUID, _ int) ([]*domain.Ticket, error) {
			return []*domain.Ticket{repositorymock.NewTicket(uuid.New(), gID, 5000)}, nil
		},
		BulkReserveFn: func(_ context.Context, _ domain.TxContext, _ []uuid.UUID, _ uuid.UUID) error {
			return wantErr
		},
	}

	uc := newReservationUseCase(
		&repositorymock.ReservationRepository{},
		ticketRepo,
		gameRepo,
		&repositorymock.TxManager{},
	)

	_, err := uc.CreateReservation(context.Background(), usecase.CreateReservationInput{
		UserID: uuid.New(),
		GameID: gameID,
		Seats:  []usecase.SeatRequest{{SeatGrade: "SS", Quantity: 1}},
	})

	if !errors.Is(err, wantErr) {
		t.Errorf("want: BulkReserve error, got: %v", err)
	}
}

func TestCreateReservation_TxError(t *testing.T) {
	wantErr := fmt.Errorf("tx begin error")
	gameRepo := &repositorymock.GameRepository{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Game, error) {
			return repositorymock.NewGame(id), nil
		},
	}
	txManager := &repositorymock.TxManager{
		RunInTxFn: func(_ context.Context, _ func(domain.TxContext) error) error {
			return wantErr
		},
	}

	uc := newReservationUseCase(&repositorymock.ReservationRepository{}, &repositorymock.TicketRepository{}, gameRepo, txManager)

	_, err := uc.CreateReservation(context.Background(), usecase.CreateReservationInput{
		UserID: uuid.New(),
		GameID: uuid.New(),
		Seats:  []usecase.SeatRequest{{SeatGrade: "SS", Quantity: 1}},
	})

	if !errors.Is(err, wantErr) {
		t.Errorf("want: tx begin error, got: %v", err)
	}
}

// ============================================================
// ListReservations テスト
// ============================================================

func TestListReservations_Success(t *testing.T) {
	userID := uuid.New()
	reservations := []*domain.Reservation{
		repositorymock.NewReservation(uuid.New(), userID, uuid.New(), "pending", time.Now().Add(10*time.Minute)),
		repositorymock.NewReservation(uuid.New(), userID, uuid.New(), "confirmed", time.Now().Add(10*time.Minute)),
	}

	reservationRepo := &repositorymock.ReservationRepository{
		ExpireOverdueFn: func(_ context.Context) error { return nil },
		FindByUserIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) ([]*domain.Reservation, error) {
			return reservations, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &repositorymock.TicketRepository{}, &repositorymock.GameRepository{}, &repositorymock.TxManager{})

	out, err := uc.ListReservations(context.Background(), usecase.ListReservationsInput{UserID: userID})

	if err != nil {
		t.Fatalf("want: no error, got: %v", err)
	}
	if len(out.Reservations) != 2 {
		t.Errorf("want: 2 reservations, got: %d", len(out.Reservations))
	}
}

func TestListReservations_ExpireOverdueError(t *testing.T) {
	wantErr := fmt.Errorf("expire overdue error")
	reservationRepo := &repositorymock.ReservationRepository{
		ExpireOverdueFn: func(_ context.Context) error { return wantErr },
	}

	uc := newReservationUseCase(reservationRepo, &repositorymock.TicketRepository{}, &repositorymock.GameRepository{}, &repositorymock.TxManager{})

	_, err := uc.ListReservations(context.Background(), usecase.ListReservationsInput{UserID: uuid.New()})

	if !errors.Is(err, wantErr) {
		t.Errorf("want: expire overdue error, got: %v", err)
	}
}

// ============================================================
// GetReservation テスト
// ============================================================

func TestGetReservation_Success(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	res := repositorymock.NewReservation(reservationID, userID, uuid.New(), "pending", time.Now().Add(10*time.Minute))

	reservationRepo := &repositorymock.ReservationRepository{
		ExpireOverdueFn: func(_ context.Context) error { return nil },
		FindByIDWithTicketsFn: func(_ context.Context, id uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &repositorymock.TicketRepository{}, &repositorymock.GameRepository{}, &repositorymock.TxManager{})

	out, err := uc.GetReservation(context.Background(), usecase.GetReservationInput{
		UserID:        userID,
		ReservationID: reservationID,
	})

	if err != nil {
		t.Fatalf("want: no error, got: %v", err)
	}
	if out.Reservation.ID != reservationID {
		t.Errorf("want: ReservationID %v, got: %v", reservationID, out.Reservation.ID)
	}
}

func TestGetReservation_NotFound(t *testing.T) {
	reservationRepo := &repositorymock.ReservationRepository{
		ExpireOverdueFn: func(_ context.Context) error { return nil },
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return nil, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &repositorymock.TicketRepository{}, &repositorymock.GameRepository{}, &repositorymock.TxManager{})

	_, err := uc.GetReservation(context.Background(), usecase.GetReservationInput{
		UserID:        uuid.New(),
		ReservationID: uuid.New(),
	})

	if !errors.Is(err, usecase.ErrReservationNotFound) {
		t.Errorf("want: ErrReservationNotFound, got: %v", err)
	}
}

func TestGetReservation_Forbidden(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	reservationID := uuid.New()
	res := repositorymock.NewReservation(reservationID, ownerID, uuid.New(), "pending", time.Now().Add(10*time.Minute))

	reservationRepo := &repositorymock.ReservationRepository{
		ExpireOverdueFn: func(_ context.Context) error { return nil },
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &repositorymock.TicketRepository{}, &repositorymock.GameRepository{}, &repositorymock.TxManager{})

	_, err := uc.GetReservation(context.Background(), usecase.GetReservationInput{
		UserID:        otherUserID,
		ReservationID: reservationID,
	})

	if !errors.Is(err, usecase.ErrReservationForbidden) {
		t.Errorf("want: ErrReservationForbidden, got: %v", err)
	}
}

// ============================================================
// PurchaseReservation テスト
// ============================================================

func TestPurchaseReservation_Success(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	res := repositorymock.NewReservation(reservationID, userID, uuid.New(), "pending", time.Now().Add(10*time.Minute))

	reservationRepo := &repositorymock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
		UpdateStatusFn: func(_ context.Context, _ domain.TxContext, _ uuid.UUID, _ string) error {
			return nil
		},
	}
	ticketRepo := &repositorymock.TicketRepository{
		BulkUpdateStatusFn: func(_ context.Context, _ domain.TxContext, _ uuid.UUID, _ string) error {
			return nil
		},
	}

	uc := newReservationUseCase(reservationRepo, ticketRepo, &repositorymock.GameRepository{}, &repositorymock.TxManager{})

	out, err := uc.PurchaseReservation(context.Background(), usecase.PurchaseReservationInput{
		UserID:        userID,
		ReservationID: reservationID,
	})

	if err != nil {
		t.Fatalf("want: no error, got: %v", err)
	}
	if out.Reservation.Status != "confirmed" {
		t.Errorf("want: status 'confirmed', got: %q", out.Reservation.Status)
	}
}

func TestPurchaseReservation_NotFound(t *testing.T) {
	reservationRepo := &repositorymock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return nil, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &repositorymock.TicketRepository{}, &repositorymock.GameRepository{}, &repositorymock.TxManager{})

	_, err := uc.PurchaseReservation(context.Background(), usecase.PurchaseReservationInput{
		UserID:        uuid.New(),
		ReservationID: uuid.New(),
	})

	if !errors.Is(err, usecase.ErrReservationNotFound) {
		t.Errorf("want: ErrReservationNotFound, got: %v", err)
	}
}

func TestPurchaseReservation_Forbidden(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	reservationID := uuid.New()
	res := repositorymock.NewReservation(reservationID, ownerID, uuid.New(), "pending", time.Now().Add(10*time.Minute))

	reservationRepo := &repositorymock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &repositorymock.TicketRepository{}, &repositorymock.GameRepository{}, &repositorymock.TxManager{})

	_, err := uc.PurchaseReservation(context.Background(), usecase.PurchaseReservationInput{
		UserID:        otherUserID,
		ReservationID: reservationID,
	})

	if !errors.Is(err, usecase.ErrReservationForbidden) {
		t.Errorf("want: ErrReservationForbidden, got: %v", err)
	}
}

func TestPurchaseReservation_Expired_ByStatus(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	res := repositorymock.NewReservation(reservationID, userID, uuid.New(), "expired", time.Now().Add(-1*time.Minute))

	reservationRepo := &repositorymock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &repositorymock.TicketRepository{}, &repositorymock.GameRepository{}, &repositorymock.TxManager{})

	_, err := uc.PurchaseReservation(context.Background(), usecase.PurchaseReservationInput{
		UserID:        userID,
		ReservationID: reservationID,
	})

	if !errors.Is(err, usecase.ErrReservationExpired) {
		t.Errorf("want: ErrReservationExpired, got: %v", err)
	}
}

func TestPurchaseReservation_Expired_ByTime(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	res := repositorymock.NewReservation(reservationID, userID, uuid.New(), "pending", time.Now().Add(-1*time.Minute))

	reservationRepo := &repositorymock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &repositorymock.TicketRepository{}, &repositorymock.GameRepository{}, &repositorymock.TxManager{})

	_, err := uc.PurchaseReservation(context.Background(), usecase.PurchaseReservationInput{
		UserID:        userID,
		ReservationID: reservationID,
	})

	if !errors.Is(err, usecase.ErrReservationExpired) {
		t.Errorf("want: ErrReservationExpired, got: %v", err)
	}
}

func TestPurchaseReservation_NotPending(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	res := repositorymock.NewReservation(reservationID, userID, uuid.New(), "confirmed", time.Now().Add(10*time.Minute))

	reservationRepo := &repositorymock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &repositorymock.TicketRepository{}, &repositorymock.GameRepository{}, &repositorymock.TxManager{})

	_, err := uc.PurchaseReservation(context.Background(), usecase.PurchaseReservationInput{
		UserID:        userID,
		ReservationID: reservationID,
	})

	if !errors.Is(err, usecase.ErrReservationNotPending) {
		t.Errorf("want: ErrReservationNotPending, got: %v", err)
	}
}

// ============================================================
// CancelReservation テスト
// ============================================================

func TestCancelReservation_Success(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	res := repositorymock.NewReservation(reservationID, userID, uuid.New(), "pending", time.Now().Add(10*time.Minute))

	reservationRepo := &repositorymock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
		UpdateStatusFn: func(_ context.Context, _ domain.TxContext, _ uuid.UUID, _ string) error {
			return nil
		},
	}
	ticketRepo := &repositorymock.TicketRepository{
		BulkReleaseFn: func(_ context.Context, _ domain.TxContext, _ uuid.UUID) error {
			return nil
		},
	}

	uc := newReservationUseCase(reservationRepo, ticketRepo, &repositorymock.GameRepository{}, &repositorymock.TxManager{})

	err := uc.CancelReservation(context.Background(), usecase.CancelReservationInput{
		UserID:        userID,
		ReservationID: reservationID,
	})

	if err != nil {
		t.Fatalf("want: no error, got: %v", err)
	}
}

func TestCancelReservation_NotFound(t *testing.T) {
	reservationRepo := &repositorymock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return nil, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &repositorymock.TicketRepository{}, &repositorymock.GameRepository{}, &repositorymock.TxManager{})

	err := uc.CancelReservation(context.Background(), usecase.CancelReservationInput{
		UserID:        uuid.New(),
		ReservationID: uuid.New(),
	})

	if !errors.Is(err, usecase.ErrReservationNotFound) {
		t.Errorf("want: ErrReservationNotFound, got: %v", err)
	}
}

func TestCancelReservation_Forbidden(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	reservationID := uuid.New()
	res := repositorymock.NewReservation(reservationID, ownerID, uuid.New(), "pending", time.Now().Add(10*time.Minute))

	reservationRepo := &repositorymock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &repositorymock.TicketRepository{}, &repositorymock.GameRepository{}, &repositorymock.TxManager{})

	err := uc.CancelReservation(context.Background(), usecase.CancelReservationInput{
		UserID:        otherUserID,
		ReservationID: reservationID,
	})

	if !errors.Is(err, usecase.ErrReservationForbidden) {
		t.Errorf("want: ErrReservationForbidden, got: %v", err)
	}
}

func TestCancelReservation_AlreadyPurchased(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	res := repositorymock.NewReservation(reservationID, userID, uuid.New(), "confirmed", time.Now().Add(10*time.Minute))

	reservationRepo := &repositorymock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &repositorymock.TicketRepository{}, &repositorymock.GameRepository{}, &repositorymock.TxManager{})

	err := uc.CancelReservation(context.Background(), usecase.CancelReservationInput{
		UserID:        userID,
		ReservationID: reservationID,
	})

	if !errors.Is(err, usecase.ErrReservationAlreadyPurchased) {
		t.Errorf("want: ErrReservationAlreadyPurchased, got: %v", err)
	}
}

func TestCancelReservation_NotCancellable_Cancelled(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	res := repositorymock.NewReservation(reservationID, userID, uuid.New(), "cancelled", time.Now().Add(10*time.Minute))

	reservationRepo := &repositorymock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &repositorymock.TicketRepository{}, &repositorymock.GameRepository{}, &repositorymock.TxManager{})

	err := uc.CancelReservation(context.Background(), usecase.CancelReservationInput{
		UserID:        userID,
		ReservationID: reservationID,
	})

	if !errors.Is(err, usecase.ErrReservationNotCancellable) {
		t.Errorf("want: ErrReservationNotCancellable, got: %v", err)
	}
}

func TestCancelReservation_NotCancellable_Expired(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	res := repositorymock.NewReservation(reservationID, userID, uuid.New(), "expired", time.Now().Add(10*time.Minute))

	reservationRepo := &repositorymock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &repositorymock.TicketRepository{}, &repositorymock.GameRepository{}, &repositorymock.TxManager{})

	err := uc.CancelReservation(context.Background(), usecase.CancelReservationInput{
		UserID:        userID,
		ReservationID: reservationID,
	})

	if !errors.Is(err, usecase.ErrReservationExpired) {
		t.Errorf("want: ErrReservationExpired, got: %v", err)
	}
}

func TestCancelReservation_Expired_ByTime(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	res := repositorymock.NewReservation(reservationID, userID, uuid.New(), "pending", time.Now().Add(-1*time.Minute))

	reservationRepo := &repositorymock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
	}

	uc := newReservationUseCase(reservationRepo, &repositorymock.TicketRepository{}, &repositorymock.GameRepository{}, &repositorymock.TxManager{})

	err := uc.CancelReservation(context.Background(), usecase.CancelReservationInput{
		UserID:        userID,
		ReservationID: reservationID,
	})

	if !errors.Is(err, usecase.ErrReservationExpired) {
		t.Errorf("want: ErrReservationExpired, got: %v", err)
	}
}

func TestCancelReservation_TxError(t *testing.T) {
	userID := uuid.New()
	reservationID := uuid.New()
	wantErr := fmt.Errorf("tx error")
	res := repositorymock.NewReservation(reservationID, userID, uuid.New(), "pending", time.Now().Add(10*time.Minute))

	reservationRepo := &repositorymock.ReservationRepository{
		FindByIDWithTicketsFn: func(_ context.Context, _ uuid.UUID) (*domain.Reservation, error) {
			return res, nil
		},
	}
	txManager := &repositorymock.TxManager{
		RunInTxFn: func(_ context.Context, _ func(domain.TxContext) error) error {
			return wantErr
		},
	}

	uc := newReservationUseCase(reservationRepo, &repositorymock.TicketRepository{}, &repositorymock.GameRepository{}, txManager)

	err := uc.CancelReservation(context.Background(), usecase.CancelReservationInput{
		UserID:        userID,
		ReservationID: reservationID,
	})

	if !errors.Is(err, wantErr) {
		t.Errorf("want: tx error, got: %v", err)
	}
}
