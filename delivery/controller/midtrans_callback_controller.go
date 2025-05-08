package controller

import (
	"konsep_project/model"
	"konsep_project/usecase"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// MidtransCallbackHandler adalah interface untuk handler callback Midtrans
type MidtransCallbackHandler interface {
	HandleCallback(c *gin.Context)
	Route()
}

// midtransCallbackHandler adalah implementasi dari MidtransCallbackHandler
type midtransCallbackHandler struct {
	paymentUsecase usecase.PaymentUsecase
	rg             *gin.RouterGroup
}

// NewMidtransCallbackHandler membuat instance baru dari MidtransCallbackHandler
func NewMidtransCallbackHandler(
	rg *gin.RouterGroup,
	paymentUsecase usecase.PaymentUsecase,
) MidtransCallbackHandler {
	return &midtransCallbackHandler{
		rg:             rg,
		paymentUsecase: paymentUsecase,
	}
}

// Route mengatur endpoint untuk callback Midtrans
func (h *midtransCallbackHandler) Route() {
	h.rg.POST("/midtrans/callback", h.HandleCallback)
}

// HandleCallback menangani callback dari Midtrans
func (h *midtransCallbackHandler) HandleCallback(c *gin.Context) {
	var callback model.MidtransCallbackRequest
	if err := c.ShouldBindJSON(&callback); err != nil {
		log.Printf("Error binding callback JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log callback untuk debugging
	log.Printf("Received Midtrans callback: %+v", callback)

	// Proses callback menggunakan usecase
	err := h.paymentUsecase.ProcessCallback(callback)
	if err != nil {
		log.Printf("Error processing callback: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process callback"})
		return
	}

	// Berikan response OK ke Midtrans
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}
