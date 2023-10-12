package handlers

import (
	"encoding/json"
	"net/http"

	db "github.com/Tokebay/yandex-diplom/database"
	"github.com/Tokebay/yandex-diplom/domain/models"
)

type UserHandler struct {
	userRepository db.UserRepository
}

func NewUserHandler(userRepository db.UserRepository) *UserHandler {
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

	// Проверка на уникальность логина и сохранение пользователя в базу данных
	err = h.userRepository.CreateUser(user.Login, user.Password)
	if err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	// tokenString, err := service.GenerateToken(userID, tokenTTL, yourSecretKey)
	// if err != nil {
	// 	// Обработка ошибки
	// 	return
	// }
	// В случае успешной регистрации, автоматическая аутентификация пользователя

	// Отправка успешного ответа
	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Проверка логина и пароля в базе данных
	// fetchedUser, err := h.userRepository.GetUserByLogin(user.Login, user.Password)
	// if err != nil {
	// 	http.Error(w, "Invalid login or password", http.StatusUnauthorized)
	// 	return
	// }

	// Проверка пароля

	// В случае успешной аутентификации, генерация токена и отправка его пользователю

	// Отправка успешного ответа
	w.WriteHeader(http.StatusOK)
}
