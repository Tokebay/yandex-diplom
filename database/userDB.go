package database

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Tokebay/yandex-diplom/api/logger"
	"github.com/Tokebay/yandex-diplom/domain/models"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

var ErrUserNotFound = errors.New("user not found")
var ErrAlreadyUserExist = errors.New("user already exist")

type UserRepository interface {
	CreateUser(ctx context.Context, user models.User) (int64, error)
	GetUser(ctx context.Context, login string) (*models.User, error)
}

func (p *PostgreStorage) CreateUser(ctx context.Context, user models.User) (int64, error) {
	// Проверяем существование пользователя по логину
	var existingUserID int64
	err := p.db.QueryRowContext(ctx, "SELECT id FROM users WHERE login = $1", user.Login).Scan(&existingUserID)
	if err == nil {
		// Пользователь уже существует, возвращаем ошибку конфликта
		return 0, ErrAlreadyUserExist
	}

	// Пользователь не существует, создаем нового пользователя
	var userID int64
	err = p.db.QueryRowContext(ctx, "INSERT INTO users (login, password, created_at) VALUES ($1, $2, $3) RETURNING id", user.Login, user.Password, user.CreatedAt).Scan(&userID)
	if err != nil {
		logger.Log.Error("Error creating user", zap.Error(err))
		return 0, err
	}

	return userID, nil
}

func (p *PostgreStorage) GetUser(ctx context.Context, login string) (*models.User, error) {
	var user models.User
	err := p.db.QueryRowContext(ctx, "SELECT ID,login,password FROM users WHERE login = $1", login).
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
