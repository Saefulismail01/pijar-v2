package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"pijar/model"
)

type GeminiClient struct {
	APIKey      string
	SystemPrompt string
	UserPrompt  string
	Temperature float64
	MaxTokens   int
}

// GetAIResponseWithContext mengirim permintaan ke Gemini API dengan konteks percakapan yang ada
func (g *GeminiClient) GetAIResponseWithContext(messages []model.Message) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=%s", g.APIKey)
	
	// Konversi dari model.Message ke format Gemini API
	contents := make([]map[string]interface{}, 0)
	
	// Gabungkan system prompt dengan user input pertama jika ada
	var firstUserMessage string
	if g.SystemPrompt != "" && len(messages) > 0 {
		firstUserMessage = g.SystemPrompt + "\n\n" + messages[0].Content
		messages = messages[1:] // Skip first message karena sudah digabung
	}
	
	// Tambahkan user input pertama (dengan system prompt jika ada)
	if firstUserMessage != "" {
		content := map[string]interface{}{
			"role": "user",
			"parts": []map[string]string{
				{"text": firstUserMessage},
			},
		}
		contents = append(contents, content)
	}
	
	// Tambahkan message lainnya dengan role mapping
	for _, msg := range messages {
		role := "user"
		if msg.Role == "assistant" {
			role = "model"
		}
		
		content := map[string]interface{}{
			"role": role,
			"parts": []map[string]string{
				{"text": msg.Content},
			},
		}
		contents = append(contents, content)
	}
	
	// Jika tidak ada konten sama sekali, buat dari system prompt saja
	if len(contents) == 0 {
		prompt := g.SystemPrompt
		if prompt == "" {
			prompt = "Hello"
		}
		content := map[string]interface{}{
			"role": "user",
			"parts": []map[string]string{
				{"text": prompt},
			},
		}
		contents = append(contents, content)
	}
	
	payload := map[string]interface{}{
		"contents": contents,
	}
	
	// Gemini API menggunakan generationConfig untuk parameter seperti temperature
	generationConfig := map[string]interface{}{}
	if g.Temperature > 0 {
		generationConfig["temperature"] = g.Temperature
	}
	if g.MaxTokens > 0 {
		generationConfig["maxOutputTokens"] = g.MaxTokens
	}
	
	if len(generationConfig) > 0 {
		payload["generationConfig"] = generationConfig
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
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("gagal mengirim request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("error dari API (status %d): %s", resp.StatusCode, string(body))
	}
	
	// Struktur respons Gemini API
	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("gagal mendecode respons: %w", err)
	}
	
	if len(result.Candidates) == 0 {
		return "", fmt.Errorf("tidak ada respons yang diterima")
	}
	
	if len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("tidak ada konten dalam respons")
	}
	
	return result.Candidates[0].Content.Parts[0].Text, nil
}

// GetAIResponse adalah wrapper untuk kompatibilitas mundur
func (g *GeminiClient) GetAIResponse(userInput string) (string, error) {
	messages := []model.Message{{
		Role:    "user",
		Content: userInput,
	}}
	return g.GetAIResponseWithContext(messages)
}

func NewGeminiClient(apiKey string) *GeminiClient {
	return &GeminiClient{APIKey: apiKey}
}