package model_util

type GeneratedArticle struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Source  string `json:"source"`
	TopicID int    `json:"topic_id"`
}

type DeepseekChatRequest struct {
	Model    string                `json:"model"`
	Messages []DeepseekChatMessage `json:"messages"`
}

type DeepseekChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type DeepseekChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error map[string]interface{} `json:"error"`
}
