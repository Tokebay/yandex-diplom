package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Tokebay/yandex-diplom/api/logger"
	"github.com/Tokebay/yandex-diplom/database"
	"github.com/Tokebay/yandex-diplom/domain/models"
	"go.uber.org/zap"

	"net/http"
)

type BalanceHandler struct {
	balanceRepository database.UserBalanceRepository
}

func NewBalanceHandler(balanceRepository database.UserBalanceRepository) *BalanceHandler {
	return &BalanceHandler{
		balanceRepository: balanceRepository,
	}
}

// GetBalanceHandler данные о текущей сумме баллов лояльности, а также сумме использованных за весь период регистрации баллов
func (h *BalanceHandler) GetBalanceHandler(w http.ResponseWriter, r *http.Request) {
	// Извлекаем идентификатор пользователя из контекста запроса
	userID, err := GetUserCookie(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем сумму баллов лояльности
	userTotalBonus, err := h.balanceRepository.GetBonusBalance(r.Context(), userID)
	if err != nil {
		logger.Log.Error("Error getting total bonuses", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	fmt.Printf("userTotalBonus %f \n", userTotalBonus)

	// узнаем баланс списанных бонусов пользователя
	totalWithdrawn, err := h.balanceRepository.WithdrawBalance(r.Context(), userID)
	if err != nil {
		logger.Log.Error("Error getting total Withdrawn", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Вычитаем из общей суммы начисленных бонусов количество использованных бонусов
	balanceAfterWithdrawn := userTotalBonus - totalWithdrawn

	// Создаем ответ с балансами
	response := models.BalanceResponse{
		Current:   balanceAfterWithdrawn,
		Withdrawn: totalWithdrawn,
	}

	// Преобразуем ответ в JSON и отправляем клиенту
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Log.Error("Error encoding JSON", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// WithdrawBalanceHandler Запрос на списание средств
func (h *BalanceHandler) WithdrawBalanceHandler(w http.ResponseWriter, r *http.Request) {
	// Проверка авторизации пользователя
	userID, err := GetUserCookie(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Чтение данных запроса
	var wRequest models.WithdrawRequest

	// Декодирование JSON-данных запроса
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&wRequest); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Проверка корректности номера заказа и суммы
	if wRequest.OrderID == "" || wRequest.Sum <= 0 {
		http.Error(w, "Invalid order number or withdrawal amount", http.StatusUnprocessableEntity)
		return
	}

	// Проверка корректности номера заказа
	if !isValidLuhnAlgorithm(wRequest.OrderID) {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	// проверка номера заказа
	isOrderExist, err := h.balanceRepository.CheckOrder(userID, string(wRequest.OrderID))
	if err != nil {
		logger.Log.Error("Error order exist", zap.Error(err))
	}
	if !isOrderExist {
		http.Error(w, "Order was uploaded by another user", http.StatusUnprocessableEntity)
		return
	}

	err = h.balanceRepository.Withdraw(r.Context(), userID, wRequest.OrderID, wRequest.Sum)
	if err != nil {
		if errors.Is(err, database.ErrNotEnoughBalance) {
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}
		http.Error(w, "Failed to withdraw balance", http.StatusInternalServerError)
		return
	}

	// Успешное списание средств
	w.WriteHeader(http.StatusOK)
}
