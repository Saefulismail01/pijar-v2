package usecase

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"pijar/model"
	"pijar/utils/service"
	"time"
	"golang.org/x/crypto/bcrypt"
)

func (u *userUsecase) GenerateOTP(email string) (string, error) {
	// Validate email
	if !service.IsValidEmail(email) {
		return "", errors.New("invalid email format")
	}

	// Check if email exists
	exists, err := u.UserRepo.IsEmailExists(email)
	if err != nil {
		return "", err
	}
	if exists {
		return "", errors.New("email already registered")
	}

	// Generate 6 digit OTP
	rand.Seed(time.Now().UnixNano())
	optCode := fmt.Sprintf("%06d", rand.Intn(999999))
	
	// Create OTP record
	opt := model.OTP{
		Email:     email,
		Code:      optCode,
		ExpiresAt: time.Now().Add(10 * time.Minute),
		Attempts:  0,
	}
	
	// Save OTP to database
	if err := u.UserRepo.SaveOTP(&opt); err != nil {
		return "", err
	}
	
	// Send OTP via email (implement email service)
	// For now, we'll just log it
	log.Printf("OTP sent to %s: %s", email, optCode)
	
	return optCode, nil
}

func (u *userUsecase) VerifyOTP(email string, otp string) (model.Users, error) {
	// Get OTP from database
	storedOTP, err := u.UserRepo.GetOTPByCode(otp)
	if err != nil {
		return model.Users{}, err
	}
	
	// Check if OTP expired
	if time.Now().After(storedOTP.ExpiresAt) {
		return model.Users{}, errors.New("OTP has expired")
	}
	
	// Check if too many attempts
	if storedOTP.Attempts >= 3 {
		return model.Users{}, errors.New("too many attempts")
	}
	
	// Create user
	user := model.Users{
		Email:        email,
		PasswordHash: storedOTP.Code, // Changed to use PasswordHash
	}
	
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return model.Users{}, err
	}
	user.PasswordHash = string(hashedPassword)
	
	// Save user
	createdUser, err := u.UserRepo.CreateUser(user)
	if err != nil {
		return model.Users{}, err
	}
	
	// Delete used OTP
	if err := u.UserRepo.DeleteOTP(otp); err != nil {
		log.Printf("Failed to delete OTP: %v", err)
	}
	
	return createdUser, nil
}
