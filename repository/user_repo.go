package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"pijar/model"
	"time"
)


type UserRepo struct {
    DB *sql.DB
}

func NewUserRepo (db *sql.DB) *UserRepo {
    return &UserRepo{DB: db}
}


type UserRepoInterface interface {
	IsEmailExists(email string) (bool, error)
	CreateUser(user model.Users) (model.Users, error)
	GetAllUsers() ([]model.Users, error)
	GetUserByID(id int) (model.Users, error)
	UpdateUser(user model.Users) (model.Users, error)
	DeleteUser(id int) error
	GetUserByEmail(email string) (model.Users, error)
}

// Ensure *UserRepo implements UserRepoInterface
var _ UserRepoInterface = (*UserRepo)(nil)

func (r *UserRepo) IsEmailExists(email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1);`
	err := r.DB.QueryRow(query, email).Scan(&exists)
	return exists, err
}

func (r *UserRepo) CreateUser(user model.Users) (model.Users, error) {
	// ✅ Set role ke "USER" agar tidak bisa dimanipulasi dari luar
	user.Role = "USER"

	// Mulai transaksi
	tx, err := r.DB.Begin()
	if err != nil {
		return model.Users{}, fmt.Errorf("failed to begin transaction: %v", err)
	}

	fmt.Println("Insert user with:", user.Name, user.Email, user.PasswordHash, user.BirthYear, user.Phone, user.Role)

	insertUserSQL := `
		INSERT INTO users (name, email, password_hash, birth_year, phone, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at;
	`

	var createdAt, updatedAt time.Time
	var id int

	err = tx.QueryRow(insertUserSQL,
		user.Name, user.Email, user.PasswordHash, user.BirthYear, user.Phone, user.Role).
		Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		tx.Rollback()
		return model.Users{}, fmt.Errorf("failed to insert user: %v", err)
	}

	// Commit transaksi
	if err := tx.Commit(); err != nil {
		return model.Users{}, fmt.Errorf("failed to commit transaction: %v", err)
	}

	user.ID = id
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt

	fmt.Println("User successfully registered!")
	return user, nil
}



// GetAllUser
func (r *UserRepo) GetAllUsers() ([]model.Users, error) {
	query := `
		SELECT id, name, email, password_hash, birth_year, phone, created_at, updated_at
		FROM users
	`

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %v", err)
	}
	defer rows.Close()

	var users []model.Users

	for rows.Next() {
		var user model.Users
		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.PasswordHash,
			&user.BirthYear,
			&user.Phone,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %v", err)
		}
		users = append(users, user)
	}

	// Cek error akhir pada iterasi rows
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %v", err)
	}

	return users, nil
}

func (r *UserRepo) GetUserByID(id int) (model.Users, error) {
	query := `
		SELECT id, name, email, password_hash, birth_year, phone, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user model.Users
	err := r.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.BirthYear,
		&user.Phone,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return model.Users{}, errors.New("user not found")
		}
		return model.Users{}, fmt.Errorf("failed to retrieve user: %v", err)
	}

	return user, nil
}



// UpdateUser updates the user details in the database
func (r *UserRepo) UpdateUser(user model.Users) (model.Users, error) {
	// Start a transaction
	tx, err := r.DB.Begin()
	if err != nil {
		return model.Users{}, fmt.Errorf("failed to begin transaction: %v", err)
	}

	// SQL query to update the user
	updateUserSQL := `
		UPDATE users 
		SET name = $1, email = $2, password_hash = $3, birth_year = $4, phone = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
		RETURNING id, name, email, password_hash, birth_year, phone, created_at, updated_at;
	`

	var updatedUser model.Users
	err = tx.QueryRow(updateUserSQL, user.Name, user.Email, user.PasswordHash, user.BirthYear, user.Phone, user.ID).
		Scan(&updatedUser.ID, &updatedUser.Name, &updatedUser.Email, &updatedUser.PasswordHash, &updatedUser.BirthYear, &updatedUser.Phone, &updatedUser.CreatedAt, &updatedUser.UpdatedAt)

	if err != nil {
		tx.Rollback()
		return model.Users{}, fmt.Errorf("failed to update user: %v", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return model.Users{}, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return updatedUser, nil
}



// DeleteUser deletes a user from the database by ID
func (r *UserRepo) DeleteUser(id int) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	deleteSQL := `DELETE FROM users WHERE id = $1`

	res, err := tx.Exec(deleteSQL, id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user: %v", err)
	}

    //validasi Delete success
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to retrieve affected rows: %v", err)
	}

	if rowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("user not found")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (r *UserRepo) GetUserByEmail(email string) (model.Users, error) {
	query := `
		SELECT id, name, email, password_hash, birth_year, phone, role, created_at, updated_at
		FROM users WHERE LOWER(email) = LOWER($1)
	`
	var user model.Users
	err := r.DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.BirthYear,
		&user.Phone,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	// if err != nil {
    //     if errors.Is(err, sql.ErrNoRows) {
    //         return model.Users{}, nil 
    //     }
    //     return model.Users{}, err 
    // }
    // return user, nil
	if err != nil {
		log.Println("ERROR GetUserByEmail:", err)
		return model.Users{}, err
	}

	log.Println("User ditemukan:", user.Email)
	return user, nil
}




