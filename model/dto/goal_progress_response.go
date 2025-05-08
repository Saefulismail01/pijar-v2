package dto

import "pijar/model"

type ArticleProgress struct {
	ArticleID     int64  `json:"article_id"`
	Completed     bool   `json:"completed"`
	DateCompleted string `json:"date_completed,omitempty"`
}

type GoalProgressResponse struct {
	ID             int               `json:"id"`
	Title          string            `json:"title"`
	Task           string            `json:"task"`
	Articles       []ArticleProgress `json:"articles"`
	Completed      bool              `json:"completed"`
	CreatedAt      string            `json:"created_at"`
	TotalCompleted int               `json:"total_completed"`
	TotalArticles  int               `json:"total_articles"`
}

type GoalProgressInfo struct {
	Goal     model.UserGoal
	Progress []ArticleProgress
}
