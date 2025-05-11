package model

import (
	"time"
)

type Transaction struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	ProductID  int       `json:"product_id"`
	Amount     int       `json:"amount"`
	Status     string    `json:"status"`
	OrderID    string    `json:"order_id"`
	PaymentURL string    `json:"payment_url"`
	MidtransID string    `json:"midtrans_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
