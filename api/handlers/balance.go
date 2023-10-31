package handlers

import (
	"encoding/json"
	"errors"
	"github.com/Tokebay/yandex-diplom/api/logger"
	"github.com/Tokebay/yandex-diplom/database"
	"github.com/Tokebay/yandex-diplom/domain/models"
	"go.uber.org/zap"

	"net/http"
)

type Balance struct {
	balanceRepo database.UserBalanceRepository
}

func NewBalance(repo database.UserBalanceRepository) *Balance {
	return &Balance{
		balanceRepo: repo,
	}
}

// GetBalanceHandler данные о текущей сумме баллов лояльности, а также сумме использованных за весь период регистрации баллов
func (h *Balance) GetBalance(w http.ResponseWriter, r *http.Request) {
	// Извлекаем идентификатор пользователя из контекста запроса
	userID, err := GetUserCookie(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем сумму баллов лояльности
	userTotalBonus, err := h.balanceRepo.GetBonusBalance(r.Context(), userID)
	if err != nil {
		logger.Log.Error("Error getting total bonuses", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	//fmt.Printf("userTotalBonus %f \n", userTotalBonus)

	// узнаем баланс списанных бонусов пользователя
	totalWithdrawn, err := h.balanceRepo.WithdrawBalance(r.Context(), userID)
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
func (h *Balance) WithdrawBalance(w http.ResponseWriter, r *http.Request) {
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
		logger.Log.Error("invalid request format", zap.Error(err))
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	logger.Log.Info("Received withdrawal request", zap.String("OrderID", wRequest.OrderID), zap.Float64("Sum", wRequest.Sum))

	// Проверка корректности номера заказа и суммы
	if wRequest.OrderID == "" || wRequest.Sum <= 0 {
		logger.Log.Error("invalid order number or withdrawal amount", zap.Error(err))
		http.Error(w, "Invalid order number or withdrawal amount", http.StatusUnprocessableEntity)
		return
	}

	// Проверка корректности номера заказа
	if !isValidLuhnAlgorithm(wRequest.OrderID) {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}
	totalWithdrawn, err := h.balanceRepo.WithdrawBalance(r.Context(), userID)
	if err != nil {
		logger.Log.Error("Error getting total Withdrawn", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	//fmt.Printf("requestSum %f; \n", wRequest.Sum)

	err = h.balanceRepo.Withdraw(r.Context(), userID, wRequest.OrderID, wRequest.Sum, totalWithdrawn)
	if err != nil {
		if errors.Is(err, database.ErrNotEnoughBalance) {
			logger.Log.Error("error not enough balance", zap.Error(err))
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}
		http.Error(w, "Failed to withdraw balance", http.StatusInternalServerError)
		return
	}

	// Успешное списание средств
	w.WriteHeader(http.StatusOK)
}

func (h *Balance) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	// Проверка авторизации пользователя
	userID, err := GetUserCookie(r)
	if err != nil {
		logger.Log.Error("Unauthorized", zap.Error(err))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получение списка выводов средств из базы данных
	withdrawals, err := h.balanceRepo.GetWithdrawals(r.Context(), userID)
	if err != nil {
		logger.Log.Error("Error getting withdrawals", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Проверка на отсутствие записей о выводе средств
	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Преобразование в JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(withdrawals); err != nil {
		logger.Log.Error("Error encoding withdrawals response", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
