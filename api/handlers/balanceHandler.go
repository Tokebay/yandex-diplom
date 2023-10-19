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
	userID, ok := r.Context().Value("userID").(int)
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
