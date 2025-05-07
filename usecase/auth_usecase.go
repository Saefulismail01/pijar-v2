package usecase

import (
	"database/sql"
	"errors"
	"log"
	"pijar/model"
	"pijar/repository"
	"pijar/utils"
	"pijar/utils/service"
)

type AuthUsecase struct {
	userRepo   repository.UserRepoInterface
	jwtService service.JwtService
}

func NewAuthUsecase(userRepo repository.UserRepoInterface, jwtService service.JwtService) *AuthUsecase {
	return &AuthUsecase{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

func (u *AuthUsecase) Login(email, password string) (model.AuthResponse, error) {
	user, err := u.userRepo.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AuthResponse{}, errors.New("email not found")
		}
		return model.AuthResponse{}, errors.New("failed to retrieve user")
	}

	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		log.Println("Password mismatch for email:", email)
		return model.AuthResponse{}, errors.New("incorrect password")
	}

	token, err := u.jwtService.CreateToken(user)
	if err != nil {
		return model.AuthResponse{}, err
	}

	return model.AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (u *AuthUsecase) Register(user model.Users, plainPassword string) (map[string]interface{}, error) {
	// Cek apakah email sudah terdaftar
	exists, err := u.userRepo.IsEmailExists(user.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(plainPassword)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = hashedPassword
	user.Role = "USER" // role default user biasa

	// Buat user
	createdUser, err := u.userRepo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	// Buat token langsung setelah register
	token, err := u.jwtService.CreateToken(createdUser)
	if err != nil {
		return nil, err
	}

	// Response
	return map[string]interface{}{
		"message": "Registration successful",
		"user":    createdUser,
		"token":   token,
	}, nil
}
