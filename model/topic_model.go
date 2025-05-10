package model

type TopicUser struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	Preference string `json:"preference"`
}
