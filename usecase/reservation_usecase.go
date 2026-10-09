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
	// 予約が見つからない場合のエラー
	ErrReservationNotFound = errors.New("reservation not found")
	// 他ユーザーの予約にアクセスした場合のエラー
	ErrReservationForbidden = errors.New("forbidden: not your reservation")
	// 予約が期限切れの場合のエラー
	ErrReservationExpired = errors.New("reservation expired")
	// 予約が pending 以外のステータスの場合のエラー
	ErrReservationNotPending = errors.New("reservation is not in pending status")
	// 購入済みの予約をキャンセルしようとした場合のエラー
	ErrReservationAlreadyPurchased = errors.New("reservation already purchased")
	// キャンセル済み or 期限切れでキャンセル操作不可の場合のエラー
	ErrReservationNotCancellable = errors.New("reservation cannot be cancelled")
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

// 予約一覧取得の入力データ
type ListReservationsInput struct {
	UserID uuid.UUID
}

// 予約一覧取得の出力データ
type ListReservationsOutput struct {
	Reservations []*domain.Reservation
}

// 予約詳細取得の入力データ
type GetReservationInput struct {
	UserID        uuid.UUID
	ReservationID uuid.UUID
}

// 予約詳細取得の出力データ
type GetReservationOutput struct {
	Reservation *domain.Reservation
}

// 予約確定の入力データ
type PurchaseReservationInput struct {
	UserID        uuid.UUID
	ReservationID uuid.UUID
}

// 予約確定の出力データ
type PurchaseReservationOutput struct {
	Reservation *domain.Reservation
}

// 予約キャンセルの入力データ
type CancelReservationInput struct {
	UserID        uuid.UUID
	ReservationID uuid.UUID
}

// ============================================================
// インターフェース定義
// Handler はこのインターフェースに依存
// ============================================================

// 予約に関するビジネスロジックのインターフェース
type ReservationUseCase interface {
	// 予約を新規作成する
	CreateReservation(ctx context.Context, input CreateReservationInput) (*CreateReservationOutput, error)
	// 認証済みユーザーの予約一覧を取得する
	ListReservations(ctx context.Context, input ListReservationsInput) (*ListReservationsOutput, error)
	// 指定 ID の予約詳細を取得する
	GetReservation(ctx context.Context, input GetReservationInput) (*GetReservationOutput, error)
	// 予約を確定し、チケットを購入済みにする
	PurchaseReservation(ctx context.Context, input PurchaseReservationInput) (*PurchaseReservationOutput, error)
	// 予約をキャンセルし、チケットを解放する
	CancelReservation(ctx context.Context, input CancelReservationInput) error
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
		if err := u.reservationRepo.Create(ctx, tx, res); err != nil {
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

// 予約一覧取得のビジネスロジック
func (u *reservationUseCase) ListReservations(ctx context.Context, input ListReservationsInput) (*ListReservationsOutput, error) {
	// 期限切れ予約のステータス更新 + 紐づくチケット解放
	if err := u.reservationRepo.ExpireOverdue(ctx); err != nil {
		return nil, fmt.Errorf("ListReservations: expire overdue: %w", err)
	}

	// 認証済みユーザーの予約一覧を返す
	reservations, err := u.reservationRepo.FindByUserIDWithTickets(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("ListReservations: %w", err)
	}
	return &ListReservationsOutput{Reservations: reservations}, nil
}

// 予約詳細取得のビジネスロジック
func (u *reservationUseCase) GetReservation(ctx context.Context, input GetReservationInput) (*GetReservationOutput, error) {
	// 期限切れ予約のステータス更新 + 紐づくチケット解放
	if err := u.reservationRepo.ExpireOverdue(ctx); err != nil {
		return nil, fmt.Errorf("GetReservation: expire overdue: %w", err)
	}

	// 指定 ID の予約詳細を返す
	reservation, err := u.reservationRepo.FindByIDWithTickets(ctx, input.ReservationID)
	if err != nil {
		return nil, fmt.Errorf("GetReservation: %w", err)
	}
	if reservation == nil {
		return nil, ErrReservationNotFound
	}
	// 所有権チェック: 認証済みユーザー自身の予約のみ取得可能
	if reservation.UserID != input.UserID {
		return nil, ErrReservationForbidden
	}
	return &GetReservationOutput{Reservation: reservation}, nil
}

// 予約確定のビジネスロジック
func (u *reservationUseCase) PurchaseReservation(ctx context.Context, input PurchaseReservationInput) (*PurchaseReservationOutput, error) {
	// 予約をチケット付きで取得
	reservation, err := u.reservationRepo.FindByIDWithTickets(ctx, input.ReservationID)
	if err != nil {
		return nil, fmt.Errorf("PurchaseReservation: %w", err)
	}
	if reservation == nil {
		return nil, ErrReservationNotFound
	}

	// 所有権チェック: 認証済みユーザー自身の予約のみ購入可能
	if reservation.UserID != input.UserID {
		return nil, ErrReservationForbidden
	}

	// ステータスチェック: pending でなければ購入不可
	if reservation.Status != "pending" {
		if reservation.Status == "expired" {
			return nil, ErrReservationExpired
		}
		return nil, ErrReservationNotPending
	}

	// 期限チェック: status が pending でも期限を超えていれば期限切れ扱い
	if time.Now().After(reservation.ExpiresAt) {
		return nil, ErrReservationExpired
	}

	// トランザクション内で Reservation と Ticket を一括更新
	txErr := u.txManager.RunInTx(ctx, func(tx domain.TxContext) error {
		if err := u.reservationRepo.UpdateStatus(ctx, tx, reservation.ID, "confirmed"); err != nil {
			return fmt.Errorf("UpdateStatus reservation: %w", err)
		}
		if err := u.ticketRepo.BulkUpdateStatus(ctx, tx, reservation.ID, "sold"); err != nil {
			return fmt.Errorf("BulkUpdateStatus tickets: %w", err)
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	// 更新後の状態をエンティティに反映してレスポンス用に返す
	reservation.Status = "confirmed"
	for _, t := range reservation.Tickets {
		t.Status = "sold"
	}

	return &PurchaseReservationOutput{Reservation: reservation}, nil
}

// 予約キャンセルのビジネスロジック
func (u *reservationUseCase) CancelReservation(ctx context.Context, input CancelReservationInput) error {
	// チケット込みで予約を取得
	reservation, err := u.reservationRepo.FindByIDWithTickets(ctx, input.ReservationID)
	if err != nil {
		return fmt.Errorf("CancelReservation: %w", err)
	}
	if reservation == nil {
		return ErrReservationNotFound
	}

	// 所有権チェック: 認証済みユーザー自身の予約のみキャンセル可能
	if reservation.UserID != input.UserID {
		return ErrReservationForbidden
	}

	// 購入済みチェック: confirmed の予約はキャンセル不可
	if reservation.Status == "confirmed" {
		return ErrReservationAlreadyPurchased
	}

	// キャンセル済み or 期限切れの予約もキャンセル不可
	if reservation.Status == "cancelled" || reservation.Status == "expired" {
		return ErrReservationNotCancellable
	}

	// キャンセル期限チェック: expires_at を過ぎた pending 予約はキャンセル不可
	if time.Now().After(reservation.ExpiresAt) {
		return ErrReservationExpired
	}

	// トランザクション内で Reservation + Ticket を同時更新（原子性の担保）
	return u.txManager.RunInTx(ctx, func(tx domain.TxContext) error {
		// Reservation を更新
		if err := u.reservationRepo.UpdateStatus(ctx, tx, reservation.ID, "cancelled"); err != nil {
			return fmt.Errorf("UpdateStatus reservation: %w", err)
		}
		// Ticket を更新
		// status='available', reservation_id=NULL を同時更新
		if err := u.ticketRepo.BulkRelease(ctx, tx, reservation.ID); err != nil {
			return fmt.Errorf("BulkRelease tickets: %w", err)
		}
		return nil
	})
}

// コンパイル時に ReservationUseCase インターフェースを満たしているか検証
var _ ReservationUseCase = (*reservationUseCase)(nil)
