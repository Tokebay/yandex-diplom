package main

import (
	"context"
	"fmt"
	"github.com/Tokebay/yandex-diplom/api/accrual"
	"github.com/Tokebay/yandex-diplom/api/handlers"
	"github.com/Tokebay/yandex-diplom/api/logger"
	"github.com/Tokebay/yandex-diplom/api/middleware"
	"github.com/Tokebay/yandex-diplom/config"
	"github.com/Tokebay/yandex-diplom/database"
	"github.com/go-chi/chi/v5"
	mdlw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("Error", err)
	}

}

type App struct {
	UserHandler    *handlers.User
	OrderHandler   *handlers.Order
	BalanceHandler *handlers.Balance
	ScoringHandler *handlers.ScoringSystem
}

func run() error {
	//Инициализируется логгер
	err := logger.Initialize("info")
	if err != nil {
		logger.Log.Error("Error init logger", zap.Error(err))
	}

	cfg := config.NewConfig()
	logger.Log.Info("Server configuration:", zap.String("RunAddress", cfg.RunAddress), zap.String("DatabaseURI", cfg.DatabaseURI), zap.String("AccrualSystemAddr", cfg.AccrualSystemAddr))

	// Подключение к базе данных
	db, err := database.NewPostgreSQL(cfg.DatabaseURI)
	if err != nil {
		logger.Log.Error("Error with connection to DB", zap.Error(err))
		return err
	}

	// Инициализ хэндлеров с передачей соед с бд
	userHandler := handlers.NewUser(db)
	orderHandler := handlers.NewOrder(db)
	balanceHandler := handlers.NewBalance(db)
	scoringHandler := handlers.NewScoringSystem(db)

	app := &App{
		UserHandler:    userHandler,
		OrderHandler:   orderHandler,
		BalanceHandler: balanceHandler,
		ScoringHandler: scoringHandler,
	}

	r := createRouter(app)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Запускаем функцию ScoringSystem
	StartScoringSystem(ctx, app, cfg)

	// Запуск HTTP сервера с контекстом
	go func() {
		err := http.ListenAndServe(cfg.RunAddress, r)
		if err != nil {
			logger.Log.Fatal("Error starting HTTP server", zap.Error(err))
			cancel()
		}
	}()

	<-ctx.Done()

	return nil
}

func createRouter(app *App) chi.Router {
	// Создание роутера Chi
	r := chi.NewRouter()
	r.Use(logger.LoggerMiddleware, mdlw.Recoverer)

	r.Group(func(r chi.Router) {
		// Middleware для логирования запросов

		r.Post("/api/user/register", app.UserHandler.Register)
		r.Post("/api/user/login", app.UserHandler.Login)

		r.With(middleware.AuthMiddleware).Post("/api/user/orders", app.OrderHandler.UploadOrder)
		r.With(middleware.AuthMiddleware).Get("/api/user/orders", app.OrderHandler.GetOrders)
		r.With(middleware.AuthMiddleware).Get("/api/user/balance", app.BalanceHandler.GetBalance)
		r.With(middleware.AuthMiddleware).Post("/api/user/balance/withdraw", app.BalanceHandler.WithdrawBalance)
		r.With(middleware.AuthMiddleware).Get("/api/user/withdrawals", app.BalanceHandler.GetWithdrawals)

	})
	return r
}

func StartScoringSystem(ctx context.Context, app *App, cfg *config.Config) {

	ticker := time.NewTicker(time.Millisecond * 100)
	go func() {
		for {
			select {
			case <-ticker.C:
				apiAccrualSystem := &accrual.APIAccrualSystem{
					ScoringSystemHandler: app.ScoringHandler,
					Config:               cfg,
				}
				apiAccrualSystem.ScoringSystem()
			case <-ctx.Done():
				return
			}
		}
	}()
}
