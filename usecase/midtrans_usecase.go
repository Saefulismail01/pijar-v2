package usecase

import (
	"fmt"
	"pijar/model"
	"pijar/repository"
	"pijar/utils/service"
	"log"
	"strconv"
	"time"
)

// PaymentUsecase adalah interface untuk usecase pembayaran
type PaymentUsecase interface {
	CreatePayment(req model.PaymentRequest) (model.Transaction, error)
	GetPaymentStatus(id int) (model.Transaction, error)
	ProcessCallback(callback model.MidtransCallbackRequest) error
}

// paymentUsecase adalah implementasi dari PaymentUsecase
type paymentUsecase struct {
	midtransService  service.MidtransServiceInterface
	productRepo      repository.ProductRepository
	transactionRepo  repository.TransactionRepository
}

// NewPaymentUsecase membuat instance baru dari PaymentUsecase
func NewPaymentUsecase(
	midtransService service.MidtransServiceInterface,
	productRepo repository.ProductRepository,
	transactionRepo repository.TransactionRepository,
) PaymentUsecase {
	return &paymentUsecase{
		midtransService: midtransService,
		productRepo:     productRepo,
		transactionRepo: transactionRepo,
	}
}

// CreatePayment membuat pembayaran baru
func (p *paymentUsecase) CreatePayment(req model.PaymentRequest) (model.Transaction, error) {
	// Validasi product
	product, err := p.productRepo.GetProductByID(req.ProductID)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("product not found: %w", err)
	}

	// Generate order ID
	orderID := p.midtransService.GenerateOrderID()

	
	// Buat item untuk Midtrans
	items := []model.Item{
		{
			ID:                 strconv.Itoa(product.ID),
			Name:               product.Name,
			Price:              product.Price,
			Quantity:           req.Quantity,
			MonthlySubscription: 1, // Default 1 bulan
		},
	}

	// Hitung total amount
	totalAmount := product.Price * req.Quantity

	// Buat request untuk Midtrans
	midtransReq := model.MidtransSnapReq{
		TransactionDetails: struct {
			OrderID  string `json:"order_id"`
			GrossAmt int    `json:"gross_amount"`
		}{
			OrderID:  orderID,
			GrossAmt: totalAmount,
		},
		CustomerDetails: model.CustomerDetails{
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Email:     req.Email,
			Phone:     req.Phone,
		},
		Item: items,
	}

	// Kirim request ke Midtrans
	resp, err := p.midtransService.Pay(midtransReq)
	if err != nil {
		log.Printf("Error creating payment: %v", err)
		return model.Transaction{}, fmt.Errorf("error creating payment: %w", err)
	}

	// Simpan transaksi ke database
	transaction := model.Transaction{
		UserID:     req.UserID,
		ProductID:  req.ProductID,
		Amount:     totalAmount,
		Status:     "pending", // Status awal
		OrderID:    orderID,
		PaymentURL: resp.RedirectUrl,
		MidtransID: "", // Akan diupdate saat callback
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

// GetPaymentStatus mendapatkan status pembayaran
func (p *paymentUsecase) GetPaymentStatus(id int) (model.Transaction, error) {
	// Ambil transaksi dari database
	transaction, err := p.transactionRepo.GetTransactionByID(id)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("transaction not found: %w", err)
	}

	return transaction, nil
}

// ProcessCallback memproses callback dari Midtrans
func (p *paymentUsecase) ProcessCallback(callback model.MidtransCallbackRequest) error {
	// Verifikasi callback
	err := p.midtransService.VerifyCallback(callback)
	if err != nil {
		log.Printf("Error verifying callback: %v", err)
		return fmt.Errorf("error verifying callback: %w", err)
	}

	// Tentukan status transaksi berdasarkan status dari Midtrans
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

	// Update status transaksi di database
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