package repositorymock

import (
	"42tokyo-road-to-dena-server/domain"
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// ============================================================
// UserRepository モック
// ============================================================

// domain.UserRepository の仮実装
type UserRepository struct {
	CreateFn         func(ctx context.Context, user *domain.User) error
	FindByEmailFn    func(ctx context.Context, email string) (*domain.User, error)
	FindByUsernameFn func(ctx context.Context, username string) (*domain.User, error)
	FindByIDFn       func(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

func (m *UserRepository) Create(ctx context.Context, user *domain.User) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, user)
	}
	return nil
}

func (m *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.FindByEmailFn != nil {
		return m.FindByEmailFn(ctx, email)
	}
	return nil, nil
}

func (m *UserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	if m.FindByUsernameFn != nil {
		return m.FindByUsernameFn(ctx, username)
	}
	return nil, nil
}

func (m *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, nil
}

// コンパイル時に domain.UserRepository インターフェースを満たしているか検証
var _ domain.UserRepository = (*UserRepository)(nil)

// ============================================================
// GameRepository モック
// ============================================================

// domain.GameRepository の仮実装
type GameRepository struct {
	FindAllFn  func(ctx context.Context) ([]*domain.Game, error)
	FindByIDFn func(ctx context.Context, id uuid.UUID) (*domain.Game, error)
}

func (m *GameRepository) FindAll(ctx context.Context) ([]*domain.Game, error) {
	if m.FindAllFn != nil {
		return m.FindAllFn(ctx)
	}
	return nil, nil
}

func (m *GameRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Game, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, nil
}

// コンパイル時に domain.GameRepository インターフェースを満たしているか検証
var _ domain.GameRepository = (*GameRepository)(nil)

// ============================================================
// SeatRepository モック
// ============================================================

// domain.SeatRepository の仮実装
type SeatRepository struct {
	FindAvailabilityByGameIDFn func(ctx context.Context, gameID uuid.UUID) ([]*domain.SeatAvailability, error)
}

func (m *SeatRepository) FindAvailabilityByGameID(ctx context.Context, gameID uuid.UUID) ([]*domain.SeatAvailability, error) {
	if m.FindAvailabilityByGameIDFn != nil {
		return m.FindAvailabilityByGameIDFn(ctx, gameID)
	}
	return nil, nil
}

// コンパイル時に domain.SeatRepository インターフェースを満たしているか検証
var _ domain.SeatRepository = (*SeatRepository)(nil)

// ============================================================
// ReservationRepository モック
// ============================================================

// domain.ReservationRepository の仮実装
type ReservationRepository struct {
	CreateFn                  func(ctx context.Context, tx domain.TxContext, r *domain.Reservation) error
	FindByIDFn                func(ctx context.Context, id uuid.UUID) (*domain.Reservation, error)
	FindByUserIDWithTicketsFn func(ctx context.Context, userID uuid.UUID) ([]*domain.Reservation, error)
	FindByIDWithTicketsFn     func(ctx context.Context, id uuid.UUID) (*domain.Reservation, error)
	ExpireOverdueFn           func(ctx context.Context) error
	UpdateStatusFn            func(ctx context.Context, tx domain.TxContext, id uuid.UUID, status string) error
}

func (m *ReservationRepository) Create(ctx context.Context, tx domain.TxContext, r *domain.Reservation) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, tx, r)
	}
	return nil
}

func (m *ReservationRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Reservation, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *ReservationRepository) FindByUserIDWithTickets(ctx context.Context, userID uuid.UUID) ([]*domain.Reservation, error) {
	if m.FindByUserIDWithTicketsFn != nil {
		return m.FindByUserIDWithTicketsFn(ctx, userID)
	}
	return nil, nil
}

func (m *ReservationRepository) FindByIDWithTickets(ctx context.Context, id uuid.UUID) (*domain.Reservation, error) {
	if m.FindByIDWithTicketsFn != nil {
		return m.FindByIDWithTicketsFn(ctx, id)
	}
	return nil, nil
}

func (m *ReservationRepository) ExpireOverdue(ctx context.Context) error {
	if m.ExpireOverdueFn != nil {
		return m.ExpireOverdueFn(ctx)
	}
	return nil
}

func (m *ReservationRepository) UpdateStatus(ctx context.Context, tx domain.TxContext, id uuid.UUID, status string) error {
	if m.UpdateStatusFn != nil {
		return m.UpdateStatusFn(ctx, tx, id, status)
	}
	return nil
}

// コンパイル時に domain.ReservationRepository インターフェースを満たしているか検証
var _ domain.ReservationRepository = (*ReservationRepository)(nil)

// ============================================================
// TicketRepository モック
// ============================================================

// domain.TicketRepository の仮実装
type TicketRepository struct {
	LockAvailableFn       func(ctx context.Context, tx domain.TxContext, gameID uuid.UUID, quantity int) ([]*domain.Ticket, error)
	BulkReserveFn         func(ctx context.Context, tx domain.TxContext, ticketIDs []uuid.UUID, reservationID uuid.UUID) error
	FindByReservationIDFn func(ctx context.Context, reservationID uuid.UUID) ([]*domain.Ticket, error)
	BulkUpdateStatusFn    func(ctx context.Context, tx domain.TxContext, reservationID uuid.UUID, status string) error
	BulkReleaseFn         func(ctx context.Context, tx domain.TxContext, reservationID uuid.UUID) error
}

func (m *TicketRepository) LockAvailable(ctx context.Context, tx domain.TxContext, gameID uuid.UUID, seatGrade string, quantity int) ([]*domain.Ticket, error) {
	if m.LockAvailableFn != nil {
		return m.LockAvailableFn(ctx, tx, gameID, quantity)
	}
	return nil, nil
}

func (m *TicketRepository) BulkReserve(ctx context.Context, tx domain.TxContext, ticketIDs []uuid.UUID, reservationID uuid.UUID) error {
	if m.BulkReserveFn != nil {
		return m.BulkReserveFn(ctx, tx, ticketIDs, reservationID)
	}
	return nil
}

func (m *TicketRepository) FindByReservationID(ctx context.Context, reservationID uuid.UUID) ([]*domain.Ticket, error) {
	if m.FindByReservationIDFn != nil {
		return m.FindByReservationIDFn(ctx, reservationID)
	}
	return nil, nil
}

func (m *TicketRepository) BulkUpdateStatus(ctx context.Context, tx domain.TxContext, reservationID uuid.UUID, status string) error {
	if m.BulkUpdateStatusFn != nil {
		return m.BulkUpdateStatusFn(ctx, tx, reservationID, status)
	}
	return nil
}

func (m *TicketRepository) BulkRelease(ctx context.Context, tx domain.TxContext, reservationID uuid.UUID) error {
	if m.BulkReleaseFn != nil {
		return m.BulkReleaseFn(ctx, tx, reservationID)
	}
	return nil
}

// コンパイル時に domain.TicketRepository インターフェースを満たしているか検証
var _ domain.TicketRepository = (*TicketRepository)(nil)

// ============================================================
// PasswordHasher モック
// ============================================================

// domain.PasswordHasher の仮実装
type PasswordHasher struct {
	HashPasswordFn  func(password string) (string, error)
	CheckPasswordFn func(password, hash string) bool
}

func (m *PasswordHasher) HashPassword(password string) (string, error) {
	if m.HashPasswordFn != nil {
		return m.HashPasswordFn(password)
	}
	// デフォルト: ハッシュ化せずそのまま返す（テスト用）
	return "hashed_" + password, nil
}

func (m *PasswordHasher) CheckPassword(password, hash string) bool {
	if m.CheckPasswordFn != nil {
		return m.CheckPasswordFn(password, hash)
	}
	return true
}

// コンパイル時に domain.PasswordHasher インターフェースを満たしているか検証
var _ domain.PasswordHasher = (*PasswordHasher)(nil)

// ============================================================
// TxManager / TxContext モック
// ============================================================

// domain.TxManager の仮実装
// 実際の DB トランザクションの手順を完全に無視して、渡された処理（fn）だけをダイレクトに実行する
type TxManager struct {
	// RunInTxFn をセットすることでエラーケース等を注入できる
	RunInTxFn func(ctx context.Context, fn func(tx domain.TxContext) error) error
}

func (m *TxManager) RunInTx(ctx context.Context, fn func(tx domain.TxContext) error) error {
	if m.RunInTxFn != nil {
		return m.RunInTxFn(ctx, fn)
	}
	// デフォルト: そのまま fn を実行する（コミット/ロールバックなし）
	return fn(&TxContext{})
}

// コンパイル時に domain.TxManager インターフェースを満たしているか検証
var _ domain.TxManager = (*TxManager)(nil)

// domain.TxContext の仮実装
type TxContext struct{}

func (t *TxContext) ExecContext(_ context.Context, _ string, _ ...interface{}) (sql.Result, error) {
	return sqlResult{}, nil
}

func (t *TxContext) GetContext(_ context.Context, _ interface{}, _ string, _ ...interface{}) error {
	return nil
}

func (t *TxContext) SelectContext(_ context.Context, _ interface{}, _ string, _ ...interface{}) error {
	return nil
}

// コンパイル時に domain.TxContext インターフェースを満たしているか検証
var _ domain.TxContext = (*TxContext)(nil)

// sql.Result の最小実装
type sqlResult struct{}

func (sqlResult) LastInsertId() (int64, error) { return 0, nil }
func (sqlResult) RowsAffected() (int64, error) { return 1, nil }

// ============================================================
// テストヘルパー（ダミーエンティティ生成）
// ============================================================

// テスト用のダミー User エンティティを生成する
func NewUser(id uuid.UUID, username, email string) *domain.User {
	now := time.Now()
	return &domain.User{
		ID:           id,
		Username:     username,
		Email:        email,
		PasswordHash: "hashed_password",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// テスト用のダミー Game エンティティを生成する
func NewGame(id uuid.UUID) *domain.Game {
	return &domain.Game{
		ID:         id,
		HomeTeamID: uuid.New(),
		AwayTeamID: uuid.New(),
		GameDate:   time.Now().Add(24 * time.Hour),
		StartTime:  time.Now().Add(24 * time.Hour),
		HomeTeam:   domain.Team{ID: uuid.New(), Name: "Home Team"},
		AwayTeam:   domain.Team{ID: uuid.New(), Name: "Away Team"},
	}
}

// テスト用のダミー Ticket エンティティを生成する
func NewTicket(id, gameID uuid.UUID, price int) *domain.Ticket {
	return &domain.Ticket{
		ID:        id,
		GameID:    gameID,
		SeatID:    uuid.New(),
		Price:     price,
		Status:    "available",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// テスト用のダミー Reservation エンティティを生成する
func NewReservation(id, userID, gameID uuid.UUID, status string, expiresAt time.Time) *domain.Reservation {
	now := time.Now()
	return &domain.Reservation{
		ID:        id,
		UserID:    userID,
		GameID:    gameID,
		Status:    status,
		ExpiresAt: expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
		Tickets:   []*domain.Ticket{},
	}
}
