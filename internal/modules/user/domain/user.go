package domain

import (
	"time"
)

const UserColumns = "id, name, created_at, updated_at"

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserRegister struct {
	ID   string `json:"id"`
	Name string `json:"name" validate:"required"`
}

type UserUpdate struct {
	Name string `json:"name" validate:"required"`
}
