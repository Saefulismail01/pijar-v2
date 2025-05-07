package model_util

type GeneratedArticle struct {
	Title   string // Judul artikel
	Content string // Isi lengkap, bisa multiline
	Source  string // Referensi atau sumber artikel
	TopicID int    // Relasi ke topik yang menghasilkan artikel
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
