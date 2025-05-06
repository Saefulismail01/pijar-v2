package dto

type CreateGoalRequest struct {
	Title          string  `json:"title" binding:"required"`
	Task           string  `json:"task" binding:"required"`
	ArticlesToRead []int64 `json:"articles_to_read"`
}

type GoalResponse struct {
	ID             int     `json:"id"`
	Title          string  `json:"title"`
	Task           string  `json:"task"`
	ArticlesToRead []int64 `json:"articles_to_read"`
	Completed      bool    `json:"completed"`
	CreatedAt      string  `json:"created_at"`
}

type UpdateGoalRequest struct {
	Title          string  `json:"title" binding:"required"`
	Task           string  `json:"task" binding:"required"`
	Completed      bool    `json:"completed"`
	ArticlesToRead []int64 `json:"articles_to_read"`
}

type CompleteArticleRequest struct {
	UserID    int `json:"user_id" binding:"required"`
	GoalID    int `json:"goal_id" binding:"required"`
	ArticleID int `json:"article_id" binding:"required"`
}
