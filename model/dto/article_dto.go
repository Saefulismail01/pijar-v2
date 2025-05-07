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

type ArticleSearchRequest struct {
	Title string `json:"title" binding:"required"`
}

type ArticleSearchResponse struct {
	Found        bool            `json:"found"`
	Article      interface{}     `json:"article,omitempty"`
	Suggestions  []string        `json:"suggestions,omitempty"`
	Message      string          `json:"message"`
}