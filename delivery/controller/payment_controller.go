package controller

import (
	"log"
	"net/http"
	"pijar/middleware"
	"pijar/model"
	"pijar/model/dto"
	"pijar/usecase"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// PaymentController is an interface for payment controller
type PaymentController interface {
	CreatePayment(c *gin.Context)
	GetPaymentStatus(c *gin.Context)
	RollbackPayment(c *gin.Context)
	ForceCheckStatus(c *gin.Context)
	Route()
}

// paymentController is the implementation of PaymentController
type paymentController struct {
	paymentUsecase usecase.PaymentUsecase
	rg             *gin.RouterGroup
	aM             middleware.AuthMiddleware
}

// NewPaymentController creates a new instance of PaymentController
func NewPaymentController(
	rg *gin.RouterGroup,
	paymentUsecase usecase.PaymentUsecase,
	aM middleware.AuthMiddleware,
) PaymentController {
	return &paymentController{
		rg:             rg,
		paymentUsecase: paymentUsecase,
		aM:             aM,
	}
}

// Route sets up payment endpoints
func (p *paymentController) Route() {
	paymentRoutes := p.rg.Group("/payments")

	// Endpoint untuk user dan admin
	userRoutes := paymentRoutes.Group("")
	userRoutes.Use(p.aM.RequireToken("USER", "ADMIN"))
	{
		userRoutes.POST("/", p.CreatePayment)
		userRoutes.GET("/:id", p.GetPaymentStatus)
		userRoutes.POST("/:id/cancel", p.RollbackPayment)
	}
}

// CreatePayment creates a new payment
func (p *paymentController) CreatePayment(c *gin.Context) {
	var req model.PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Bad Request",
			Error:   err.Error(),
		})
		return
	}

	// Call usecase to create payment
	transaction, err := p.paymentUsecase.CreatePayment(req)
	if err != nil {
		log.Printf("Error creating payment: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Internal Server Error",
			Error:   err.Error(),
		})
		return
	}

	// Fetch product details for the response
	product, err := p.paymentUsecase.GetProductByID(transaction.ProductID)
	if err != nil {
		log.Printf("Error fetching product details: %v", err)
		// Continue with limited information
	}

	// Get user information
	user, err := p.paymentUsecase.GetUserByID(transaction.UserID)
	if err != nil {
		log.Printf("Error fetching user details: %v", err)
		// Continue with limited information
	}

	// Return enhanced response
	c.JSON(http.StatusOK, dto.Response{
		Message: "Payment created successfully",
		Data: gin.H{
			"transaction": gin.H{
				"id":          transaction.ID,
				"order_id":    transaction.OrderID,
				"payment_url": transaction.PaymentURL,
				"status":      transaction.Status,
				"created_at":  transaction.CreatedAt,
			},
			"customer": gin.H{
				"id":   transaction.UserID,
				"name": user.Name,
			},
			"purchase": gin.H{
				"product_id":   transaction.ProductID,
				"product_name": product.Name,
				"price":        product.Price,
			},
			"total": transaction.Amount,
		},
	})
}

// GetPaymentStatus gets payment status
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

	// Call usecase to get payment status
	transaction, err := p.paymentUsecase.GetPaymentStatus(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Message: "Not Found",
			Error:   "Transaction not found",
		})
		return
	}

	// Check if user is only accessing their own transaction
	// Get user ID from token
	userID, userIDExists := c.Get("user_id")
	role, roleExists := c.Get("role")
	if !userIDExists || !roleExists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Message: "Unauthorized",
			Error:   "Invalid token",
		})
		return
	}

	// Convert user ID to string
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Internal Server Error",
			Error:   "Failed to parse user ID from token",
		})
		return
	}

	roleStr, ok := role.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Internal Server Error",
			Error:   "Failed to parse role from token",
		})
		return
	}

	// Convert user ID string to int for comparison
	userIDInt, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Internal Server Error",
			Error:   "Failed to convert user ID to integer",
		})
		return
	}

	// Check if user ID from token matches transaction user ID
	if userIDInt != transaction.UserID && roleStr != "ADMIN" {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Message: "Forbidden",
			Error:   "You can only access your own transactions",
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

// RollbackPayment membatalkan pembayaran
func (p *paymentController) RollbackPayment(c *gin.Context) {
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

	// Ambil transaksi terlebih dahulu untuk pengecekan kepemilikan
	transaction, err := p.paymentUsecase.GetPaymentStatus(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Message: "Not Found",
			Error:   "Transaction not found",
		})
		return
	}

	// Panggil usecase untuk rollback pembayaran
	transaction, err = p.paymentUsecase.RollbackPayment(id)
	if err != nil {
		// Cek apakah error karena status transaksi tidak pending
		if strings.Contains(err.Error(), "cannot rollback non-pending") {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: "Bad Request",
				Error:   err.Error(),
			})
			return
		}

		// Error lainnya
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Internal Server Error",
			Error:   "Failed to cancel payment: " + err.Error(),
		})
		return
	}

	// Return response
	c.JSON(http.StatusOK, dto.Response{
		Message: "Payment cancelled successfully",
		Data: gin.H{
			"transaction_id": transaction.ID,
			"order_id":       transaction.OrderID,
			"status":         transaction.Status,
			"updated_at":     transaction.UpdatedAt,
		},
	})
}

// ForceCheckStatus memaksa pengecekan status pembayaran dari Midtrans
func (p *paymentController) ForceCheckStatus(c *gin.Context) {
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

	// Panggil usecase untuk memaksa pengecekan status
	transaction, err := p.paymentUsecase.ForceCheckAndUpdateStatus(id)
	if err != nil {
		// Cek apakah error karena transaksi tidak ditemukan
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Message: "Not Found",
				Error:   "Transaction not found",
			})
			return
		}

		// Error lainnya
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Internal Server Error",
			Error:   "Failed to check payment status: " + err.Error(),
		})
		return
	}

	// Return response
	c.JSON(http.StatusOK, dto.Response{
		Message: "Transaction status updated successfully",
		Data: gin.H{
			"transaction_id": transaction.ID,
			"order_id":       transaction.OrderID,
			"status":         transaction.Status,
			"payment_url":    transaction.PaymentURL,
			"created_at":     transaction.CreatedAt,
			"updated_at":     transaction.UpdatedAt,
		},
	})
}
