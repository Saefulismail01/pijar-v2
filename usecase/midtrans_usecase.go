package usecase

import (
	"fmt"
	"log"
	"pijar/model"
	"pijar/repository"
	"pijar/utils/service"
	"strconv"
	"strings"
	"time"
)

// PaymentUsecase interface for payment operations
type PaymentUsecase interface {
	CreatePayment(req model.PaymentRequest) (model.Transaction, error)
	GetPaymentStatus(id int) (model.Transaction, error)
	ProcessCallback(callback model.MidtransCallbackRequest) error
	RollbackPayment(id int) (model.Transaction, error)
	ForceCheckAndUpdateStatus(id int) (model.Transaction, error)
	GetProductByID(id int) (model.Product, error)
	GetUserByID(id int) (model.Users, error)
}

// paymentUsecase implements PaymentUsecase interface
type paymentUsecase struct {
	midtransService service.MidtransServiceInterface
	productRepo     repository.ProductRepository
	transactionRepo repository.TransactionRepository
	userRepo        repository.UserRepoInterface
}

// NewPaymentUsecase creates a new PaymentUsecase instance
func NewPaymentUsecase(
	midtransService service.MidtransServiceInterface,
	productRepo repository.ProductRepository,
	transactionRepo repository.TransactionRepository,
	userRepo repository.UserRepoInterface,
) PaymentUsecase {
	return &paymentUsecase{
		midtransService: midtransService,
		productRepo:     productRepo,
		transactionRepo: transactionRepo,
		userRepo:        userRepo,
	}
}

// CreatePayment creates a new payment
func (p *paymentUsecase) CreatePayment(req model.PaymentRequest) (model.Transaction, error) {
	// Validate product
	product, err := p.productRepo.GetProductByID(req.ProductID)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("product not found: %w", err)
	}

	// Generate order ID and set default quantity
	orderID := p.midtransService.GenerateOrderID()

	// Create item for Midtrans
	items := []model.Item{
		{
			ID:        strconv.Itoa(product.ID),
			Name:      product.Name,
			Price:     product.Price,
			Quantity:  1,
		},
	}

	// Calculate total amount
	totalAmount := product.Price * 1

	// Validate that total amount matches item total
	itemTotal := product.Price * 1
	if totalAmount != itemTotal {
		return model.Transaction{}, fmt.Errorf("total amount mismatch: expected %d, got %d", itemTotal, totalAmount)
	}

	// Create Midtrans request with minimal customer details
	midtransReq := model.MidtransSnapReq{
		TransactionDetails: struct {
			OrderID  string `json:"order_id"`
			GrossAmt int    `json:"gross_amount"`
		}{
			OrderID:  orderID,
			GrossAmt: totalAmount,
		},
		CustomerDetails: model.CustomerDetails{
			Name:  "Customer",
			Phone: "",
		},
		ItemDetails: items,
	}

	// Send request to Midtrans
	resp, err := p.midtransService.Pay(midtransReq)
	if err != nil {
		log.Printf("Error creating payment: %v", err)
		return model.Transaction{}, fmt.Errorf("error creating payment: %w", err)
	}

	// Save transaction to database
	transaction := model.Transaction{
		UserID:     req.UserID,
		ProductID:  req.ProductID,
		Amount:     totalAmount,
		Status:     "pending",
		OrderID:    orderID,
		PaymentURL: resp.RedirectUrl,
		MidtransID: "",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	savedTransaction, err := p.transactionRepo.CreateTransaction(transaction)
	if err != nil {
		log.Printf("Error saving transaction: %v", err)
		return model.Transaction{}, fmt.Errorf("error saving transaction: %w", err)
	}

	return savedTransaction, nil
}

// GetPaymentStatus retrieves the current transaction status
func (p *paymentUsecase) GetPaymentStatus(id int) (model.Transaction, error) {
	transaction, err := p.transactionRepo.GetTransactionByID(id)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("transaction not found: %w", err)
	}

	// Check latest status from Midtrans
	log.Printf("Checking latest status from Midtrans for transaction %d", id)
	updatedTransaction, err := p.ForceCheckAndUpdateStatus(id)
	if err != nil {
		log.Printf("Warning: Could not check Midtrans status: %v", err)
		return transaction, nil
	}
	return updatedTransaction, nil
}

// ProcessCallback handles Midtrans callback and updates transaction status
func (p *paymentUsecase) ProcessCallback(callback model.MidtransCallbackRequest) error {
	err := p.midtransService.VerifyCallback(callback)
	if err != nil {
		log.Printf("Error verifying callback: %v", err)
		return fmt.Errorf("error verifying callback: %w", err)
	}

	// Determine transaction status based on Midtrans status
	var status string
	switch callback.TransactionStatus {
	case "capture", "settlement":
		status = "success"
	case "pending":
		status = "pending"
	case "deny", "cancel", "expire", "failure":
		status = "failed"
	default:
		// For any other status, treat as pending to allow rollback
		status = "pending"
	}

	// Update transaction status in database
	err = p.transactionRepo.UpdateTransactionStatusByOrderID(
		callback.OrderID,
		status,
		callback.TransactionID,
	)

	if err != nil {
		log.Printf("Error updating transaction status: %v", err)
		return fmt.Errorf("error updating transaction status: %w", err)
	}

	return nil
}

// ForceCheckAndUpdateStatus forces a status check from Midtrans and updates the database
func (p *paymentUsecase) ForceCheckAndUpdateStatus(id int) (model.Transaction, error) {
	transaction, err := p.transactionRepo.GetTransactionByID(id)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("transaction not found: %w", err)
	}

	// Check transaction status from Midtrans API
	midtransStatus, err := p.midtransService.CheckTransactionStatus(transaction.OrderID)
	if err != nil {
		return transaction, fmt.Errorf("error checking transaction status: %w", err)
	}

	// Determine transaction status based on Midtrans status
	var newStatus string
	switch midtransStatus {
	case "capture", "settlement":
		newStatus = "success"
	case "pending":
		newStatus = "pending"
	case "deny", "cancel", "expire", "failure":
		newStatus = "failed"
	default:
		// For any other status, treat as pending to allow rollback
		newStatus = "pending"
	}
	if newStatus != transaction.Status {
		log.Printf("Updating transaction %d status from %s to %s", id, transaction.Status, newStatus)

		// Update transaksi di database
		transaction.Status = newStatus
		transaction.UpdatedAt = time.Now()

		err = p.transactionRepo.UpdateTransactionStatus(transaction.ID, newStatus)
		if err != nil {
			return transaction, fmt.Errorf("error updating transaction status: %w", err)
		}
	}

	return transaction, nil
}

// GetProductByID mendapatkan detail produk berdasarkan ID
func (p *paymentUsecase) GetProductByID(id int) (model.Product, error) {
	return p.productRepo.GetProductByID(id)
}

// GetUserByID mendapatkan detail user berdasarkan ID
func (p *paymentUsecase) GetUserByID(id int) (model.Users, error) {
	return p.userRepo.GetUserByID(id)
}

// RollbackPayment membatalkan pembayaran
func (p *paymentUsecase) RollbackPayment(id int) (model.Transaction, error) {
	// Ambil transaksi dari database
	transaction, err := p.transactionRepo.GetTransactionByID(id)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("transaction not found: %w", err)
	}

	// Hanya boleh rollback transaksi yang belum selesai
	if transaction.Status == "success" || transaction.Status == "cancelled" {
		return transaction, fmt.Errorf("cannot rollback transaction with status: %s", transaction.Status)
	}

	// Coba cancel transaksi di Midtrans
	err = p.midtransService.CancelTransaction(transaction.OrderID)
	if err != nil {
		// Jika transaksi tidak ditemukan di Midtrans (404), lanjutkan dengan update status
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") || 
		   strings.Contains(err.Error(), "Transaction doesn't exist") {
			log.Printf("Transaction %s not found in Midtrans, updating local status to cancelled", transaction.OrderID)
		} else {
			return transaction, fmt.Errorf("failed to cancel transaction: %w", err)
		}
	}

	// Update status transaksi di database
	transaction.Status = "cancelled"
	transaction.UpdatedAt = time.Now()

	err = p.transactionRepo.UpdateTransactionStatus(transaction.ID, "cancelled")
	if err != nil {
		return transaction, fmt.Errorf("failed to update transaction status: %w", err)
	}

	return transaction, nil
}
