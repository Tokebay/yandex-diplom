package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Tokebay/yandex-diplom/api/logger"
	"github.com/Tokebay/yandex-diplom/database"

	"github.com/Tokebay/yandex-diplom/domain/models"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// валидация структуры запроса
var validate = validator.New()

type UserHandler struct {
	userRepository database.UserRepository
}

func NewUserHandler(userRepository database.UserRepository) *UserHandler {
	return &UserHandler{
		userRepository: userRepository,
	}
}
func (h *UserHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	user.CreatedAt = time.Now()

	// Проверка формата запроса
	if err := validate.Struct(user); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Хэшируем пароль
	user.Password = getHash([]byte(user.Password))

	// Создаем пользователя в репозитории
	login, userID, err := h.userRepository.CreateUser(user)
	if err != nil && login == "" {
		w.WriteHeader(http.StatusConflict)
		return
	}

	// Генерируем JWT токен
	token, err := BuildJWTString(userID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Устанавливаем токен в куки
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		HttpOnly: true,
		Expires:  time.Now().Add(TokenExp),
	})
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

}

func (h *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var credentials struct {
		Login    string `json:"login" validate:"required,gte=2"`
		Password string `json:"password" validate:"required,gte=4"`
	}

	// Чтение данных аутентификации из тела запроса
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		logger.Log.Error("Error decoding JSON", zap.Error(err))
		return
	}

	// Проверка формата запроса
	if err := validate.Struct(credentials); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		logger.Log.Error("Invalid request format", zap.Error(err))
		return
	}

	// Проверка логина и пароля в БД
	user, err := h.userRepository.GetUser(credentials.Login)
	if err != nil {
		if errors.Is(err, database.ErrUserNotFound) {
			logger.Log.Error("Error finding user", zap.Error(err))
		}
		http.Error(w, "Invalid login or password", http.StatusUnauthorized)
		logger.Log.Error("Invalid login or password", zap.Error(err))
		return
	}

	// Проверка пароля пользователя
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
		http.Error(w, "Invalid login or password", http.StatusUnauthorized)
		logger.Log.Error("Invalid login or password", zap.Error(err))
		return
	}

	// Получил userID из куки
	userID, err := GetUserCookie(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var token string
	if userID == -1 {
		token, err = BuildJWTString(user.ID)
		fmt.Printf("RegisterHandler userID %d", user.ID)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    token,
			HttpOnly: true,
			Expires:  time.Now().Add(time.Hour),
		})
	}

	// Отправка успешного ответа
	w.WriteHeader(http.StatusOK)
}

func getHash(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		logger.Log.Error("Error while hash password", zap.Error(err))
	}
	return string(hash)
}
