package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ydiksn37/go-api-sports-ticket-reservation-system/domain"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// domain.GameRepository インターフェースの PostgreSQL 実装
type gameRepository struct {
	db *sqlx.DB
}

// コンストラクタ
func NewGameRepository(db *sqlx.DB) domain.GameRepository {
	return &gameRepository{db: db}
}

// 中間構造体
type gameRow struct {
	ID           uuid.UUID `db:"id"`
	HomeTeamID   uuid.UUID `db:"home_team_id"`
	AwayTeamID   uuid.UUID `db:"away_team_id"`
	GameDate     time.Time `db:"game_date"`
	StartTime    time.Time `db:"start_time"`
	HomeTeamName string    `db:"home_team_name"`
	AwayTeamName string    `db:"away_team_name"`
}

// gameRow から domain.Game を組み立てるヘルパー
func toGameDomain(row gameRow) *domain.Game {
	return &domain.Game{
		ID:         row.ID,
		HomeTeamID: row.HomeTeamID,
		AwayTeamID: row.AwayTeamID,
		GameDate:   row.GameDate,
		StartTime:  row.StartTime,
		HomeTeam: domain.Team{
			ID:   row.HomeTeamID,
			Name: row.HomeTeamName,
		},
		AwayTeam: domain.Team{
			ID:   row.AwayTeamID,
			Name: row.AwayTeamName,
		},
	}
}

// Query 定数
const findAllGamesQuery = `
	SELECT
		g.id,
		g.home_team_id,
		g.away_team_id,
		g.game_date,
		g.start_time,
		ht.name AS home_team_name,
		at.name AS away_team_name
	FROM games g
	JOIN teams ht ON g.home_team_id = ht.id
	JOIN teams at ON g.away_team_id = at.id
	ORDER BY g.game_date, g.start_time
`

const findGameByIDQuery = `
	SELECT
		g.id,
		g.home_team_id,
		g.away_team_id,
		g.game_date,
		g.start_time,
		ht.name AS home_team_name,
		at.name AS away_team_name
	FROM games g
	JOIN teams ht ON g.home_team_id = ht.id
	JOIN teams at ON g.away_team_id = at.id
	WHERE g.id = $1
`

// 全試合を試合日昇順で取得する
func (r *gameRepository) FindAll(ctx context.Context) ([]*domain.Game, error) {
	var rows []gameRow
	if err := r.db.SelectContext(ctx, &rows, findAllGamesQuery); err != nil {
		return nil, fmt.Errorf("gameRepository.FindAll: %w", err)
	}
	games := make([]*domain.Game, 0, len(rows))
	for _, row := range rows {
		games = append(games, toGameDomain(row))
	}
	return games, nil
}

// 指定 ID の試合を取得する
func (r *gameRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Game, error) {
	var row gameRow
	if err := r.db.GetContext(ctx, &row, findGameByIDQuery, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("gameRepository.FindByID: %w", err)
	}
	return toGameDomain(row), nil
}

// コンパイル時に domain.GameRepository インターフェースを満たしているか検証
var _ domain.GameRepository = (*gameRepository)(nil)
