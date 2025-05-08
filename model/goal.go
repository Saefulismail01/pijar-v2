package model

import "time"

type UserGoal struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	Title          string    `json:"title"`
	Task           string    `json:"task"`
	ArticlesToRead []int64   `json:"articles_to_read"`
	Completed      bool      `json:"completed"`
	CreatedAt      time.Time `json:"created_at"`
}

type GoalProgress struct {
	ID           int       `json:"id"`
	GoalID       int       `json:"goal_id"`
	ArticleID    int       `json:"article_id"`
	DateAssigned time.Time `json:"date_assigned"`
	Completed    bool      `json:"completed"`
}
