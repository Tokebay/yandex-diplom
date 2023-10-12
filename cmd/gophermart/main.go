package main

import (
	"net/http"

	"github.com/Tokebay/yandex-diplom/api/handlers"
	"github.com/Tokebay/yandex-diplom/config"
	"github.com/Tokebay/yandex-diplom/database"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.NewConfig()
	config.ParseFlags(cfg)
	// Подключение к базе данных

	db, err := database.NewPostgreSQL(cfg.DatabaseURI)
	if err != nil {
		// ошибка подключения к базе данных
		// panic(err)
	}

	// Создание роутера Chi
	r := chi.NewRouter()

	// Инициализ хэндлеров с передачей соед с бд
	userHandler := handlers.NewUserHandler(db)

	// Middleware для логирования запросов
	// r.Use(middleware.Logger)

	r.Post("/api/user/register", userHandler.RegisterHandler)
	r.Post("/api/user/login", userHandler.LoginHandler)

	// Запуск HTTP сервера
	http.ListenAndServe(cfg.RunAddress, r)
}
