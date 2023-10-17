package main

import (
	"fmt"
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
		logger.Log.Error("Error in NewPostgreSQLStorage", zap.Error(err))
		return err
	}

	// Инициализ хэндлеров с передачей соед с бд
	userHandler := handlers.NewUserHandler(db)
	orderHandler := handlers.NewOrderHandler(db)
	r := createRouter(cfg, userHandler, orderHandler)

	// Запуск HTTP сервера
	http.ListenAndServe(cfg.RunAddress, r)

	return nil
}

func createRouter(cfg *config.Config, userHandler *handlers.UserHandler, orderHandler *handlers.OrderHandler) chi.Router {
	// Создание роутера Chi
	r := chi.NewRouter()

	// Middleware для логирования запросов
	r.Use(logger.LoggerMiddleware)

	r.Post("/api/user/register", userHandler.RegisterHandler)
	r.Post("/api/user/login", userHandler.LoginHandler)
	r.Post("/api/user/orders", orderHandler.UploadOrderHandler)
	r.Get("/api/user/orders", orderHandler.GetOrdersHandler)

	return r
}
