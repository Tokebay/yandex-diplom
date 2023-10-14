package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Tokebay/yandex-diplom/api/logger"
	"github.com/Tokebay/yandex-diplom/domain/models"
	"github.com/jackc/pgx"
	_ "github.com/jackc/pgx/v5/stdlib"
	goose "github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

var ErrUserNotFound = errors.New("user not found")
var ErrAlreadyUserExist = errors.New("user already exist")

type UserRepository interface {
	CreateUser(user models.User) (string, error)
	GetUser(login string) (*models.User, error)
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

func (p *PostgreStorage) CreateUser(user models.User) (string, error) {

	var login string
	err := p.db.QueryRow("INSERT INTO users (login, password, created_at) values ($1, $2, $3) on conflict (login) do nothing RETURNING login", user.Login, user.Password, user.CreatedAt).Scan(&login)
	if err != nil {
		logger.Log.Error("Error creating user", zap.Error(err))
		return "", err
	}

	if err != nil {
		if err == pgx.ErrNoRows { // если ON CONFLICT не сработал и ни одна строка не вернулась
			fmt.Println("rowsAffected 0")
			return "", ErrAlreadyUserExist
		}
		logger.Log.Error("Error insert URL to table", zap.Error(err))
		return "", err
	}

	return login, nil
}

func (p *PostgreStorage) GetUser(login string) (*models.User, error) {
	var user models.User
	err := p.db.QueryRow("SELECT ID,login,password FROM users WHERE login = $1", login).
		Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		logger.Log.Error("Error retrieving user", zap.Error(err))
		return nil, err
	}
	return &user, nil
}
