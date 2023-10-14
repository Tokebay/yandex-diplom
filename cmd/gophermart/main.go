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
	logger.Initialize("info")

	cfg := config.NewConfig()

	// Подключение к базе данных
	db, err := database.NewPostgreSQL(cfg.DatabaseURI)
	if err != nil {
		logger.Log.Error("Error in NewPostgreSQLStorage", zap.Error(err))
		return err
	}

	// Инициализ хэндлеров с передачей соед с бд
	userHandler := handlers.NewUserHandler(db)
	r := createRouter(cfg, userHandler)

	// Запуск HTTP сервера
	http.ListenAndServe(cfg.RunAddress, r)

	return nil
}

func createRouter(cfg *config.Config, userHandler *handlers.UserHandler) chi.Router {
	// Создание роутера Chi
	r := chi.NewRouter()

	// Middleware для логирования запросов
	r.Use(logger.LoggerMiddleware)

	r.Post("/api/user/register", userHandler.RegisterHandler)
	r.Post("/api/user/login", userHandler.LoginHandler)

	return r
}
