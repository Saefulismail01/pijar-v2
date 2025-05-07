package model

import "time"

type CoachSession struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	SessionID   string    `json:"session_id"`
	Timestamp   time.Time `json:"timestamp"`
	UserInput   string    `json:"user_input"`
	AIResponse  string    `json:"ai_response"`
}

type ConversationContext struct {
	SessionID string         `json:"session_id"`
	Messages  []Message      `json:"messages"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type Message struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"`
}