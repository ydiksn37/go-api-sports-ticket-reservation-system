package main

import (
	"42tokyo-road-to-dena-server/config"
	"42tokyo-road-to-dena-server/handler"
	"42tokyo-road-to-dena-server/infrastructure/authbundle"
	"42tokyo-road-to-dena-server/repository"
	"42tokyo-road-to-dena-server/usecase"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	// 設定の読み込み
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// データベース接続
	db, err := sqlx.Connect("postgres", cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to database")

	// 依存関係の組み立て（Composition Root）
	// Infrastructure 層
	// authbundle の設定
	authCfg := &authbundle.AuthConfig{
		JWTSecret:    cfg.Auth.JWTSecret,
		JWTIssuer:    cfg.Auth.JWTIssuer,
		JWTAudience:  cfg.Auth.JWTAudience,
		AccessTTL:    cfg.Auth.AccessTTL,
		RefreshTTL:   cfg.Auth.RefreshTTL,
		CookieDomain: cfg.Auth.CookieDomain,
		CookieSecure: cfg.Auth.CookieSecure,
	}
	// Repository 層の実装を使用（domain.RefreshTokenRepository インターフェース経由で DI）
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	authBundle := authbundle.NewAuthBundle(authCfg, refreshTokenRepo)
	passwordHasher := authbundle.NewPasswordHasher()
	// Repository 層
	userRepo := repository.NewUserRepository(db)
	gameRepo := repository.NewGameRepository(db)
	seatRepo := repository.NewSeatRepository(db)
	reservationRepo := repository.NewReservationRepository(db)
	ticketRepo := repository.NewTicketRepository(db)
	txManager := repository.NewTxManager(db)
	// UseCase 層
	userUseCase := usecase.NewUserUseCase(userRepo, passwordHasher)
	gameUseCase := usecase.NewGameUseCase(gameRepo)
	seatUseCase := usecase.NewSeatUseCase(seatRepo, gameRepo, reservationRepo)
	reservationUseCase := usecase.NewReservationUseCase(reservationRepo, ticketRepo, gameRepo, txManager)

	// ログディレクトリとログファイルの準備
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}
	accessFile, err := os.OpenFile("logs/access.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open access.log: %v", err)
	}
	defer accessFile.Close()
	errorFile, err := os.OpenFile("logs/error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open error.log: %v", err)
	}
	defer errorFile.Close()
	accessLogger := log.New(accessFile, "", log.LstdFlags)
	errorLogger := log.New(errorFile, "", log.LstdFlags)

	// Handler 層
	h := handler.New(handler.Deps{
		AuthBundle:         authBundle,
		AuthConfig:         authCfg,
		UserUseCase:        userUseCase,
		GameUseCase:        gameUseCase,
		SeatUseCase:        seatUseCase,
		ReservationUseCase: reservationUseCase,
		AccessLogger:       accessLogger,
		ErrorLogger:        errorLogger,
	})

	// HTTP サーバーの設定
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      h.Routes(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// サーバーの起動（非同期）
	go func() {
		log.Printf("Starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// シグナルハンドリング
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// グレースフルシャットダウン
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
