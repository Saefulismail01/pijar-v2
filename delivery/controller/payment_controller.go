package controller

import (
	"log"
	"net/http"
	"pijar/model"
	"pijar/model/dto"
	"pijar/usecase"
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
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Bad Request",
			Error:   err.Error(),
		})
		return
	}

	// Panggil usecase untuk membuat pembayaran
	transaction, err := p.paymentUsecase.CreatePayment(req)
	if err != nil {
		log.Printf("Error creating payment: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Internal Server Error",
			Error:   err.Error(),
		})
		return
	}

	// Return response
	c.JSON(http.StatusOK, dto.Response{
		Message: "Payment created successfully",
		Data: gin.H{
			"transaction_id": transaction.ID,
			"order_id":       transaction.OrderID,
			"payment_url":    transaction.PaymentURL,
			"status":         transaction.Status,
		},
	})
}

// GetPaymentStatus mendapatkan status pembayaran
func (p *paymentController) GetPaymentStatus(c *gin.Context) {
	// Ambil ID transaksi dari parameter
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Bad Request",
			Error:   "Invalid transaction ID",
		})
		return
	}

	// Panggil usecase untuk mendapatkan status pembayaran
	transaction, err := p.paymentUsecase.GetPaymentStatus(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Message: "Not Found",
			Error:   "Transaction not found",
		})
		return
	}

	// Return status
	c.JSON(http.StatusOK, dto.Response{
		Message: "Transaction status retrieved successfully",
		Data: gin.H{
			"transaction_id": transaction.ID,
			"order_id":       transaction.OrderID,
			"payment_url":    transaction.PaymentURL,
			"status":         transaction.Status,
			"amount":         transaction.Amount,
			"created_at":     transaction.CreatedAt,
			"updated_at":     transaction.UpdatedAt,
		},
	})
}
