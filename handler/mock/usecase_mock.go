package handlermock

import (
	"42tokyo-road-to-dena-server/usecase"
	"context"
)

// ============================================================
// UserUseCase モック
// ============================================================

// usecase.UserUseCase の仮実装
type UserUseCase struct {
	SignupFn func(ctx context.Context, input usecase.SignupInput) (*usecase.SignupOutput, error)
	MeFn     func(ctx context.Context, input usecase.MeInput) (*usecase.MeOutput, error)
}

func (m *UserUseCase) Signup(ctx context.Context, input usecase.SignupInput) (*usecase.SignupOutput, error) {
	if m.SignupFn != nil {
		return m.SignupFn(ctx, input)
	}
	return nil, nil
}

func (m *UserUseCase) Me(ctx context.Context, input usecase.MeInput) (*usecase.MeOutput, error) {
	if m.MeFn != nil {
		return m.MeFn(ctx, input)
	}
	return nil, nil
}

// コンパイル時に usecase.UserUseCase インターフェースを満たしているか検証
var _ usecase.UserUseCase = (*UserUseCase)(nil)

// ============================================================
// GameUseCase モック
// ============================================================

// usecase.GameUseCase の仮実装
type GameUseCase struct {
	ListGamesFn func(ctx context.Context) (*usecase.ListGamesOutput, error)
	GetGameFn   func(ctx context.Context, input usecase.GetGameInput) (*usecase.GetGameOutput, error)
}

func (m *GameUseCase) ListGames(ctx context.Context) (*usecase.ListGamesOutput, error) {
	if m.ListGamesFn != nil {
		return m.ListGamesFn(ctx)
	}
	return &usecase.ListGamesOutput{}, nil
}

func (m *GameUseCase) GetGame(ctx context.Context, input usecase.GetGameInput) (*usecase.GetGameOutput, error) {
	if m.GetGameFn != nil {
		return m.GetGameFn(ctx, input)
	}
	return nil, nil
}

// コンパイル時に usecase.GameUseCase インターフェースを満たしているか検証
var _ usecase.GameUseCase = (*GameUseCase)(nil)

// ============================================================
// SeatUseCase モック
// ============================================================

// usecase.SeatUseCase の仮実装
type SeatUseCase struct {
	ListSeatAvailabilityFn func(ctx context.Context, input usecase.ListSeatAvailabilityInput) (*usecase.ListSeatAvailabilityOutput, error)
}

func (m *SeatUseCase) ListSeatAvailability(ctx context.Context, input usecase.ListSeatAvailabilityInput) (*usecase.ListSeatAvailabilityOutput, error) {
	if m.ListSeatAvailabilityFn != nil {
		return m.ListSeatAvailabilityFn(ctx, input)
	}
	return &usecase.ListSeatAvailabilityOutput{}, nil
}

// コンパイル時に usecase.SeatUseCase インターフェースを満たしているか検証
var _ usecase.SeatUseCase = (*SeatUseCase)(nil)

// ============================================================
// ReservationUseCase モック
// ============================================================

// usecase.ReservationUseCase の仮実装
type ReservationUseCase struct {
	CreateReservationFn   func(ctx context.Context, input usecase.CreateReservationInput) (*usecase.CreateReservationOutput, error)
	ListReservationsFn    func(ctx context.Context, input usecase.ListReservationsInput) (*usecase.ListReservationsOutput, error)
	GetReservationFn      func(ctx context.Context, input usecase.GetReservationInput) (*usecase.GetReservationOutput, error)
	PurchaseReservationFn func(ctx context.Context, input usecase.PurchaseReservationInput) (*usecase.PurchaseReservationOutput, error)
	CancelReservationFn   func(ctx context.Context, input usecase.CancelReservationInput) error
}

func (m *ReservationUseCase) CreateReservation(ctx context.Context, input usecase.CreateReservationInput) (*usecase.CreateReservationOutput, error) {
	if m.CreateReservationFn != nil {
		return m.CreateReservationFn(ctx, input)
	}
	return nil, nil
}

func (m *ReservationUseCase) ListReservations(ctx context.Context, input usecase.ListReservationsInput) (*usecase.ListReservationsOutput, error) {
	if m.ListReservationsFn != nil {
		return m.ListReservationsFn(ctx, input)
	}
	return &usecase.ListReservationsOutput{}, nil
}

func (m *ReservationUseCase) GetReservation(ctx context.Context, input usecase.GetReservationInput) (*usecase.GetReservationOutput, error) {
	if m.GetReservationFn != nil {
		return m.GetReservationFn(ctx, input)
	}
	return nil, nil
}

func (m *ReservationUseCase) PurchaseReservation(ctx context.Context, input usecase.PurchaseReservationInput) (*usecase.PurchaseReservationOutput, error) {
	if m.PurchaseReservationFn != nil {
		return m.PurchaseReservationFn(ctx, input)
	}
	return nil, nil
}

func (m *ReservationUseCase) CancelReservation(ctx context.Context, input usecase.CancelReservationInput) error {
	if m.CancelReservationFn != nil {
		return m.CancelReservationFn(ctx, input)
	}
	return nil
}

// コンパイル時に usecase.ReservationUseCase インターフェースを満たしているか検証
var _ usecase.ReservationUseCase = (*ReservationUseCase)(nil)
