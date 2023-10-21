package models

import "time"

type contextKey string

const UserIDKey contextKey = "userID"

type User struct {
	ID        int64     `json:"id"`
	Login     string    `json:"login" validate:"required,gte=2"`
	Password  string    `json:"password" validate:"required,gte=4"`
	CreatedAt time.Time `json:"created_at"`
}
