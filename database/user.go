package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/Tokebay/yandex-diplom/api/logger"
	"github.com/Tokebay/yandex-diplom/domain/models"
	"github.com/jackc/pgx"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

var ErrUserNotFound = errors.New("user not found")
var ErrAlreadyUserExist = errors.New("user already exist")

type UserRepository interface {
	CreateUser(user models.User) (string, int64, error)
	GetUser(login string) (*models.User, error)
}

func (p *PostgreStorage) CreateUser(user models.User) (string, int64, error) {

	var (
		login  string
		userID int64
		err    = p.db.QueryRow("INSERT INTO users (login, password, created_at) values ($1, $2, $3) on conflict (login) do nothing RETURNING login, id", user.Login, user.Password, user.CreatedAt).Scan(&login, &userID)
	)
	if err != nil {
		logger.Log.Error("Error creating user", zap.Error(err))
		if err == pgx.ErrNoRows { // если ON CONFLICT не сработал и ни одна строка не вернулась
			fmt.Println("rowsAffected 0")
			return "", 0, ErrAlreadyUserExist
		}
		return "", 0, err
	}

	return login, userID, nil
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
