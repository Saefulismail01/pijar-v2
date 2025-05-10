package delivery

import (
	"context"
	"fmt"
	
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

func initializeFirebase() (*firebase.App, error) {
	opt := option.WithCredentialsFile("config/firebase-credentials.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("firebase init error: %v", err)
	}
	return app, nil
}
