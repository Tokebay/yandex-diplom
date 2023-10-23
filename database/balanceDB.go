package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Tokebay/yandex-diplom/api/logger"

	"go.uber.org/zap"
)

type UserBalanceRepository interface {
	GetBonusBalance(ctx context.Context, userID int64) (float64, error)
	WithdrawBalance(ctx context.Context, userID int64) (float64, error)
	Withdraw(ctx context.Context, userID int64, orderID string, sum float64) error
	CheckOrder(userID int64, orderID string) (bool, error)
}

// GetBonusBalance общая активных баллов лояльности за весь период
func (p *PostgreStorage) GetBonusBalance(ctx context.Context, userID int64) (float64, error) {

	var userTotalBonuses sql.NullFloat64
	err := p.db.QueryRowContext(ctx, "SELECT SUM(accrual) FROM orders WHERE user_id=$1", userID).Scan(&userTotalBonuses)
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
func (p *PostgreStorage) WithdrawBalance(ctx context.Context, userID int64) (float64, error) {
	var totalWithdrawn sql.NullFloat64
	err := p.db.QueryRow("SELECT SUM(bonuses) FROM withdrawals WHERE user_id=$1", userID).
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

// Withdraw списывает указанное количество баллов с баланса пользователя в PostgreSQL
func (p *PostgreStorage) Withdraw(ctx context.Context, userID int64, orderID string, amount float64) error {
	// Начинаем транзакцию
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // Откатываем транзакцию при ошибке

	// Получаем текущий баланс пользователя
	var currentBalance float64
	err = tx.QueryRowContext(ctx, "SELECT SUM(accrual) FROM orders WHERE user_id = $1", userID).Scan(&currentBalance)
	if err != nil {
		return err
	}

	// Проверяем, достаточно ли баллов для списания
	if currentBalance >= amount {
		// Выполняем списание баллов
		_, err := tx.ExecContext(ctx, "INSERT INTO withdrawals (order_id, user_id, bonuses, uploaded_at) VALUES ($1, $2, $3, NOW())",
			orderID, userID, amount)
		if err != nil {
			return err
		}

		// Обновляем баланс в таблице orders (предполагая, что у вас есть столбец balance в таблице orders)
		_, err = tx.ExecContext(ctx, "UPDATE orders SET accrual = accrual - $1 WHERE user_id = $2", amount, userID)
		if err != nil {
			return err
		}

		// Коммитим транзакцию
		err = tx.Commit()
		if err != nil {
			return err
		}
		return nil
	}

	// Недостаточно баллов для списания, откатываем транзакцию
	return ErrNotEnoughBalance
}

func (p *PostgreStorage) CheckOrder(userID int64, orderNumber string) (bool, error) {
	fmt.Println("OrderExists")
	var exists bool
	err := p.db.QueryRow("SELECT EXISTS(SELECT 1 FROM orders WHERE user_id = $1 AND order_id = $2)", userID, orderNumber).Scan(&exists)
	if err != nil {
		logger.Log.Error("Order exist", zap.Error(err))
		return false, ErrOrderExistsForUser
	}
	return exists, nil
}
