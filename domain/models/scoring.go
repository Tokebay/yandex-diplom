package models

type ScoringSystem struct {
	OrderID string      `json:"order"`
	Status  OrderStatus `json:"status"`
	Accrual float64     `json:"accrual"`
}
