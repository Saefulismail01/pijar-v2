package service

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"pijar/model"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// MidtransServiceInterface adalah interface untuk layanan Midtrans
type MidtransServiceInterface interface {
	Pay(req model.MidtransSnapReq) (model.MidtransSnapResp, error)
	VerifyCallback(callback model.MidtransCallbackRequest) error
	GenerateOrderID() string
	CheckTransactionStatus(orderID string) (string, error)
	CancelTransaction(orderID string) error
}

// midtransService adalah implementasi dari MidtransServiceInterface
type midtransService struct {
	client    *resty.Client
	url       string
	serverKey string
}

// Pay membuat transaksi pembayaran baru di Midtrans
func (m *midtransService) Pay(payload model.MidtransSnapReq) (model.MidtransSnapResp, error) {
	// Validasi server key
	if m.serverKey == "" {
		return model.MidtransSnapResp{}, errors.New("SERVER_KEY tidak ditemukan")
	}

	// Encode server key untuk Basic Auth
	encodedKey := base64.StdEncoding.EncodeToString([]byte(m.serverKey))

	// Kirim request ke Midtrans
	resp, err := m.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeader("Authorization", "Basic "+encodedKey).
		SetBody(payload).
		Post(m.url)

	// Log response untuk debugging
	log.Printf("Midtrans response status: %s", resp.Status())
	log.Printf("Midtrans response body: %s", string(resp.Body()))

	if err != nil {
		log.Printf("Error sending request to Midtrans: %v", err)
		return model.MidtransSnapResp{}, fmt.Errorf("error sending request to Midtrans: %w", err)
	}

	// Parse response
	var snapResp model.MidtransSnapResp
	err = json.Unmarshal(resp.Body(), &snapResp)
	if err != nil {
		log.Printf("Error unmarshalling Midtrans response: %v", err)
		return model.MidtransSnapResp{}, fmt.Errorf("error unmarshalling Midtrans response: %w", err)
	}

	// Check for error messages
	if len(snapResp.ErrorMessage) > 0 {
		errorMsg := strings.Join(snapResp.ErrorMessage, ", ")
		log.Printf("Midtrans error: %s", errorMsg)
		return model.MidtransSnapResp{}, errors.New(errorMsg)
	}

	// Generate redirect URL
	redirectURL := fmt.Sprintf("https://app.sandbox.midtrans.com/snap/v2/vtweb/%s", snapResp.Token)
	snapResp.RedirectUrl = redirectURL

	return snapResp, nil
}

// VerifyCallback memverifikasi callback dari Midtrans
func (m *midtransService) VerifyCallback(callback model.MidtransCallbackRequest) error {
	// Validasi server key
	if m.serverKey == "" {
		return errors.New("SERVER_KEY tidak ditemukan")
	}

	// Encode server key untuk Basic Auth
	encodedKey := base64.StdEncoding.EncodeToString([]byte(m.serverKey))

	// URL untuk status transaksi
	statusURL := fmt.Sprintf("https://api.sandbox.midtrans.com/v2/%s/status", callback.OrderID)

	// Kirim request ke Midtrans untuk verifikasi status
	resp, err := m.client.R().
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Basic "+encodedKey).
		Get(statusURL)

	if err != nil {
		log.Printf("Error verifying transaction status: %v", err)
		return fmt.Errorf("error verifying transaction status: %w", err)
	}

	// Parse response
	var statusResp map[string]interface{}
	err = json.Unmarshal(resp.Body(), &statusResp)
	if err != nil {
		log.Printf("Error unmarshalling status response: %v", err)
		return fmt.Errorf("error unmarshalling status response: %w", err)
	}

	// Verifikasi transaction_id
	transactionID, ok := statusResp["transaction_id"].(string)
	if !ok || transactionID != callback.TransactionID {
		log.Printf("Transaction ID mismatch: %s vs %s", transactionID, callback.TransactionID)
		return errors.New("transaction ID mismatch")
	}

	// Verifikasi status transaksi
	transactionStatus, ok := statusResp["transaction_status"].(string)
	if !ok || transactionStatus != callback.TransactionStatus {
		log.Printf("Transaction status mismatch: %s vs %s", transactionStatus, callback.TransactionStatus)
		return errors.New("transaction status mismatch")
	}

	return nil
}

// GenerateOrderID menghasilkan ID order unik
func (m *midtransService) GenerateOrderID() string {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("ORDER-%d", timestamp)
}

// CheckTransactionStatus memeriksa status transaksi di Midtrans
func (m *midtransService) CheckTransactionStatus(orderID string) (string, error) {
	// Validasi server key
	if m.serverKey == "" {
		return "", errors.New("SERVER_KEY tidak ditemukan")
	}

	// Encode server key untuk Basic Auth
	encodedKey := base64.StdEncoding.EncodeToString([]byte(m.serverKey))

	// URL untuk status transaksi
	statusURL := fmt.Sprintf("https://api.sandbox.midtrans.com/v2/%s/status", orderID)

	// Kirim request ke Midtrans untuk mendapatkan status
	resp, err := m.client.R().
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Basic "+encodedKey).
		Get(statusURL)

	if err != nil {
		log.Printf("Error checking transaction status: %v", err)
		return "", fmt.Errorf("error checking transaction status: %w", err)
	}

	// Parse response
	var statusResp struct {
		TransactionStatus string `json:"transaction_status"`
		StatusCode        string `json:"status_code"`
		StatusMessage     string `json:"status_message"`
	}

	err = json.Unmarshal(resp.Body(), &statusResp)
	if err != nil {
		log.Printf("Error unmarshalling status response: %v", err)
		return "", fmt.Errorf("error unmarshalling status response: %w", err)
	}

	// Log status for debugging
	log.Printf("Transaction status for order %s: %s", orderID, statusResp.TransactionStatus)

	return statusResp.TransactionStatus, nil
}

// CancelTransaction membatalkan transaksi di Midtrans
func (m *midtransService) CancelTransaction(orderID string) error {
	// Validasi server key
	if m.serverKey == "" {
		return errors.New("SERVER_KEY tidak ditemukan")
	}

	// Encode server key untuk Basic Auth
	encodedKey := base64.StdEncoding.EncodeToString([]byte(m.serverKey))

	// URL untuk cancel transaksi
	cancelURL := fmt.Sprintf("https://api.sandbox.midtrans.com/v2/%s/cancel", orderID)

	// Kirim request ke Midtrans untuk cancel transaksi
	resp, err := m.client.R().
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Basic "+encodedKey).
		Post(cancelURL)

	if err != nil {
		log.Printf("Error cancelling transaction: %v", err)
		return fmt.Errorf("error cancelling transaction: %w", err)
	}

	// Parse response
	var cancelResp struct {
		StatusCode    string `json:"status_code"`
		StatusMessage string `json:"status_message"`
	}

	err = json.Unmarshal(resp.Body(), &cancelResp)
	if err != nil {
		log.Printf("Error unmarshalling cancel response: %v", err)
		return fmt.Errorf("error unmarshalling cancel response: %w", err)
	}

	// Check if cancel was successful
	if cancelResp.StatusCode != "200" {
		log.Printf("Failed to cancel transaction: %s", cancelResp.StatusMessage)
		return fmt.Errorf("failed to cancel transaction: %s", cancelResp.StatusMessage)
	}

	log.Printf("Transaction %s cancelled successfully", orderID)
	return nil
}

// NewMidtransService membuat instance baru dari MidtransService
func NewMidtransService(client *resty.Client) MidtransServiceInterface {
	serverKey := os.Getenv("SERVER_KEY")
	if serverKey == "" {
		log.Println("WARNING: SERVER_KEY environment variable is not set")
	}

	return &midtransService{
		client:    client,
		url:       "https://app.sandbox.midtrans.com/snap/v1/transactions",
		serverKey: serverKey,
	}
}
