package usecase

import (
	"errors"
	"fmt"
	"pijar/model"
	"pijar/repository"
	"pijar/utils/service"
)

type UserUsecase interface {
	CreateUserUsecase(user model.Users) (model.Users, error)
	GetAllUsersUsecase() ([]model.Users, error)
	GetUserByIDUsecase(id int) (model.Users, error)
	GetUserByEmail(email string) (model.Users, error)
	UpdateUserUsecase(id int, user model.Users) (model.Users, error)
	DeleteUserUsecase(id int) error
}

type userUsecase struct {
	UserRepo repository.UserRepoInterface
}

func NewUserUsecase(repo repository.UserRepoInterface) *userUsecase {
	return &userUsecase{
		UserRepo: repo,
	}
}

func (u *userUsecase) CreateUserUsecase(user model.Users) (model.Users, error) {
	// Validate user input
	if !service.IsValidEmail(user.Email) {
		return model.Users{}, errors.New("invalid email format")
	}

	if len(user.PasswordHash) < 8 {
		return model.Users{}, errors.New("password must be at least 8 characters")
	}

	exists, err := u.UserRepo.IsEmailExists(user.Email)
	if err != nil {
		return model.Users{}, err
	}
	if exists {
		return model.Users{}, errors.New("email already in use")
	}

	createdUser, err := u.UserRepo.CreateUser(user)
	if err != nil {
		return model.Users{}, err
	}

	return createdUser, nil
}

func (u *userUsecase) GetAllUsersUsecase() ([]model.Users, error) {
	return u.UserRepo.GetAllUsers()
}

func (u *userUsecase) GetUserByIDUsecase(id int) (model.Users, error) {
	return u.UserRepo.GetUserByID(id)
}

func (u *userUsecase) GetUserByEmail(email string) (model.Users, error) {
	return u.UserRepo.GetUserByEmail(email)
}

// UpdateUser updates a user's information
func (u *userUsecase) UpdateUserUsecase(id int, user model.Users) (model.Users, error) {
	// Check if the user exists
	existingUser, err := u.UserRepo.GetUserByID(id)
	if err != nil {
		return model.Users{}, fmt.Errorf("failed to retrieve user: %v", err)
	}

	// If the user doesn't exist, return an error
	if existingUser.ID == 0 {
		return model.Users{}, fmt.Errorf("user not found")
	}

	// Proceed to update the user
	updatedUser, err := u.UserRepo.UpdateUser(user)
	if err != nil {
		return model.Users{}, fmt.Errorf("failed to update user: %v", err)
	}

	return updatedUser, nil
}

func (u *userUsecase) DeleteUserUsecase(id int) error {
	_, err := u.UserRepo.GetUserByID(id)
	if err != nil {
		return fmt.Errorf("cannot delete: %v", err)
	}
	if err := u.UserRepo.DeleteUser(id); err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}

	return nil
}
