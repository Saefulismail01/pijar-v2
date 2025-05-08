package usecase

import (
	"errors"
	"fmt"
	"pijar/model"
	"pijar/repository"
	"pijar/utils/service"
)


type UserUsecase struct {
	UserRepo repository.UserRepoInterface
}

func NewUserUsecase(repo repository.UserRepoInterface) *UserUsecase {
	return &UserUsecase{
		UserRepo: repo,
	}
}

func (u *UserUsecase) CreateUserUsecase(user model.Users) (model.Users, error) {
	// Validasi email format
	if !service.IsValidEmail(user.Email) {
		return model.Users{}, errors.New("invalid email format")
	}

	// Validasi password minimal
	if len(user.PasswordHash) < 8 {
		return model.Users{}, errors.New("password must be at least 8 characters")
	}

	// Cek apakah email sudah dipakai
	exists, err := u.UserRepo.IsEmailExists(user.Email)
	if err != nil {
		return model.Users{}, err
	}
	if exists {
		return model.Users{}, errors.New("email already in use, please use another one")
	}

	// Simpan user
	createdUser, err := u.UserRepo.CreateUser(user)
	if err != nil {
		return model.Users{}, err
	}

	return createdUser, nil
}


func (u *UserUsecase) GetAllUsersUsecase() ([]model.Users, error) {
	return u.UserRepo.GetAllUsers()
}

func (u *UserUsecase) GetUserByIDUsecase(id int) (model.Users, error) {
	return u.UserRepo.GetUserByID(id)
}


// UpdateUser updates a user's information
func (u *UserUsecase) UpdateUserUsecase(user model.Users) (model.Users, error) {
	// Check if the user exists
	existingUser, err := u.UserRepo.GetUserByID(user.ID)
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

func (u *UserUsecase) DeleteUserUsecase(id int) error {
	// Pastikan user ada terlebih dahulu
	_, err := u.UserRepo.GetUserByID(id)
	if err != nil {
		return fmt.Errorf("cannot delete: %v", err)
	}

	// Lanjutkan penghapusan
	if err := u.UserRepo.DeleteUser(id); err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}

	return nil
}