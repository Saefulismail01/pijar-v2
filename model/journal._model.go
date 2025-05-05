package model

type Journal struct {
	ID       int    `json:"id"`
	UserID   int    `json:"user_id"`
	Judul    string `json:"judul"`
	Isi      string `json:"isi"`
	Perasaan string `json:"perasaan"`
}