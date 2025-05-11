package dto

import "pijar/model"

type ArticleProgress struct {
	ArticleID     int64  `json:"article_id" example:"1"`
	Completed     bool   `json:"completed" example:"true"`
	DateCompleted string `json:"date_completed,omitempty" example:"2023-08-15 14:30:00"`
}

type GoalProgressResponse struct {
	ID             int               `json:"id" example:"1"`
	Title          string            `json:"title" example:"Learn Golang"`
	Task           string            `json:"task" example:"Study Go basics"`
	Articles       []ArticleProgress `json:"articles"`
	Completed      bool              `json:"completed" example:"false"`
	CreatedAt      string            `json:"created_at" example:"2023-08-15 14:30:00"`
	TotalCompleted int               `json:"total_completed" example:"2"`
	TotalArticles  int               `json:"total_articles" example:"3"`
}

type GoalProgressInfo struct {
	Goal     model.UserGoal
	Progress []ArticleProgress
}
