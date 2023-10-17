package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/Tokebay/yandex-diplom/api/logger"
	"github.com/Tokebay/yandex-diplom/domain/models"
	"github.com/jackc/pgx"
	_ "github.com/jackc/pgx/v5/stdlib"
	goose "github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

var ErrUserNotFound = errors.New("user not found")
var ErrAlreadyUserExist = errors.New("user already exist")

var ErrOrderExistsForUser = errors.New("order already exists for the user")
var ErrOrderExists = errors.New("order already exists")
var ErrDataNotFound = errors.New("data not found")

type UserRepository interface {
	CreateUser(user models.User) (string, int, error)
	GetUser(login string) (*models.User, error)
}

type OrderRepository interface {
	OrderExists(userID int, orderID string) (bool, error)
	OrderExistsByNumber(orderID string) (bool, error)
	UploadOrder(order models.Order) error
	GetOrdersByUserID(useID int) ([]models.OrderResponse, error)
}

type PostgreStorage struct {
	db *sql.DB
}

// закрываем подключение к базе данных.
func (p *PostgreStorage) Close() error {
	return p.db.Close()
}

func NewPostgreSQL(dsn string) (*PostgreStorage, error) {
	// миграции
	db, err := goose.OpenDBWithDriver("pgx", dsn)
	if err != nil {
		logger.Log.Error("Error open conn", zap.Error(err))
		return nil, err
	}
	err = goose.Up(db, "./migrations")
	if err != nil {
		logger.Log.Error("Error goose UP", zap.Error(err))
		return nil, err
	}

	return &PostgreStorage{db: db}, nil
}

func (p *PostgreStorage) CreateUser(user models.User) (string, int, error) {

	var (
		login  string
		userID int
		err    = p.db.QueryRow("INSERT INTO users (login, password, created_at) values ($1, $2, $3) on conflict (login) do nothing RETURNING login, id", user.Login, user.Password, user.CreatedAt).Scan(&login, &userID)
	)
	if err != nil {
		logger.Log.Error("Error creating user", zap.Error(err))
		if err == pgx.ErrNoRows { // если ON CONFLICT не сработал и ни одна строка не вернулась
			fmt.Println("rowsAffected 0")
			return "", 0, ErrAlreadyUserExist
		}
		return "", 0, err
	}

	return login, userID, nil
}

func (p *PostgreStorage) GetUser(login string) (*models.User, error) {
	var user models.User
	err := p.db.QueryRow("SELECT ID,login,password FROM users WHERE login = $1", login).
		Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		logger.Log.Error("Error retrieving user", zap.Error(err))
		return nil, err
	}
	return &user, nil
}

func (p *PostgreStorage) OrderExists(userID int, orderNumber string) (bool, error) {
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
	fmt.Println("UploadOrder")
	//fmt.Printf("%d %s %s %s \n", order.UserID, order.Number, order.Status, order.CreatedAt)
	// Insert нового заказа
	_, err := p.db.Exec("INSERT INTO orders (user_id, order_id, status, uploaded_at) VALUES ($1, $2, $3, $4)",
		order.UserID, order.Number, order.Status, order.CreatedAt)
	if err != nil {
		logger.Log.Error("Error inserting order into orders table", zap.Error(err))
		return err
	}
	return nil
}

func (p *PostgreStorage) GetOrdersByUserID(userID int) ([]models.OrderResponse, error) {

	rows, err := p.db.Query("SELECT order_id, status, accrual, uploaded_at FROM orders WHERE user_id = $1", userID)
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
