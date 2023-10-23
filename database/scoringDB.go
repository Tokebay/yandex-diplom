package database

import (
	"context"
	"fmt"
	"github.com/Tokebay/yandex-diplom/domain/models"
)

type ScoringRepository interface {
	GetOrderStatus(ctx context.Context) (string, error)
	UpdateOrder(ctx context.Context, order models.ScoringSystem) error
}

// GetOrderStatus получение статуса заказа
func (p *PostgreStorage) GetOrderStatus(ctx context.Context) (string, error) {
	var orderID string
	err := p.db.QueryRowContext(ctx, "SELECT order_id FROM orders WHERE status NOT IN ('PROCESSED', 'INVALID') LIMIT 1").
		Scan(&orderID)
	if err != nil {
		return "", fmt.Errorf("postgreSQL: getOrderStatus %s", err)
	}
	return orderID, nil
}

// UpdateOrder обновление статуса заказа системой внешней системой accrual
func (p *PostgreStorage) UpdateOrder(ctx context.Context, order models.ScoringSystem) error {
	_, err := p.db.ExecContext(ctx, "UPDATE orders SET status=$1, accrual=$2 WHERE order_id=$3", order.Status, order.Accrual, order.OrderID)
	if err != nil {
		return fmt.Errorf("postgreSQL: updateOrder %s", err)
	}
	return nil
}
