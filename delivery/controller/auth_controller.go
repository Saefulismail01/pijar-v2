package controller

import (
	"net/http"
	"pijar/model"
	"pijar/usecase"
	"pijar/utils"
	"pijar/utils/service"

	"github.com/gin-gonic/gin"
)


type AuthController struct {
	rg          *gin.RouterGroup
	jwtService  service.JwtService
	AuthUsecase *usecase.AuthUsecase
}


func NewAuthController(rg *gin.RouterGroup, jwtService service.JwtService, authUsecase *usecase.AuthUsecase) *AuthController {
	return &AuthController{
		rg:          rg, 
		jwtService:  jwtService,
		AuthUsecase: authUsecase,
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengenkripsi password"})
		return
	}

	user := model.Users{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: hashedPassword,
		BirthYear:    input.BirthYear,
		Phone:        input.Phone,
		Role: 		  "user",
	}

	authResp, err := a.AuthUsecase.Register(user, input.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}


	authResp, err := a.AuthUsecase.Login(input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, authResp)
}
