package service

import (
	"time"

	"github.com/golang-jwt/jwt"
)

type UserRepository interface {
	CreateUser(login, password string) error
	GetUser(login, password string) (string, error)
}

type PasswordHasher interface {
	Hash(password string) (string, error)
}

type Users struct {
	repo   UserRepository
	hasher PasswordHasher
}

func NewUsers(repo UserRepository, hasher PasswordHasher, secret []byte, ttl time.Duration) *Users {
	return &Users{
		repo:   repo,
		hasher: hasher,
	}
}

func (u *Users) RegisterUser(login string, password string) error {
	// регистрации пользователя с использованием u.repo.CreateUser
	return u.repo.CreateUser(login, password)
}

func (u *Users) AuthenticateUser(login string, password string) (string, error) {
	// Реализация аутентификации пользователя с использованием u.repo.GetUser
	return u.repo.GetUser(login, password)
}

func (u *Users) GenerateToken(login string, password string, tokenTime time.Duration, secret []byte) (string, error) {
	// Проверка логина и пароля в хранилище пользователей (u.repo.GetUser)
	userID, err := u.repo.GetUser(login, password)
	if err != nil {
		return "", err
	}

	// Генерация токена
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Subject:   userID,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(tokenTime).Unix(),
	})

	// Подпись токена и возвращение его строкового представления
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
