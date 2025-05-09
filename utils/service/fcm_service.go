package service

import (
	"context"

	"firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
)

type FCMClient struct {
	client *messaging.Client
}

// Inisialisasi FCM Client
func NewFCMClient(app *firebase.App) (*FCMClient, error) {
	client, err := app.Messaging(context.Background())
	if err != nil {
		return nil, err
	}
	return &FCMClient{client: client}, nil
}

// Method untuk mengirim notifikasi
func (f *FCMClient) SendNotification(tokens []string, title, body string) error {
	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
	}
	_, err := f.client.SendMulticast(context.Background(), message)
	return err
}
