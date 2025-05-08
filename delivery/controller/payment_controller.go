package controller

import (
	"konsep_project/model"
	"konsep_project/usecase"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// PaymentController adalah interface untuk controller pembayaran
type PaymentController interface {
	CreatePayment(c *gin.Context)
	GetPaymentStatus(c *gin.Context)
	Route()
}

// paymentController adalah implementasi dari PaymentController
type paymentController struct {
	paymentUsecase usecase.PaymentUsecase
	rg             *gin.RouterGroup
}

// NewPaymentController membuat instance baru dari PaymentController
func NewPaymentController(
	rg *gin.RouterGroup,
	paymentUsecase usecase.PaymentUsecase,
) PaymentController {
	return &paymentController{
		rg:             rg,
		paymentUsecase: paymentUsecase,
	}
}

// Route mengatur endpoint untuk payment
func (p *paymentController) Route() {
	paymentRoutes := p.rg.Group("/payments")
	paymentRoutes.POST("/", p.CreatePayment)
	paymentRoutes.GET("/:id", p.GetPaymentStatus)
}

// CreatePayment membuat pembayaran baru
func (p *paymentController) CreatePayment(c *gin.Context) {
	var req model.PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Panggil usecase untuk membuat pembayaran
	transaction, err := p.paymentUsecase.CreatePayment(req)
	if err != nil {
		log.Printf("Error creating payment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"transaction_id": transaction.ID,
		"order_id":       transaction.OrderID,
		"payment_url":    transaction.PaymentURL,
		"status":         transaction.Status,
	})
}

// GetPaymentStatus mendapatkan status pembayaran
func (p *paymentController) GetPaymentStatus(c *gin.Context) {
	// Ambil ID transaksi dari parameter
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	// Panggil usecase untuk mendapatkan status pembayaran
	transaction, err := p.paymentUsecase.GetPaymentStatus(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	// Return status
	c.JSON(http.StatusOK, gin.H{
		"transaction_id": transaction.ID,
		"order_id":       transaction.OrderID,
		"payment_url":    transaction.PaymentURL,
		"status":         transaction.Status,
		"amount":         transaction.Amount,
		"created_at":     transaction.CreatedAt,
		"updated_at":     transaction.UpdatedAt,
	})
}
