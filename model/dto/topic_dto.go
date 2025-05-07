package dto

type InputTopic struct {
	Preference string `json:"preference" binding:"required"`
	UserID     int    `json:"user_id" binding:"required"`	
}