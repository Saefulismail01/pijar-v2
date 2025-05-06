package dto

import (
	"pijar/model"
)

type Register struct {
	Name  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	BirthYear int    `json:"birth_year" binding:"required,gte=1900,lte=2025"`
	Phone     string `json:"phone" binding:"required"`
}

type Login struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserCreationResponse struct {
	User  model.Users`json:"user"`
	Error string     `json:"error,omitempty"`
}