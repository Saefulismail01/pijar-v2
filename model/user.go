package model

import (
	"time"
)

type Users struct {
	ID           int       `json:"id"` // id di DB == INT dan auto-increment
	Name         string    `json:"name"`
	Email        string    `json:"email" binding:"required,email"` // varchar + unique → string
	PasswordHash string    `json:"password"`
	BirthYear    int       `json:"birth_year"`
	Phone        string    `json:"phone"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Role         string    `json:"role"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  Users  `json:"user"`
}
