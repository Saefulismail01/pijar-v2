package service

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type DeepSeekClient struct {
	APIKey       string
	SystemPrompt string
	Temperature  float64
	MaxTokens    int
}

func (d *DeepSeekClient) GetAIResponse(userInput string) (string, error) {
	url := "https://api.deepseek.com/v1/chat/completions"
	
	messages := []map[string]string{}
	
	if d.SystemPrompt != "" {
		messages = append(messages, map[string]string{
			"role":    "system",
			"content": d.SystemPrompt,
		})
	}
	
	messages = append(messages, map[string]string{
		"role":    "user",
		"content": userInput,
	})
	
	payload := map[string]interface{}{
		"model":       "deepseek-chat",
		"messages":    messages,
		"temperature": d.Temperature,
	}
	
	if d.MaxTokens > 0 {
		payload["max_tokens"] = d.MaxTokens
	}
	
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	resBody, _ := ioutil.ReadAll(resp.Body)
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	json.Unmarshal(resBody, &result)

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}
	return "", nil
}

func (d *DeepSeekClient) WithSystemPrompt(prompt string) *DeepSeekClient {
	d.SystemPrompt = prompt
	return d
}

func (d *DeepSeekClient) WithTemperature(temp float64) *DeepSeekClient {
	d.Temperature = temp
	return d
}

func (d *DeepSeekClient) WithMaxTokens(tokens int) *DeepSeekClient {
	d.MaxTokens = tokens
	return d
}


func NewDeepSeekClient(apiKey string) *DeepSeekClient {
	return &DeepSeekClient{
		APIKey:       apiKey,
		SystemPrompt: "",
		Temperature:  0.7,
		MaxTokens:    1000,
	}
}
