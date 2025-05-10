package usecase

import (
	"fmt"
	"log"
	"pijar/model"
	"pijar/repository"
	"pijar/utils/service"
	"strconv"
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
	quantity := 1

	// Create item for Midtrans
	items := []model.Item{
		{
			ID:                  strconv.Itoa(product.ID),
			Name:                product.Name,
			Price:               product.Price,
			Quantity:            quantity,
			MonthlySubscription: 1,
		},
	}

	// Calculate total amount
	totalAmount := product.Price * quantity

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
			FirstName: "Customer",
			LastName:  "",
			Email:     "",
			Phone:     "",
		},
		Item: items,
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
	transaction, err = p.ForceCheckAndUpdateStatus(id)
	if err != nil {
		log.Printf("Warning: Could not check Midtrans status: %v", err)
	}

	return transaction, nil
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
		status = "unknown"
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
		newStatus = "unknown"
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

	// Hanya boleh rollback transaksi yang masih pending
	if transaction.Status != "pending" {
		return transaction, fmt.Errorf("cannot rollback non-pending transaction with status: %s", transaction.Status)
	}

	// Panggil Midtrans API untuk cancel transaksi
	err = p.midtransService.CancelTransaction(transaction.OrderID)
	if err != nil {
		return transaction, fmt.Errorf("failed to cancel transaction: %w", err)
	}

	// Update status transaksi
	transaction.Status = "cancelled"
	transaction.UpdatedAt = time.Now()

	err = p.transactionRepo.UpdateTransactionStatus(transaction.ID, "cancelled")
	if err != nil {
		return transaction, fmt.Errorf("failed to update transaction status: %w", err)
	}

	return transaction, nil
}
