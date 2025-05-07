package model

import "time"

type Journal struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Judul     string    `json:"judul"`
	Isi       string    `json:"isi"`
	Perasaan  string    `json:"perasaan"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}