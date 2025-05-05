package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"pijar/utils/model_util"
)

type deepseekChatRequest struct {
	Model    string                `json:"model"`
	Messages []deepseekChatMessage `json:"messages"`
}

type deepseekChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type deepseekChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error map[string]interface{} `json:"error"` // jaga-jaga kalau API kirim error
}

// GenerateArticleFromDeepseek mengirim prompt dan mengembalikan hasil sebagai struct article
func GenerateArticleFromDeepseek(topic string, preference string, topicID int) (*model_util.GeneratedArticle, error) {
	prompt := fmt.Sprintf(`Buat artikel tentang %s dengan format ketat:

1. **Judul:** [1 judul informatif]
2. **Isi:** 
   - [minimal 300 kata, masing-masing 1 paragraf pendek]
   - Setiap paragraf dipisahkan oleh baris baru
3. **Sumber:** [1 referensi/sumber]
4. **Preferensi:** [%s]`, topic, preference)

	reqBody := deepseekChatRequest{
		Model: "deepseek-chat",
		Messages: []deepseekChatMessage{
			{Role: "user", Content: prompt},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("gagal encode JSON: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.deepseek.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("gagal buat request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("DEEPSEEK_API_KEY"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal kirim request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal baca response: %w", err)
	}

	var result deepseekChatResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		// Tampilkan isi respon mentah jika gagal decode
		fmt.Println("Respon mentah:\n", string(respBody))
		return nil, fmt.Errorf("gagal decode response JSON: %w", err)
	}

	// Validasi apakah respons berisi artikel
	if len(result.Choices) == 0 || result.Choices[0].Message.Content == "" {
		fmt.Println("Respon mentah:\n", string(respBody)) // debug isi JSON
		return nil, fmt.Errorf("tidak ada hasil dari Deepseek")
	}

	content := result.Choices[0].Message.Content
	article := parseGeneratedArticle(content, topicID)

	return article, nil
}

func parseGeneratedArticle(raw string, topicID int) *model_util.GeneratedArticle {
	lines := strings.Split(strings.TrimSpace(raw), "\n")

	var title, source string
	var contentLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "**Judul:**"):
			title = strings.TrimSpace(strings.TrimPrefix(line, "**Judul:**"))

		case strings.HasPrefix(line, "**Sumber:**"):
			source = strings.TrimSpace(strings.TrimPrefix(line, "**Sumber:**"))

		case line != "" && !strings.HasPrefix(line, "**"):
			contentLines = append(contentLines, line)
		}
	}

	return &model_util.GeneratedArticle{
		Title:   title,
		Content: strings.Join(contentLines, "\n"),
		Source:  source,
		TopicID: topicID,
	}
}
