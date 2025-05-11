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

type Pagination struct {
	CurrentPage int   `json:"current_page"`
	TotalPages  int   `json:"total_pages"`
	TotalItems  int64 `json:"total_items"`
	Limit       int   `json:"limit"`
}

type ArticleResponse struct {
	Articles   []Article  `json:"articles"`
	Pagination Pagination `json:"pagination"`
}
