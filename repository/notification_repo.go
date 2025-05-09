package repository

import (
	"database/sql"
)

type NotifRepoInterface interface {
	SaveDeviceToken(userID int, token, platform string) error
	GetDeviceTokens(userID int) ([]string, error)
}

type notificationRepo struct {
	DB *sql.DB
}

func NewNotificationRepo(db *sql.DB) *notificationRepo {
	return &notificationRepo{DB: db}
}

// device token for push notif
func (r *notificationRepo) SaveDeviceToken(userID int, token, platform string) error {
	_, err := r.DB.Exec(`
        INSERT INTO user_devices (user_id, device_token, platform)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id, device_token) DO UPDATE
        SET platform = $3`,
		userID, token, platform,
	)
	return err
}

// get device token
func (r *notificationRepo) GetDeviceTokens(userID int) ([]string, error) {
	rows, err := r.DB.Query(`
        SELECT device_token 
        FROM user_devices 
        WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}
