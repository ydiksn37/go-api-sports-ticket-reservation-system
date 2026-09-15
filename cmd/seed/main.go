package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Team struct {
	ID   uuid.UUID `db:"id"`
	Name string    `db:"name"`
}

type Seat struct {
	ID         uuid.UUID `db:"id"`
	SeatGrade  string    `db:"seat_grade"`
	Price      int       `db:"price"`
	TotalSeats int       `db:"total_seats"`
}

type Game struct {
	ID         uuid.UUID `db:"id"`
	HomeTeamID uuid.UUID `db:"home_team_id"`
	AwayTeamID uuid.UUID `db:"away_team_id"`
	GameDate   time.Time `db:"game_date"`
	StartTime  time.Time `db:"start_time"`
}

type Ticket struct {
	ID     uuid.UUID `db:"id"`
	GameID uuid.UUID `db:"game_id"`
	SeatID uuid.UUID `db:"seat_id"`
	Status string    `db:"status"`
}

func main() {
	runMigration()
	// .env から DATABASE_URL を取得
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("db connection error: %v", err)
	}
	defer db.Close()
	if err := runSeeder(db); err != nil {
		log.Fatalf("failed seed process: %v", err)
	}
}

func runMigration() {
	log.Println("running migrations...")
	// .env から DATABASE_URL を取得
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		log.Fatalf("migration initialization failed: %v", err)
	}
	defer m.Close()
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("database is already up to date.")
		} else {
			log.Fatalf("migration execution failed: %v", err)
		}
	} else {
		log.Println("migration completed successfully!")
	}
}

// トランザクションの管理
func runSeeder(db *sqlx.DB) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	// パニックなどで途中で落ちた場合、ロールバックする
	defer tx.Rollback()
	// 各テーブルの Seed 処理を順次実行
	if err := truncateAllTables(tx); err != nil {
		return err
	}
	teams, err := seedTeams(tx)
	if err != nil {
		return err
	}
	seats, err := seedSeats(tx)
	if err != nil {
		return err
	}
	games, err := seedGames(tx, teams)
	if err != nil {
		return err
	}
	if err := seedTickets(tx, games, seats); err != nil {
		return err
	}
	// 全て成功したらコミット
	return tx.Commit()
}

func truncateAllTables(tx *sqlx.Tx) error {
	_, err := tx.Exec(`TRUNCATE TABLE tickets, reservations, games, seats, teams CASCADE`)
	if err != nil {
		return fmt.Errorf("failed to truncate tables: %w", err)
	}
	return nil
}

func seedTeams(tx *sqlx.Tx) ([]Team, error) {
	teamNames := []string{
		"読売ジャイアンツ", "阪神タイガース", "東京ヤクルトスワローズ", "中日ドラゴンズ", "横浜DeNAベイスターズ", "広島東洋カープ",
		"福岡ソフトバンクホークス", "オリックス・バファローズ", "北海道日本ハムファイターズ", "千葉ロッテマリーンズ", "東北楽天ゴールデンイーグルス", "埼玉西武ライオンズ",
	}
	var teams []Team
	for _, name := range teamNames {
		team := Team{ID: uuid.New(), Name: name}
		_, err := tx.NamedExec(`INSERT INTO teams (id, name) VALUES (:id, :name)`, team)
		if err != nil {
			return nil, fmt.Errorf("failed to insert team (%s): %w", name, err)
		}
		teams = append(teams, team)
	}
	return teams, nil
}

func seedSeats(tx *sqlx.Tx) ([]Seat, error) {
	seats := []Seat{
		{ID: uuid.New(), SeatGrade: "SS", Price: 6800, TotalSeats: 10},
		{ID: uuid.New(), SeatGrade: "S", Price: 5500, TotalSeats: 20},
		{ID: uuid.New(), SeatGrade: "A", Price: 4000, TotalSeats: 30},
		{ID: uuid.New(), SeatGrade: "B", Price: 2500, TotalSeats: 40},
	}
	for _, seat := range seats {
		_, err := tx.NamedExec(`INSERT INTO seats (id, seat_grade, price, total_seats) VALUES (:id, :seat_grade, :price, :total_seats)`, seat)
		if err != nil {
			return nil, fmt.Errorf("failed to insert seat (%s): %w", seat.SeatGrade, err)
		}
	}
	return seats, nil
}

func seedGames(tx *sqlx.Tx, teams []Team) ([]Game, error) {
	loc, _ := time.LoadLocation("Asia/Tokyo")
	var games []Game
	// 基準日
	baseDate := time.Date(2025, 8, 10, 18, 0, 0, 0, loc)
	// 10日分の試合データをランダムに生成
	for i := 0; i < 10; i++ {
		// ランダムにホームとアウェイを決定
		homeIdx := rand.IntN(len(teams))
		awayIdx := rand.IntN(len(teams))
		// ホームとアウェイが同じチームにならないように再抽選
		for homeIdx == awayIdx {
			awayIdx = rand.IntN(len(teams))
		}
		// 基準日から i 日ずつずらして試合日時を設定
		gameDate := baseDate.AddDate(0, 0, i)
		// 開始時間は 13:00, 14:00, 18:00 のいずれかに設定
		startHour := []int{13, 14, 18}[rand.IntN(3)]
		startTime := time.Date(gameDate.Year(), gameDate.Month(), gameDate.Day(), startHour, 0, 0, 0, loc)
		game := Game{
			ID:         uuid.New(),
			HomeTeamID: teams[homeIdx].ID,
			AwayTeamID: teams[awayIdx].ID,
			GameDate:   time.Date(gameDate.Year(), gameDate.Month(), gameDate.Day(), 0, 0, 0, 0, loc), // 時刻を切り捨てた日付のみ
			StartTime:  startTime,
		}
		_, err := tx.NamedExec(`INSERT INTO games (id, home_team_id, away_team_id, game_date, start_time) VALUES (:id, :home_team_id, :away_team_id, :game_date, :start_time)`, game)
		if err != nil {
			return nil, fmt.Errorf("failed to insert game (day %d): %w", i, err)
		}
		games = append(games, game)
	}
	return games, nil
}

func seedTickets(tx *sqlx.Tx, games []Game, seats []Seat) error {
	var tickets []Ticket
	for _, game := range games {
		for _, seat := range seats {
			for i := 0; i < seat.TotalSeats; i++ {
				tickets = append(tickets, Ticket{
					ID:     uuid.New(),
					GameID: game.ID,
					SeatID: seat.ID,
					Status: "available",
				})
			}
		}
	}
	// tickets が空の場合はスキップ
	if len(tickets) == 0 {
		return nil
	}
	// 大量データを NamedExec でバルクインサート
	_, err := tx.NamedExec(`INSERT INTO tickets (id, game_id, seat_id, status) VALUES (:id, :game_id, :seat_id, :status)`, tickets)
	if err != nil {
		return fmt.Errorf("failed to insert tickets: %w", err)
	}
	return nil
}
