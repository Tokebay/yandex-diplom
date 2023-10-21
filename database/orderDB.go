package database

import (
	"errors"
	"fmt"
	"github.com/Tokebay/yandex-diplom/api/logger"
	"github.com/Tokebay/yandex-diplom/domain/models"
	"go.uber.org/zap"
)

var ErrOrderExistsForUser = errors.New("order already exists for the user")
var ErrOrderExists = errors.New("order already exists")
var ErrDataNotFound = errors.New("data not found")

type OrderRepository interface {
	OrderExists(userID int64, orderID string) (bool, error)
	OrderExistsByNumber(orderID string) (bool, error)
	UploadOrder(order models.Order) error
	GetOrdersByUserID(userID int64) ([]models.OrderResponse, error)
}

func (p *PostgreStorage) OrderExists(userID int64, orderNumber string) (bool, error) {
	fmt.Println("OrderExists")
	var exists bool
	err := p.db.QueryRow("SELECT EXISTS(SELECT 1 FROM orders WHERE user_id = $1 AND order_id = $2)", userID, orderNumber).Scan(&exists)
	if err != nil {
		logger.Log.Error("Order exist", zap.Error(err))
		return false, ErrOrderExistsForUser
	}
	return exists, nil
}

func (p *PostgreStorage) OrderExistsByNumber(orderNumber string) (bool, error) {
	fmt.Println("OrderExistsByNumber")
	var exists bool
	err := p.db.QueryRow("SELECT EXISTS(SELECT 1 FROM orders WHERE order_id = $1)", orderNumber).Scan(&exists)
	if err != nil {
		logger.Log.Error("Order exist By number", zap.Error(err))
		return false, ErrOrderExists
	}
	return exists, nil
}

func (p *PostgreStorage) UploadOrder(order models.Order) error {

	// fmt.Printf("%d %s %s %s \n", order.UserID, order.Number, order.Status, order.UploadedAt)
	// Insert нового заказа
	_, err := p.db.Exec("INSERT INTO orders (user_id, order_id, status, uploaded_at, accrual) VALUES ($1, $2, $3, $4, $5)",
		order.UserID, order.Number, order.Status, order.UploadedAt, order.Accrual)
	if err != nil {
		logger.Log.Error("Error inserting order into orders table", zap.Error(err))
		return err
	}
	return nil
}

func (p *PostgreStorage) GetUserBalanceWithLock(userID int) (float64, error) {
	var balance float64
	err := p.db.QueryRow("SELECT balance FROM users WHERE id = $1 FOR UPDATE", userID).Scan(&balance)
	if err != nil {
		logger.Log.Error("Error getting user balance with lock", zap.Error(err))
		return 0, err
	}
	return balance, nil
}

func (p *PostgreStorage) GetOrdersByUserID(userID int64) ([]models.OrderResponse, error) {

	rows, err := p.db.Query("SELECT order_id, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []models.OrderResponse{}
	for rows.Next() {
		var order models.OrderResponse

		err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			logger.Log.Error("Error scan orders table", zap.Error(err))
			return nil, err
		}

		orders = append(orders, order)
	}
	err = rows.Err()
	if err != nil {
		logger.Log.Error("Error get orders", zap.Error(err))
		return nil, err
	}

	if len(orders) == 0 {
		return nil, ErrDataNotFound
	}

	return orders, nil
}
