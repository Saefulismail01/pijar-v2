package dto

type CreateGoalRequest struct {
	Title          string  `json:"title" binding:"required" example:"Learn Golang"`
	Task           string  `json:"task" binding:"required" example:"Study Go basics"`
	ArticlesToRead []int64 `json:"articles_to_read,omitempty" example:"1,2,3"`
}

type GoalResponse struct {
	ID             int     `json:"id" example:"1"`
	Title          string  `json:"title" example:"Learn Golang"`
	Task           string  `json:"task" example:"Study Go basics"`
	ArticlesToRead []int64 `json:"articles_to_read" example:"1,2,3"`
	Completed      bool    `json:"completed" example:"false"`
	CreatedAt      string  `json:"created_at" example:"2023-08-15 14:30:00"`
}

type UpdateGoalRequest struct {
	Title          string  `json:"title" binding:"required" example:"Advanced Golang"`
	Task           string  `json:"task" binding:"required" example:"Study concurrency"`
	Completed      bool    `json:"completed" example:"false"`
	ArticlesToRead []int64 `json:"articles_to_read,omitempty" example:"4,5,6"`
}

type CompleteArticleRequest struct {
	UserID    int `json:"user_id" binding:"required" example:"1"`
	GoalID    int `json:"goal_id" binding:"required" example:"1"`
	ArticleID int `json:"article_id" binding:"required" example:"1"`
}
