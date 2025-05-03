package service

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type DeepSeekClient struct {
	APIKey string
}

func (d *DeepSeekClient) GetAIResponse(userInput string) (string, error) {
	url := "https://api.deepseek.com/v1/chat/completions"
	payload := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "user", "content": userInput},
		},
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

func NewDeepSeekClient(apiKey string) *DeepSeekClient {
	return &DeepSeekClient{APIKey: apiKey}
}
