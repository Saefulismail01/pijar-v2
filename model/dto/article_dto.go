package dto

type ArticleDto struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Source  string `json:"source"`
	IDTopic int    `json:"id_topic"`
}

type GenerateArticleInput struct {
	Preferences []string `json:"preference" binding:"required"`
}

// dto/article_dto.go
type GenerateArticleRequest struct {
	TopicID int `json:"topic_id" binding:"required"`
}
