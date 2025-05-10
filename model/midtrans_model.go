package model

type Item struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Price              int    `json:"price"`
	Quantity           int    `json:"quantity"`
	MonthlySubscription int   `json:"monthly_subscription"`
}

type CustomerDetails struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

type MidtransSnapReq struct {
	TransactionDetails struct {
		OrderID  string `json:"order_id"`
		GrossAmt int    `json:"gross_amount"`
	} `json:"transaction_details"`
	Customer        string         `json:"customer"`
	CustomerDetails CustomerDetails `json:"customer_details,omitempty"`
	Item            []Item         `json:"item_details"`
}

type MidtransSnapResp struct {
	Token         string   `json:"token"`
	RedirectUrl   string   `json:"redirect_url"`
	ErrorMessage  []string `json:"error_messages,omitempty"`
	StatusCode    string   `json:"status_code,omitempty"`
	StatusMessage string   `json:"status_message,omitempty"`
}

// MidtransCallbackRequest adalah model untuk callback dari Midtrans
type MidtransCallbackRequest struct {
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	TransactionID     string `json:"transaction_id"`
	StatusMessage     string `json:"status_message"`
	StatusCode        string `json:"status_code"`
	SignatureKey      string `json:"signature_key"`
	PaymentType       string `json:"payment_type"`
	OrderID           string `json:"order_id"`
	MerchantID        string `json:"merchant_id"`
	GrossAmount       string `json:"gross_amount"`
	FraudStatus       string `json:"fraud_status"`
	Currency          string `json:"currency"`
}

// PaymentRequest adalah model untuk request pembayaran dari client
type PaymentRequest struct {
	UserID    int `json:"user_id"`
	ProductID int `json:"product_id"`
}