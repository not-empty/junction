package user

import "time"

type User struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserRegister struct {
	Name string `json:"name" validate:"required"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewUserResponse(user User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
