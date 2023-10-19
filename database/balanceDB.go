package database

import (
	"database/sql"
	"github.com/Tokebay/yandex-diplom/api/logger"
	"go.uber.org/zap"
)

type UserBalanceRepository interface {
	GetBonusBalance(userID int) (float64, error)
	WithdrawBalance(userID int) (float64, error)
}

// GetBonusBalance общая активных баллов лояльности за весь период
func (p *PostgreStorage) GetBonusBalance(userID int) (float64, error) {

	var userTotalBonuses sql.NullFloat64
	err := p.db.QueryRow("SELECT SUM(accrual) AS total_accrual FROM orders WHERE user_id=$1", userID).Scan(&userTotalBonuses)
	if err != nil {
		logger.Log.Error("Error get user total accrual ", zap.Error(err))
		return 0, err
	}

	// Пользователь не имеет начисленных баллов, возвращаем 0
	if !userTotalBonuses.Valid {
		return 0, nil
	}

	balance := userTotalBonuses.Float64
	return balance, nil
}

// WithdrawBalance общая сумма использованных баллов лояльности за весь период
func (p *PostgreStorage) WithdrawBalance(userID int) (float64, error) {
	var totalWithdrawn sql.NullFloat64
	err := p.db.QueryRow("SELECT SUM(accrual) AS total_withdrawn FROM withdrawals WHERE user_id=$1", userID).
		Scan(&totalWithdrawn)
	if err != nil {
		logger.Log.Error("Error get total withdrawn sum", zap.Error(err))
		return 0, err
	}
	if !totalWithdrawn.Valid {
		return 0, nil
	}

	total := totalWithdrawn.Float64
	return total, nil
}
