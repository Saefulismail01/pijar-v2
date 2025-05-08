package model

import (
	"time"
)

// model/article.go
type Article struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Source    string    `json:"source"`
	IDTopic   int       `json:"id_topic"`
	CreatedAt time.Time `json:"created_at"`
}
