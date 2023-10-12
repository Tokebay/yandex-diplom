package models

import "time"

type User struct {
	ID       int       `json:"id"`
	Login    string    `json:"login"`
	Password string    `json:"password"`
	Created  time.Time `json:"created"`
}
