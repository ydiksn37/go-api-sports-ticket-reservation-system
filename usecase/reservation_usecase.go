package usecase

import (
	"42tokyo-road-to-dena-server/domain"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ============================================================
// センチネルエラー定義
// ============================================================

var (
	// 指定グレードの空き座席が不足している場合のエラー
	ErrInsufficientSeats = errors.New("insufficient available seats")
)

// ============================================================
// DTO 定義
// ============================================================

// 座席グレードと枚数の入力
type SeatRequest struct {
	SeatGrade string
	Quantity  int
}

// 予約作成の入力データ
type CreateReservationInput struct {
	UserID uuid.UUID
	GameID uuid.UUID
	Seats  []SeatRequest
}

// 予約作成の出力データ
type CreateReservationOutput struct {
	Reservation *domain.Reservation
}

// ============================================================
// インターフェース定義
// Handler はこのインターフェースに依存
// ============================================================

// 予約に関するビジネスロジックのインターフェース
type ReservationUseCase interface {
	// 予約を新規作成する
	CreateReservation(ctx context.Context, input CreateReservationInput) (*CreateReservationOutput, error)
}

// ============================================================
// UseCase ロジック
// ============================================================

// DIP のため各 domain インターフェースに依存
type reservationUseCase struct {
	reservationRepo domain.ReservationRepository
	ticketRepo      domain.TicketRepository
	gameRepo        domain.GameRepository
	txManager       domain.TxManager
}

// コンストラクタ
func NewReservationUseCase(
	reservationRepo domain.ReservationRepository,
	ticketRepo domain.TicketRepository,
	gameRepo domain.GameRepository,
	txManager domain.TxManager,
) ReservationUseCase {
	return &reservationUseCase{
		reservationRepo: reservationRepo,
		ticketRepo:      ticketRepo,
		gameRepo:        gameRepo,
		txManager:       txManager,
	}
}

// 予約作成のビジネスロジック
func (u *reservationUseCase) CreateReservation(ctx context.Context, input CreateReservationInput) (*CreateReservationOutput, error) {
	// ゲーム存在確認
	game, err := u.gameRepo.FindByID(ctx, input.GameID)
	if err != nil {
		return nil, fmt.Errorf("CreateReservation: %w", err)
	}
	if game == nil {
		return nil, ErrGameNotFound
	}

	// トランザクション内で Reservation + Ticket を処理
	now := time.Now()
	res := &domain.Reservation{
		ID:        uuid.New(),
		UserID:    input.UserID,
		GameID:    input.GameID,
		Status:    "pending",
		ExpiresAt: now.Add(15 * time.Minute),
		CreatedAt: now,
		UpdatedAt: now,
	}

	var allTickets []*domain.Ticket

	txErr := u.txManager.RunInTx(ctx, func(tx domain.TxContext) error {
		var reserved []*domain.Ticket

		// 各グレードについてチケットをロック取得
		for _, seat := range input.Seats {
			tickets, err := u.ticketRepo.LockAvailable(ctx, tx, input.GameID, seat.SeatGrade, seat.Quantity)
			if err != nil {
				return fmt.Errorf("LockAvailable(%s): %w", seat.SeatGrade, err)
			}
			// 残席不足チェック
			if len(tickets) < seat.Quantity {
				return fmt.Errorf("%w: grade=%s requested=%d available=%d",
					ErrInsufficientSeats, seat.SeatGrade, seat.Quantity, len(tickets))
			}
			reserved = append(reserved, tickets...)
		}

		// Reservation を DB に保存
		if err := u.reservationRepo.Create(ctx, res); err != nil {
			return fmt.Errorf("reservationRepo.Create: %w", err)
		}

		// チケット ID を収集して一括更新
		ticketIDs := make([]uuid.UUID, 0, len(reserved))
		for _, t := range reserved {
			ticketIDs = append(ticketIDs, t.ID)
		}
		if err := u.ticketRepo.BulkReserve(ctx, tx, ticketIDs, res.ID); err != nil {
			return fmt.Errorf("ticketRepo.BulkReserve: %w", err)
		}

		allTickets = reserved
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	// Reservation にチケット一覧を付与して返す
	for _, t := range allTickets {
		// BulkReserve で更新済みの状態を反映するため reservation_id と status を手動で補完
		t.ReservationID = &res.ID
		t.Status = "reserved"
	}
	res.Tickets = allTickets

	return &CreateReservationOutput{Reservation: res}, nil
}

// コンパイル時に ReservationUseCase インターフェースを満たしているか検証
var _ ReservationUseCase = (*reservationUseCase)(nil)
