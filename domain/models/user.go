package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Login     string    `json:"login" validate:"required,gte=2"`
	Password  string    `json:"password" validate:"required,gte=4"`
	CreatedAt time.Time `json:"created_at"`
}
