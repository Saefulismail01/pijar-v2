package model

import "time"

type CoachSession struct {
	ID          int
	UserID      int
	Timestamp   time.Time
	UserInput   string
	AIResponse  string
}