package database

import (
	"database/sql"
	"errors"

	"github.com/Tokebay/yandex-diplom/api/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	goose "github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	CreateUser(login, password string) error
	GetUser(login, password string) (string, error)
}

type PostgreStorage struct {
	db *sql.DB
}

// закрываем подключение к базе данных.
func (p *PostgreStorage) Close() error {
	return p.db.Close()
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

func (p *PostgreStorage) CreateUser(login, password string) error {
	res, err := p.db.Exec("INSERT INTO users (login, password, registered_at) values ($1, $2, $3) on conflict (login) do nothing", login, password)
	if err != nil {
		// logger.Log.Error("Error creating user", zap.Error(err))
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return err
	}

	return nil
}

func (p *PostgreStorage) GetUser(login, password string) (string, error) {
	var user string
	err := p.db.QueryRow("SELECT login FROM users WHERE login = $1 AND password = $2", login, password).Scan(&user)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrUserNotFound
		}
		// logger.Log.Error("Error retrieving user", zap.Error(err))
		return "", err
	}
	return user, nil
}
