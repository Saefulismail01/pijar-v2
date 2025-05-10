package controller

import (
	"log"
	"net/http"
	"pijar/model"
	"pijar/model/dto"
	"pijar/usecase"

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
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Invalid callback request",
			Error:   err.Error(),
		})
		return
	}

	// Log callback untuk debugging
	log.Printf("Received Midtrans callback: %+v", callback)

	// Proses callback menggunakan usecase
	err := h.paymentUsecase.ProcessCallback(callback)
	if err != nil {
		log.Printf("Error processing callback: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Failed to process callback",
			Error:   err.Error(),
		})
		return
	}

	// Berikan response OK ke Midtrans
	c.JSON(http.StatusOK, dto.Response{
		Message: "OK",
		Data:    "OK",
	})
}
