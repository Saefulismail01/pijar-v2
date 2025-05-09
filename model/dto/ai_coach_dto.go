package dto

import "pijar/model"

type CoachRequest struct {
	UserInput string `json:"user_input"`
}

type StartSessionResponse struct {
	SessionID string `json:"session_id"`
	Response  string `json:"response"`
}

type ContinueSessionRequest struct {
	UserInput string `json:"user_input"`
}

type SessionHistoryResponse struct {
	SessionID string          `json:"session_id"`
	Messages  []model.Message `json:"messages"`
}