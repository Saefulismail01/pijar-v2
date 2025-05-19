package repository

import (
	"database/sql"
	"errors"
	"pijar/model"
)

// OTPRepository interface untuk repository OTP
type OTPRepository interface {
	SaveOTP(otp *model.OTP) error
	GetOTPByCode(code string) (*model.OTP, error)
	DeleteOTP(code string) error
	DeleteExpiredOTPs() error
}

// Implementasi repository OTP
func (r *UserRepo) SaveOTP(otp *model.OTP) error {
	query := `
		INSERT INTO otps (email, code, expires_at, attempts, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	
	_, err := r.DB.Exec(query,
		otp.Email,
		otp.Code,
		otp.ExpiresAt,
		otp.Attempts,
	)
	return err
}

func (r *UserRepo) GetOTPByCode(code string) (*model.OTP, error) {
	query := `
		SELECT id, email, code, expires_at, attempts, created_at, updated_at
		FROM otps
		WHERE code = $1
	`
	
	var otp model.OTP
	err := r.DB.QueryRow(query, code).Scan(
		&otp.ID,
		&otp.Email,
		&otp.Code,
		&otp.ExpiresAt,
		&otp.Attempts,
		&otp.CreatedAt,
		&otp.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("OTP not found")
		}
		return nil, err
	}
	
	return &otp, nil
}

func (r *UserRepo) DeleteOTP(code string) error {
	query := `DELETE FROM otps WHERE code = $1`
	_, err := r.DB.Exec(query, code)
	return err
}

func (r *UserRepo) DeleteExpiredOTPs() error {
	query := `DELETE FROM otps WHERE expires_at < CURRENT_TIMESTAMP`
	_, err := r.DB.Exec(query)
	return err
}
