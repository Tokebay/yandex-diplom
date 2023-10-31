package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Tokebay/yandex-diplom/domain/models"
)

type UserBalanceRepository interface {
	GetBonusBalance(ctx context.Context, userID int64) (float64, error)
	WithdrawBalance(ctx context.Context, userID int64) (float64, error)
	Withdraw(ctx context.Context, userID int64, orderID string, sum float64, totalWithSum float64) error
	GetWithdrawals(ctx context.Context, userID int64) ([]models.Withdraw, error)
}

func (p *PostgreStorage) GetWithdrawals(ctx context.Context, userID int64) ([]models.Withdraw, error) {
	rows, err := p.db.QueryContext(ctx, "SELECT order_id, bonuses, uploaded_at FROM withdrawals WHERE user_id = $1 ORDER BY uploaded_at ASC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []models.Withdraw
	for rows.Next() {
		var withdrawal models.Withdraw
		err := rows.Scan(&withdrawal.OrderID, &withdrawal.Sum, &withdrawal.ProcessedAt)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return withdrawals, nil
}

// GetBonusBalance общая активных баллов лояльности за весь период
func (p *PostgreStorage) GetBonusBalance(ctx context.Context, userID int64) (float64, error) {

	var userTotalBonuses sql.NullFloat64
	err := p.db.QueryRowContext(ctx, "SELECT SUM(accrual) FROM orders WHERE user_id=$1", userID).Scan(&userTotalBonuses)
	if err != nil {
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
		return 0, err
	}
	if !totalWithdrawn.Valid {
		return 0, nil
	}

	total := totalWithdrawn.Float64
	return total, nil
}

// Withdraw списывает указанное количество баллов с баланса пользователя в PostgreSQL
func (p *PostgreStorage) Withdraw(ctx context.Context, userID int64, orderID string, amount float64, totalWithSum float64) error {
	// Начинаем транзакцию
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // Откатываем транзакцию при ошибке

	// Блокируем строку пользователя с использованием FOR UPDATE
	_, err = tx.ExecContext(ctx, "SELECT id FROM users WHERE id = $1 FOR UPDATE", userID)
	if err != nil {
		return err
	}

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

		// Проверяем, не превышает ли сумма списания текущий баланс пользователя
		var newBalance float64
		err = tx.QueryRowContext(ctx, "SELECT SUM(accrual) FROM orders WHERE user_id = $1", userID).Scan(&newBalance)
		if err != nil {
			return err
		}

		fmt.Printf("Withdraw. newBalance %f \n", newBalance)
		if newBalance >= 0 {
			// Коммитим транзакцию
			err = tx.Commit()
			if err != nil {
				return err
			}
			return nil
		}
	}

	// Недостаточно баллов для списания или сумма списания превышает текущий баланс, откатываем транзакцию
	return ErrNotEnoughBalance
}
