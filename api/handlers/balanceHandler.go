package handlers

import (
	"encoding/json"
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
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем сумму баллов лояльности
	userTotalBonus, err := h.balanceRepository.GetBonusBalance(userID)
	if err != nil {
		logger.Log.Error("Error getting total bonuses", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	fmt.Printf("userTotalBonus %f \n", userTotalBonus)

	// узнаем баланс списанных бонусов пользователя
	totalWithdrawn, err := h.balanceRepository.WithdrawBalance(userID)
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

func (h *BalanceHandler) WithdrawBalanceHandler(w http.ResponseWriter, r *http.Request) {
	// Проверка авторизации пользователя
	_, err := GetUserCookie(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Чтение данных запроса
	var withdrawRequest struct {
		Order string `json:"order"`
		Sum   int    `json:"sum"`
	}

	// Декодирование JSON-данных запроса
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&withdrawRequest); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Проверка корректности номера заказа
	if !isValidLuhnAlgorithm(withdrawRequest.Order) {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	// Начало транзакции с пессимистической блокировкой
	//tx, err := h.db.Begin()
	//if err != nil {
	//	http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
	//	return
	//}
	//defer tx.Rollback() // Откатываем транзакцию в случае ошибки

	// Получение текущего баланса пользователя с пессимистической блокировкой
	//userBalance, err := h.balanceRepository.GetUserBalanceWithLock(tx, userID)
	//if err != nil {
	//	http.Error(w, "Failed to get user balance", http.StatusInternalServerError)
	//	return
	//}
	//
	//// Проверка наличия достаточного количества баллов на счете пользователя
	//if userBalance < withdrawRequest.Sum {
	//	http.Error(w, "Insufficient funds", http.StatusPaymentRequired)
	//	return
	//}

	//	err = h.balanceRepository.WithdrawBalance(tx, userID, withdrawRequest.Sum)
	//	if err != nil {
	//		http.Error(w, "Failed to withdraw balance", http.StatusInternalServerError)
	//		return
	//	} Проведение операции списания

	// Фиксация транзакции
	//err = tx.Commit()
	//if err != nil {
	//	http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
	//	return
	//}

	// Успешное списание средств
	w.WriteHeader(http.StatusOK)
}
