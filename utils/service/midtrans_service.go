package service

import (
	"encoding/json"
	"fmt"
	"os"
	"pijar/model"
	"time"
	"github.com/go-resty/resty/v2"
)

// MidtransServiceInterface defines the interface for Midtrans payment service
type MidtransServiceInterface interface {
	CreateTransaction(orderID string, amount int, customerDetails model.CustomerDetails, itemDetails []model.Item) (string, error)
	CheckTransactionStatus(orderID string) (model.MidtransCallbackRequest, error)
	GenerateOrderID() string
	Pay(req model.MidtransSnapReq) (model.MidtransSnapResp, error)
	VerifyCallback(callback model.MidtransCallbackRequest) error
}

// MidtransService implements the MidtransServiceInterface
type MidtransService struct {
	client *resty.Client
	serverKey string
}

// NewMidtransService creates a new instance of MidtransService
func NewMidtransService(client *resty.Client) MidtransServiceInterface {
	serverKey := os.Getenv("SERVER_KEY")
	if serverKey == "" {
		panic("SERVER_KEY environment variable is not set")
	}
	
	return &MidtransService{
		client:     client,
		serverKey: serverKey,
	}
}

// CreateTransaction creates a new payment transaction in Midtrans
func (s *MidtransService) CreateTransaction(orderID string, amount int, customerDetails model.CustomerDetails, itemDetails []model.Item) (string, error) {
	// Prepare request body
	reqBody := map[string]interface{}{
		"transaction_details": map[string]interface{}{
			"order_id":  orderID,
			"gross_amount": amount,
		},
		"customer_details": map[string]interface{}{
			"first_name": customerDetails.FirstName,
			"last_name":  customerDetails.LastName,
			"email":     customerDetails.Email,
			"phone":     customerDetails.Phone,
		},
		"item_details": itemDetails,
	}

	// Set headers
	s.client.SetHeader("Accept", "application/json")
	s.client.SetHeader("Content-Type", "application/json")
	s.client.SetHeader("Authorization", fmt.Sprintf("Basic %s", s.serverKey))

	// Make API call
	resp, err := s.client.R().
		SetBody(reqBody).
		Post("https://api.sandbox.midtrans.com/v2/charge")

	if err != nil {
		return "", fmt.Errorf("error creating transaction: %v", err)
	}

	if resp.IsError() {
		return "", fmt.Errorf("midtrans API error: %s", resp.Status())
	}

	return resp.String(), nil
}

// CheckTransactionStatus checks the status of a transaction in Midtrans
func (s *MidtransService) CheckTransactionStatus(orderID string) (model.MidtransCallbackRequest, error) {
	// Make API call to get transaction status
	s.client.SetHeader("Accept", "application/json")
	s.client.SetHeader("Content-Type", "application/json")
	s.client.SetHeader("Authorization", fmt.Sprintf("Basic %s", s.serverKey))

	resp, err := s.client.R().
		Get(fmt.Sprintf("https://api.sandbox.midtrans.com/v2/%s/status", orderID))

	if err != nil {
		return model.MidtransCallbackRequest{}, fmt.Errorf("error checking transaction status: %v", err)
	}

	if resp.IsError() {
		return model.MidtransCallbackRequest{}, fmt.Errorf("midtrans API error: %s", resp.Status())
	}

	// Parse response
	var callback model.MidtransCallbackRequest
	if err := json.Unmarshal(resp.Body(), &callback); err != nil {
		return model.MidtransCallbackRequest{}, fmt.Errorf("error parsing response: %v", err)
	}

	return callback, nil
}

// GenerateOrderID generates a unique order ID for Midtrans transactions
func (s *MidtransService) GenerateOrderID() string {
	// Generate a unique order ID based on timestamp
	return fmt.Sprintf("ORDER-%d", time.Now().UnixNano())
}

// Pay processes a payment request through Midtrans Snap API
func (s *MidtransService) Pay(req model.MidtransSnapReq) (model.MidtransSnapResp, error) {
	// Prepare request body
	reqBody := map[string]interface{}{
		"transaction_details": map[string]interface{}{
			"order_id":  req.TransactionDetails.OrderID,
			"gross_amount": req.TransactionDetails.GrossAmt,
		},
		"customer_details": map[string]interface{}{
			"first_name": req.CustomerDetails.FirstName,
			"last_name":  req.CustomerDetails.LastName,
			"email":     req.CustomerDetails.Email,
			"phone":     req.CustomerDetails.Phone,
		},
		"item_details": req.Item,
	}

	// Set headers
	s.client.SetHeader("Accept", "application/json")
	s.client.SetHeader("Content-Type", "application/json")
	s.client.SetHeader("Authorization", fmt.Sprintf("Basic %s", s.serverKey))

	// Make API call
	resp, err := s.client.R().
		SetBody(reqBody).
		Post("https://app.sandbox.midtrans.com/snap/v2/transactions")

	if err != nil {
		return model.MidtransSnapResp{}, fmt.Errorf("error creating payment: %v", err)
	}

	if resp.IsError() {
		return model.MidtransSnapResp{}, fmt.Errorf("midtrans API error: %s", resp.Status())
	}

	// Parse response
	var snapResp model.MidtransSnapResp
	if err := json.Unmarshal(resp.Body(), &snapResp); err != nil {
		return model.MidtransSnapResp{}, fmt.Errorf("error parsing response: %v", err)
	}

	return snapResp, nil
}

// VerifyCallback verifies the callback notification from Midtrans
func (s *MidtransService) VerifyCallback(callback model.MidtransCallbackRequest) error {
	// Verify signature
	// In a real implementation, you would:
	// 1. Calculate the signature using the server key and order ID
	// 2. Compare it with the signature in the callback
	// 3. Return error if they don't match
	
	// For now, just return success
	return nil
}
