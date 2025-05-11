package controller

import (
	"log"
	"net/http"
	"pijar/middleware"
	"pijar/model"
	"pijar/model/dto"
	"pijar/usecase"
	"pijar/utils/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	UserUsecase    usecase.UserUsecase
	rg             *gin.RouterGroup
	jwtService     service.JwtService
	authMiddleware *middleware.AuthMiddleware
}

func NewUserController(rg *gin.RouterGroup, userUsecase usecase.UserUsecase, jwtService service.JwtService, authMiddleware *middleware.AuthMiddleware) *UserController {
	return &UserController{
		UserUsecase:    userUsecase,
		rg:             rg,
		jwtService:     jwtService,
		authMiddleware: authMiddleware,
	}
}

func (uc *UserController) Route() {
	// Admin-only protected routes with JWT authentication
	adminProtected := uc.rg.Group("/users")
	adminProtected.Use(uc.authMiddleware.RequireToken("ADMIN"))
	adminProtected.GET("/", uc.GetAllUsersController)
	adminProtected.GET("/:id", uc.GetUserByIDController)
	adminProtected.PUT("/:id", uc.UpdateUserController)
	adminProtected.DELETE("/:id", uc.DeleteUserController)
	adminProtected.GET("/email/:email", uc.GetUserByEmail)

	// Endpoint for creating new user (admin only)
	log.Println("Registering POST /users endpoint for creating new users")
	adminProtected.POST("/", uc.CreateUserController)
	log.Println("POST /users endpoint registered")

	// User profile routes - accessible by any authenticated user
	userProfile := uc.rg.Group("/profile")
	userProfile.Use(uc.authMiddleware.RequireToken("USER", "ADMIN")) // Both users and admins can access
	userProfile.GET("/", uc.GetOwnProfileController)
	userProfile.PUT("/:id", uc.UpdateOwnProfileController)
}

func (uc *UserController) CreateUserController(c *gin.Context) {
	var userInput model.Users

	// Bind JSON from request body to userInput struct
	if err := c.ShouldBindJSON(&userInput); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Bad Request",
			Error:   "Invalid input",
		})
		return
	}

	// Call Usecase to create user
	createdUser, err := uc.UserUsecase.CreateUserUsecase(userInput)
	if err != nil {
		// Return the error that occurred
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Bad Request",
			Error:   err.Error(),
		})
		return
	}

	// Return the successfully created user data
	c.JSON(http.StatusOK, dto.Response{
		Message: "User created successfully",
		Data:    createdUser,
	})
}

func (uc *UserController) GetAllUsersController(c *gin.Context) {
	users, err := uc.UserUsecase.GetAllUsersUsecase()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Internal Server Error",
			Error:   "Failed to fetch users",
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Message: "Users retrieved successfully",
		Data:    users,
	})
}

func (uc *UserController) GetUserByIDController(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Bad Request",
			Error:   "Invalid user ID",
		})
		return
	}

	user, err := uc.UserUsecase.GetUserByIDUsecase(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Message: "Not Found",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Message: "User retrieved successfully",
		Data:    user,
	})
}

func (uc *UserController) UpdateUserController(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Bad Request",
			Error:   "Invalid user ID",
		})
		return
	}

	var userInput model.Users
	if err := c.ShouldBindJSON(&userInput); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Bad Request",
			Error:   "Invalid input",
		})
		return
	}

	updatedUser, err := uc.UserUsecase.UpdateUserUsecase(id, userInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Internal Server Error",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Message: "User updated successfully",
		Data:    updatedUser,
	})
}

func (uc *UserController) DeleteUserController(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Bad Request",
			Error:   "Invalid user ID",
		})
		return
	}

	err = uc.UserUsecase.DeleteUserUsecase(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Internal Server Error",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Message: "User deleted successfully",
	})
}

func (uc *UserController) GetUserByEmail(c *gin.Context) {
	email := c.Param("email")
	user, err := uc.UserUsecase.GetUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Message: "Not Found",
			Error:   "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Message: "User retrieved successfully",
		Data:    user,
	})
}

func (uc *UserController) GetOwnProfileController(c *gin.Context) {
	// Get user ID from JWT token
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Message: "Unauthorized",
			Error:   "No token provided",
		})
		return
	}

	// Get user ID from the JWT token
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Message: "Unauthorized",
			Error:   "User ID not found in token",
		})
		return
	}

	userID, err := strconv.Atoi(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Invalid user ID",
			Error:   "Invalid user ID format",
		})
		return
	}

	// Get user profile
	user, err := uc.UserUsecase.GetUserByIDUsecase(int(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Message: "User not found",
			Error:   "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// UpdateOwnProfileController allows users to update their own profile
func (uc *UserController) UpdateOwnProfileController(c *gin.Context) {
	// Get user ID from the JWT token
	userIDStr := c.Param("id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Message: "Unauthorized",
			Error:   "User ID not found in token",
		})
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Invalid user ID format",
			Error:   "invalid user ID format",
		})
		return
	}

	// Get existing user to preserve password if not updating it
	existingUser, err := uc.UserUsecase.GetUserByIDUsecase(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Message: "User not found",
			Error:   "user not found",
		})
		return
	}

	// Parse the request body
	var updateRequest struct {
		Name      string `json:"name"`
		Email     string `json:"email"`
		Password  string `json:"password,omitempty"`
		BirthYear int    `json:"birth_year"`
		Phone     string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Invalid request body",
			Error:   "invalid request body",
		})
		return
	}

	// Prepare user object for update
	user := existingUser
	user.Name = updateRequest.Name
	user.Email = updateRequest.Email
	user.BirthYear = updateRequest.BirthYear
	user.Phone = updateRequest.Phone
	// Keep the original role - users can't change their own role

	// Only update password if provided
	if updateRequest.Password != "" {
		hashedPassword, err := service.HashPassword(updateRequest.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Message: "Failed to hash password",
				Error:   "failed to hash password",
			})
			return
		}
		user.PasswordHash = hashedPassword
	}

	// Call the usecase to update the user
	updatedUser, err := uc.UserUsecase.UpdateUserUsecase(userID, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Failed to update user",
			Error:   "failed to update user",
		})
		return
	}

	// Return the updated user data
	c.JSON(http.StatusOK, dto.Response{
		Message: "Profile updated successfully",
		Data:    updatedUser,
	})
}

// Login method has been moved to AuthController
