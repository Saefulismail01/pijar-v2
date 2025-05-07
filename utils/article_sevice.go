package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"pijar/utils/model_util"

	"github.com/gin-gonic/gin"
)

func GenerateArticles(c *gin.Context, preferences []string) ([]*model_util.GeneratedArticle, error) {
	var result []*model_util.GeneratedArticle
	rand.Seed(time.Now().UnixNano())

	preferensiPool := []string{"teknologi", "bisnis"}

	limit := len(preferences)
	if limit > 2 {
		limit = 2
	}

	for i := 0; i < limit; i++ {
		preference := preferences[i]

		// Override preferensi jika ada 3 atau lebih
		if len(preferences) >= 3 {
			preference = preferensiPool[rand.Intn(len(preferensiPool))]
		}

		fmt.Printf("🔍 Memproses preferensi: %s\n", preference)

		article, err := GenerateArticleFromDeepseek(c, preference, preference, i+1) // TopicID bisa dummy
		if err != nil {
			fmt.Printf("❌ Gagal generate artikel untuk preferensi '%s': %v\n", preference, err)
			continue
		}

		result = append(result, article)
	}

	return result, nil
}

func GenerateArticleFromDeepseek(c *gin.Context, topic string, preference string, topicID int) (*model_util.GeneratedArticle, error) {
	// Get API key from context that was set in middleware
	apiKey, exists := c.Get("deepseek_api_key")
	if !exists {
		return nil, fmt.Errorf("deepseek API key not found in context")
	}

	prompt := fmt.Sprintf(`Buat artikel tentang %s dengan format ketat:

1. **Judul:** [1 judul informatif]
2. **Isi:** 
   - [minimal 300 kata, masing-masing 1 paragraf pendek]
   - Setiap paragraf dipisahkan oleh baris baru
3. **Sumber:** [1 referensi/sumber]
4. **Preferensi:** [%s]`, topic, preference)

	reqBody := model_util.DeepseekChatRequest{
		Model: "deepseek-chat",
		Messages: []model_util.DeepseekChatMessage{
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

	// Use API key from context instead of environment variable
	req.Header.Set("Authorization", "Bearer "+apiKey.(string))
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

	var result model_util.DeepseekChatResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Println("Respon mentah:\n", string(respBody))
		return nil, fmt.Errorf("gagal decode response JSON: %w", err)
	}

	if len(result.Choices) == 0 || result.Choices[0].Message.Content == "" {
		fmt.Println("Respon mentah:\n", string(respBody))
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
