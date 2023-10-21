package database

import (
	"database/sql"
	"github.com/Tokebay/yandex-diplom/api/logger"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

type PostgreStorage struct {
	db *sql.DB
}

// закрываем подключение к базе данных.
func (p *PostgreStorage) Close() error {
	return p.db.Close()
}

func (p *PostgreStorage) Begin() (*sql.Tx, error) {
	tx, err := p.db.Begin()
	if err != nil {
		logger.Log.Error("Error starting transaction", zap.Error(err))
		return nil, err
	}
	return tx, nil
}

func NewPostgreSQL(dsn string) (*PostgreStorage, error) {
	// миграции
	db, err := goose.OpenDBWithDriver("pgx", dsn)
	if err != nil {
		logger.Log.Error("Error open conn", zap.Error(err))
		return nil, err
	}
	err = goose.Up(db, "./migrations")
	if err != nil {
		logger.Log.Error("Error goose UP", zap.Error(err))
		return nil, err
	}

	return &PostgreStorage{db: db}, nil
}
