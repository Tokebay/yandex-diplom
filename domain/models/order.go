package models

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type OrderStatus string

type Order struct {
	Number    string      `json:"number"`
	Status    OrderStatus `json:"status"`
	Accrual   float32     `json:"accrual"`
	CreatedAt string      `json:"created_at"`
	UserID    int         `json:"-"`
}
