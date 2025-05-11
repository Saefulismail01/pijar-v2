package controller

import (
	"net/http"
	"pijar/model"
	"pijar/model/dto"
	"pijar/usecase"
	"pijar/utils/service"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	rg          *gin.RouterGroup
	jwtService  service.JwtService
	AuthUsecase usecase.AuthUsecase
}

func NewAuthController(rg *gin.RouterGroup, jwt service.JwtService, authUC usecase.AuthUsecase) *AuthController {
	return &AuthController{
		rg:          rg,
		jwtService:  jwt,
		AuthUsecase: authUC,
	}
}

func (ac *AuthController) Route() {
	ac.rg.POST("/register", ac.Register)
	ac.rg.POST("/login", ac.Login)
}

func (a *AuthController) Register(c *gin.Context) {
	var input struct {
		Name      string `json:"name"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		BirthYear int    `json:"birth_year"`
		Phone     string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Bad Request",
			Error:   "invalid input",
		})
		return
	}

	// Hash password
	hashedPassword, err := service.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Failed to encrypt password",
			Error:   "gagal mengenkripsi password",
		})
		return
	}

	user := model.Users{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: hashedPassword,
		BirthYear:    input.BirthYear,
		Phone:        input.Phone,
		Role:         "user",
	}

	authResp, err := a.AuthUsecase.Register(user, input.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Registration failed",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, authResp)
}

func (a *AuthController) Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	authResp, err := a.AuthUsecase.Login(input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Message: "Invalid credentials",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, authResp)
}
