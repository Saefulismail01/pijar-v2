package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"pijar/model"
)

type DeepSeekClient struct {
	APIKey string
	SystemPrompt string
	UserPrompt string
	Temperature float64
	MaxTokens int
}

// GetAIResponseWithContext mengirim permintaan ke DeepSeek API dengan konteks percakapan yang ada
func (d *DeepSeekClient) GetAIResponseWithContext(messages []model.Message) (string, error) {
	url := "https://api.deepseek.com/v1/chat/completions"

	// Konversi dari model.Message ke format yang diharapkan API
	apiMessages := make([]map[string]string, 0, len(messages))
	for _, msg := range messages {
		apiMessages = append(apiMessages, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	// Jika ada system prompt, tambahkan di awal
	if d.SystemPrompt != "" {
		systemMsg := map[string]string{
			"role":    "system",
			"content": d.SystemPrompt,
		}
		apiMessages = append([]map[string]string{systemMsg}, apiMessages...)
	}

	payload := map[string]interface{}{
		"model":       "deepseek-chat",
		"messages":    apiMessages,
		"temperature": d.Temperature,
	}

	if d.MaxTokens > 0 {
		payload["max_tokens"] = d.MaxTokens
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("gagal mengencode payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("gagal membuat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("gagal mengirim request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("error dari API: %s", string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("gagal mendecode respons: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("tidak ada respons yang diterima")
	}

	return result.Choices[0].Message.Content, nil
}

// GetAIResponse adalah wrapper untuk kompatibilitas mundur
func (d *DeepSeekClient) GetAIResponse(userInput string) (string, error) {
	messages := []model.Message{{
		Role:    "user",
		Content: userInput,
	}}
	return d.GetAIResponseWithContext(messages)
}

func NewDeepSeekClient(apiKey string) *DeepSeekClient {
	return &DeepSeekClient{APIKey: apiKey}
}
