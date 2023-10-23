package main

import (
	"context"
	"fmt"
	"github.com/Tokebay/yandex-diplom/api/accrual"
	"github.com/Tokebay/yandex-diplom/api/middleware"
	"net/http"

	"github.com/Tokebay/yandex-diplom/api/handlers"
	"github.com/Tokebay/yandex-diplom/api/logger"
	"github.com/Tokebay/yandex-diplom/config"
	"github.com/Tokebay/yandex-diplom/database"
	"github.com/go-chi/chi/v5"

	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("Error", err)
	}

}

type App struct {
	UserHandler    *handlers.UserHandler
	OrderHandler   *handlers.OrderHandler
	BalanceHandler *handlers.BalanceHandler
	ScoringHandler *handlers.ScoringSystemHandler
}

func run() error {
	//Инициализируется логгер
	err := logger.Initialize("info")
	if err != nil {
		logger.Log.Error("Error init logger", zap.Error(err))
	}

	cfg := config.NewConfig()

	// Подключение к базе данных
	db, err := database.NewPostgreSQL(cfg.DatabaseURI)
	if err != nil {
		logger.Log.Error("Error with connection to DB", zap.Error(err))
		return err
	}

	// Инициализ хэндлеров с передачей соед с бд
	userHandler := handlers.NewUserHandler(db)
	orderHandler := handlers.NewOrderHandler(db)
	balanceHandler := handlers.NewBalanceHandler(db)
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

	// Запуск HTTP сервера с контекстом
	go func() {
		err := http.ListenAndServe(cfg.RunAddress, r)
		if err != nil {
			logger.Log.Fatal("Error starting HTTP server", zap.Error(err))
			cancel()
		}
	}()

	// Запуск системы начисления баллов в фоне
	go func() {
		apiAccrualSystem := &accrual.APIAccrualSystem{
			ScoringSystemHandler: scoringHandler,
			Config:               cfg,
		}
		apiAccrualSystem.ScoringSystem()
	}()

	<-ctx.Done()

	return nil
}

func createRouter(app *App) chi.Router {
	// Создание роутера Chi
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		// Middleware для логирования запросов
		r.Use(logger.LoggerMiddleware)
		r.Post("/api/user/register", app.UserHandler.RegisterHandler)
		r.Post("/api/user/login", app.UserHandler.LoginHandler)

		r.With(middleware.AuthMiddleware).Post("/api/user/orders", app.OrderHandler.UploadOrderHandler)
		r.With(middleware.AuthMiddleware).Get("/api/user/orders", app.OrderHandler.GetOrdersHandler)
		r.With(middleware.AuthMiddleware).Get("/api/user/balance", app.BalanceHandler.GetBalanceHandler)
		r.With(middleware.AuthMiddleware).Post("/api/user/balance/withdraw", app.BalanceHandler.WithdrawBalanceHandler)

	})
	return r
}
