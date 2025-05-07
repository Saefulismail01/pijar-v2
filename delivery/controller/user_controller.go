package controller

import (
	"konsep_project/middleware"
	"konsep_project/model"
	"konsep_project/repository"
	"konsep_project/usecase"
	"konsep_project/utils/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)



type UserController struct {
	UserUsecase *usecase.UserUsecase
	rg             *gin.RouterGroup
	userRepo       repository.UserRepoInterface
	jwtService     service.JwtService
	authMiddleware *middleware.AuthMiddleware
	
}


func NewUserController(rg *gin.RouterGroup, userUsecase *usecase.UserUsecase, userRepo repository.UserRepoInterface, jwtService service.JwtService, authMiddleware *middleware.AuthMiddleware) *UserController{
	return &UserController{
		rg:             rg,
		userRepo:       userRepo,
		jwtService:     jwtService,
		authMiddleware: authMiddleware,
		UserUsecase:    userUsecase,
	}
}

// func (uc *UserController) Route() {
// 	// uc.rg.POST("/register", uc.CreateUserController)
// 	// uc.rg.POST("/login", uc.Login)

// 	// hanya admin
// 	uc.rg.GET("/users", uc.authMiddleware.RequireToken("admin"), uc.GetAllUsersController)
// 	uc.rg.GET("/users/:id", uc.authMiddleware.RequireToken("admin"), uc.GetUserByIDController)
// 	uc.rg.GET("/users/email/:email", uc.authMiddleware.RequireToken("admin"), uc.GetUserByEmail)
// 	uc.rg.PUT("/users", uc.authMiddleware.RequireToken("admin"), uc.UpdateUserController)
// 	uc.rg.DELETE("/users/:id", uc.authMiddleware.RequireToken("admin"), uc.DeleteUserController)
// }

func (uc *UserController) CreateUserController(c *gin.Context) {
	var userInput model.Users

	// Bind JSON dari request body ke struct userInput
	if err := c.ShouldBindJSON(&userInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// panggil Usecase to create user
	createdUser, err := uc.UserUsecase.CreateUserUsecase(userInput)
	if err != nil {
		// Kembalikan error yang terjadi
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return data user yang berhasil
	c.JSON(http.StatusOK, gin.H{
		"user": createdUser,
	})
}



func (uc *UserController) GetAllUsersController(c *gin.Context) {
	users, err := uc.UserUsecase.GetAllUsersUsecase()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch users",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": users,
	})
}


func (uc *UserController) GetUserByIDController(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := uc.UserUsecase.GetUserByIDUsecase(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (uc *UserController) UpdateUserController(c *gin.Context) {
	// Get user ID from URL parameter
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse the request body
	var user model.Users
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Set the user ID from the URL parameter
	user.ID = userID

	// Call the usecase to update the user
	updatedUser, err := uc.UserUsecase.UpdateUserUsecase(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the updated user data
	c.JSON(http.StatusOK, updatedUser)
}

func (uc *UserController) DeleteUserController(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = uc.UserUsecase.DeleteUserUsecase(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (uc *UserController) GetUserByEmail(c *gin.Context) {
	email := c.Param("email")

	user, err := uc.userRepo.GetUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (uc *UserController) Login(c *gin.Context) {
	var loginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user, err := uc.userRepo.GetUserByEmail(loginReq.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email not found"})
		return
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginReq.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate token
	token, err := uc.jwtService.CreateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
